# William

This library aims to prevent writing too much code when using william on your application.

## Creating a service

```go
import william github.com/topfreegames/Will.IAM/client

// [...]

will := william.New("URL OF WILLIAM", "ServiceName")
```
You can set the HTTP Client that will made the request.

```go
will.SetClient(
    &http.Client{
        Timeout: time.Second * 2,
    },
)
```

If your service will use any internal function of william like a service, eg: adding permission directly from your service you need to set a KeyPair to auth.

```go
will.SetKeyPair(
    "ServiceKeyID",
    "ServiceKeySecret",
)
```

By default, Will.IAM is enabled, You can disable it using will.Bypass()

```go
will.Bypass()
```

## Permissions Handler
You want to validate if a user have a valid permission for a specific Handler. If your application don't care about resource you can use william.Generate("RL", "Action") that will wrap this for you.

```go
mux := http.NewServeMux()
mux.HandleFunc("/action",
    will.HandlerFunc(will.Generate("RL", "Action"),
    func(w http.ResponseWriter, r *http.Request) {
        // On Request you will have access to william.
        auth := william.Auth(r.Context())
        fmt.Fprintf(w, "Your mail is: %s!", auth.Email())
    }),
)
```
It's recommended always to register you actions using the method by your william instance, so you can have a generated list for /am.

If you want to create a custom permission check, provide a function like this:

```go
permission := func(r *http.Request) string {
    // You can get access to your William Instance.
    if will := william.Get(r.Context()); will != nil {
        return fmt.Sprintf(
            "%s::%s::%s::%s",
            will.GetServiceName(),
            "RL",
            "Action",
            "*",
        )
    }

    return ""
}

mux.HandleFunc("/action",
    will.HandlerFunc(permission,
    func(w http.ResponseWriter, r *http.Request) {
        // On Request you will have access to william.
        auth := william.Auth(r.Context())
        fmt.Fprintf(w, "Your mail is: %s!", auth.Email())
    }),
)
```

### Another example if you use Gorrila Mux:

If you have a route that looks like:
```
/action/{gameid}
```
and you want to check the permission for every gameid, you can use

```go
permission := func(r *http.Request) string {
    if will := william.Get(r.Context()); will != nil {
        vars := mux.Vars(r)

        return fmt.Sprintf(
            "%s::%s::%s::%s",
            will.GetServiceName(),
            "RL",
            "Action",
            vars["gameid"],
        )
    }

    return ""
}

mux.HandleFunc("/action/{gameid}",
    will.HandlerFunc(permission,
    func(w http.ResponseWriter, r *http.Request) {
        // On Request you will have access to william.
        auth := william.Auth(r.Context())
        fmt.Fprintf(w, "Your mail is: %s!", auth.Email())
    }),
)
```

## /am
The most challenging part, but with some helper, you can easily archive this.
To get a list of actions and resource that your service has, you need to provide an endpoint.

If you use generate on your william interface, all that action will be register and exported automatically.
If not, you can use:

```go
will.AddAction("Action")
```
that will register an action like: ::Action::* 
but if you have a resource, you will need to use AddActionFunc.
An example of how to use:

```go
will := william.New(...)

will.AddActionFunc("EditCategory", func(ctx context.Context, prefix string) []william.AmPermission {
    categories, err := s.repo.ListCategories(ctx, "")
    if err != nil {
        return nil
    }

    list := make([]william.AmPermission, len(categories))
    for i, category := range categories {
        list[i] = william.AmPermission{
            Complete: true,
            Alias:    category.Name,
            Prefix:   prefix + strconv.Itoa(int(category.ID)),
        }
    }
    return list
})
```

Now you only need to expose a /am

```go
func (s *Server) runInternal() {
    will := william.New(...)

    // [...]

	mux := http.NewServeMux()
	mux.HandleFunc("/am", will.AmHandler)

	_ = http.ListenAndServe(":9090", mux)
}
```
