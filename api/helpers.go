package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"

	"github.com/topfreegames/Will.IAM/constants"
	"github.com/topfreegames/Will.IAM/errors"
	"github.com/topfreegames/Will.IAM/repositories"
)

// ListResponse is used when returning an array in results, normally used alongside
// repositories.ListOptions
type ListResponse struct {
	Count   int64       `json:"count"`
	Results interface{} `json:"results"`
}

// ErrorResponse is used when responding with an error
type ErrorResponse struct {
	Error string `json:"error"`
}

func keepJSONFieldsSl(
	isl interface{}, keep ...string,
) ([]map[string]interface{}, error) {
	bts, err := json.Marshal(isl)
	if err != nil {
		return nil, err
	}
	var mSl []map[string]interface{}
	if err := json.Unmarshal(bts, &mSl); err != nil {
		return nil, err
	}
	kmSl := make([]map[string]interface{}, len(mSl))
	for i := range mSl {
		kmSl[i] = map[string]interface{}{}
		for _, f := range keep {
			kmSl[i][f] = mSl[i][f]
		}
	}
	return kmSl, nil
}

func keepJSONFieldsOne(
	i interface{}, keep ...string,
) (map[string]interface{}, error) {
	bts, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(bts, &m); err != nil {
		return nil, err
	}
	km := map[string]interface{}{}
	for _, f := range keep {
		km[f] = m[f]
	}
	return km, nil
}

func keepJSONFields(i interface{}, keep ...string) (interface{}, error) {
	switch reflect.TypeOf(i).Kind() {
	case reflect.Slice:
		return keepJSONFieldsSl(i, keep...)
	default:
		return keepJSONFieldsOne(i, keep...)
	}
}

func keepJSONFieldsBytes(i interface{}, keep ...string) ([]byte, error) {
	ri, err := keepJSONFields(i, keep...)
	if err != nil {
		return nil, err
	}
	return json.Marshal(ri)
}

func buildListOptions(r *http.Request) (*repositories.ListOptions, error) {
	str := r.URL.Query().Get("page")
	if str == "" {
		str = "0"
	}
	page, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return nil, errors.NewInvalidPageError(str)
	}
	pageSize := constants.DefaultListOptionsPageSize
	str = r.URL.Query().Get("pageSize")
	if str != "" {
		pageSize64, err := strconv.ParseInt(str, 10, 32)
		if err != nil {
			return nil, errors.NewInvalidPageSizeError(str)
		}
		pageSize = int(pageSize64)
	}
	return &repositories.ListOptions{
		Page:     int(page),
		PageSize: pageSize,
	}, nil
}

// unmarshalBodyTo unmarshal content from r.Body to i and calls r.Body.Close()
func unmarshalBodyTo(r *http.Request, i interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, i)
}
