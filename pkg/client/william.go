package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// William Interface
type William interface {
	Enable()
	SetClient(client HttpClient)
	SetKey(id, secret string)
	GetServiceName() string
	ListPermission(ctx context.Context, ownershipLevel, action string, resourceHierarchy ...string) ([]byte, error)
	HandlerFunc(permission func(r *http.Request) string, next http.HandlerFunc) http.HandlerFunc
	AmHandler(w http.ResponseWriter, r *http.Request)
	AddAction(action string)
	AddActionFunc(action string, f ActionResourceFunc)
	Generate(ownershipLevel, action string, resourceHierarchy ...string) func(r *http.Request) string
}

type AmPermission struct {
	Alias    string `json:"alias,omitempty"`
	Prefix   string `json:"prefix"`
	Complete bool   `json:"complete"`
}

type ActionResourceFunc func(context.Context, string) []AmPermission

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type william struct {
	sync.RWMutex
	actions map[string]ActionResourceFunc

	keyID     string
	keySecret string

	serviceName string
	baseURL     string
	bypass      bool

	client HttpClient
}

func New(baseURL, serviceName string) William {
	wi := &william{
		baseURL:     baseURL,
		actions:     make(map[string]ActionResourceFunc),
		serviceName: serviceName,
		bypass:      true,
		client:      http.DefaultClient,
	}

	return wi
}

func (wi *william) Enable() {
	wi.bypass = false
}

func (wi *william) SetClient(client HttpClient) {
	wi.Lock()
	defer wi.Unlock()

	wi.client = client
}

func (wi *william) SetKey(id, secret string) {
	wi.keyID = id
	wi.keySecret = secret
}

func (wi *william) GetServiceName() string { return wi.serviceName }

func (wi *william) hasPermission(ctx context.Context, accessToken, permission string) (*AuthInfo, error) {
	if len(accessToken) < 8 {
		return nil, errors.New("invalid auth token")
	}

	auth, err := wi.get(ctx, "/permissions/has?permission="+url.QueryEscape(permission), accessToken)
	if err != nil {
		return nil, err
	}

	auth.permission = permission
	return auth, err
}

func (wi *william) ListPermission(ctx context.Context, ownershipLevel, action string, resourceHierarchy ...string) ([]byte, error) {
	params := url.Values{}
	permission := append([]string{wi.serviceName, ownershipLevel, action}, resourceHierarchy...)
	params.Set("permission", strings.Join(permission, "::"))
	params.Set("pageSize", "0")

	authToken := fmt.Sprintf("KeyPair %s:%s", wi.keyID, wi.keySecret)
	authInfo, err := wi.get(ctx, "/service_accounts/with_permission?"+params.Encode(), authToken)
	if err != nil {
		return nil, err
	}

	return authInfo.body, nil
}

func (wi *william) get(ctx context.Context, path, authtoken string) (*AuthInfo, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		wi.baseURL+path,
		nil,
	)
	if err != nil {
		return nil, err
	}

	if ctx != nil {
		req = req.WithContext(ctx)
	}

	req.Header.Set("Authorization", authtoken)

	resp, err := wi.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return &AuthInfo{
		code:  resp.StatusCode,
		name:  resp.Header.Get("x-service-account-name"),
		token: resp.Header.Get("x-access-token"),
		email: resp.Header.Get("x-email"),
		body:  body,
	}, err
}

func (wi *william) HandlerFunc(permission func(r *http.Request) string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !wi.bypass && permission != nil {
			auth := r.Header.Get("Authorization")
			a, err := wi.hasPermission(r.Context(), auth, permission(setWilliam(r, wi)))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if a.code != http.StatusOK && a.permission != "" {
				w.WriteHeader(a.code)
				return
			}

			parts := strings.Split(auth, " ")
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				if a.token != "" && a.token != parts[1] {
					w.Header().Set("x-access-token", a.token)
				}

				if a.email != "" {
					w.Header().Set("x-email", a.email)
				}
			}

			next(w, setAuth(r, a))
			return
		}

		next(w, r)
	}
}

func (wi *william) AmHandler(w http.ResponseWriter, r *http.Request) {
	list := []AmPermission{}
	wi.RLock()
	defer wi.RUnlock()

	prefix := r.URL.Query().Get("prefix")
	if strings.HasSuffix(prefix, "::") {
		action := strings.Split(prefix, "::")[0]
		if f, ok := wi.actions[action]; ok {
			list = append(list, AmPermission{
				Prefix:   prefix + "*",
				Complete: true,
			})
			if f != nil {
				list = append(list, f(r.Context(), prefix)...)
			}
		}
	} else {
		for action := range wi.actions {
			list = append(list, AmPermission{
				Prefix:   action,
				Complete: false,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

func (wi *william) AddAction(action string) {
	wi.AddActionFunc(action, nil)
}

func (wi *william) AddActionFunc(action string, f ActionResourceFunc) {
	wi.Lock()
	defer wi.Unlock()

	if f != nil {
		wi.actions[action] = f
	} else if _, ok := wi.actions[action]; !ok {
		wi.actions[action] = nil
	}
}

func (wi *william) Generate(ownershipLevel, action string, resourceHierarchy ...string) func(r *http.Request) string {
	wi.AddAction(action)
	return Generate(ownershipLevel, action, resourceHierarchy...)
}
