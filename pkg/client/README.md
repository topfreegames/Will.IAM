# William

This library aims to prevent writing to much code using william on your application.

## Creating a service
```go
wi := william.New("URL OF WILLIAM", "ServiceName")
```
You can set the HTTP Client that will made the request also the tracing if anyone. (if not set nil)
```go
wi.SetClient(
    &http.Client{
        Timeout: time.Second * 2,
    },
    nil,
)
```

If your service will use any internal function of william like a service, Ej: adding permission directly from your service and not from the Hub you need to set a KeyPair to auth.

```go
wi.SetKeyPair(
    "ServiceKeyID",
    "ServiceKeySecret",
)
```

By default William is created enable, you can bypass him with:
```go
wi.ByPass()
```

## Permissions Handler
You want to validate if a user have a valid permission for a specific Handler. If your application don't care about resource you can use william.Generate("RL", "Action") that will wrap this for you.
```go
mux := http.NewServeMux()
mux.HandleFunc("/action",
    wi.HandlerFunc(wi.Generate("RL", "Action"),
    func(w http.ResponseWriter, r *http.Request) {
        // On Request you will have access to william.
        auth := william.Auth(r)
        fmt.Fprintf(w, "Your mail is: %s!", auth.Email())
    }),
)
```
I recommend always to register you actions using the method by your william instance, so you can have a generated list for /am.

If you want to create a custom permission check, provide a function like this: 
```go
permission := func(r *http.Request) string {
    // You can get access to your William Instance.
    if wi := william.Get(r); wi != nil {
        return fmt.Sprintf(
            "%s::%s::%s::%s",
            wi.GetServiceName(),
            "RL",
            "Action",
            "*",
        )
    }

    return ""
}
```

### Another example if you use Gorrila Mux:

You have a route that is something like:
```
/action/{gameid}
```
and you want to check the permission for every gameid, you can use

```go
permission := func(r *http.Request) string {
    if wi := william.Get(r); wi != nil {
        vars := mux.Vars(r)

        return fmt.Sprintf(
            "%s::%s::%s::%s",
            wi.GetServiceName(),
            "RL",
            "Action",
            vars["gameid"],
        )
    }

    return ""
}
```

## /am
The most challenging part, but with some helper, you can easily archive this.
To get a list of actions and resource that your service has, you need to provide an endpoint.

If you use generate on your william interface, all that action will be register and exported automatically.
If not, you can use:
```go
wi.AddAction("Action")
```
that will register an action like: ::Action::* 
but if you have a resource, you will need to use AddActionFunc.
An example of how to use:
```go
s.william.AddActionFunc("EditCategory", func(ctx context.Context, prefix string) []william.AmPermission {
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
I recommend to use an internal port and not expose to the internet. (On the hub you set the internal cluster DNS to the service)
```go
func (s *Server) runInternal() {
	mux := http.NewServeMux()
	mux.HandleFunc("/am", s.william.AmHandler)

	_ = http.ListenAndServe(":9090", mux)
}
```
