# Goesi

A Go library to interact with the EVE Swagger Interface (ESI).

This library uses [Gabs](https://github.com/Jeffail/gabs) to handle the JSON; static references are not used.

## Initialization

Call `New()`, passing in your EVE app information:

```go
esi := goesi.New(
    "clientID",
    "clientSecret",
    "clientCallbackURL",
)
```

## Getting data from ESI

Call `Get()`, passing in the URL path. If you wanted to get all wars, your path is just `"wars"` - don't pass in the ESI root URL.

```go
data, err := esi.Get("wars")
if err != nil {
    // handle error
}
fmt.Println(data)
```

`Get()` supports string formatting, saving you a call to `fmt.Sprintf`:

```go
data, err := esi.Get("wars/%d", 10000)
if err != nil {
    // handle error
}
fmt.Println(data)
```

Responses to GET requests are cached for the duration set by the response from ESI. If you need to override the cache for some reason, there's an `esi.ClearCache()` method.

## Posting data to ESI

Call `Post()`, again passing both the target URL path and the _string_ request body. When passing in JSON, you need to convert it to a string yourself.

```go
data, err := esi.Post("characters/affiliation", fmt.Sprintf("[%d]", 100000))
if err != nil {
    // handle error
}
fmt.Println(data)
```

As `Post()` has to take the body as a parameter, there's no automatic string formatting on this method.
