# Version

## Getting started

To use in project, run the following command in the terminal

```bash
$ go get github.com/adoublef-go/version
```

## Usage

An example of how this project can be used:

```go
mux := chi.NewMux()
mux.Use(v.Version("vnd.api+json"))
mux.Mount("/", v.Match(v.Map{"^1": apiHandlerV1, "^2": apiHandlerV2 }))
```

## Running Locally

You can clone the example package has a small project that can be ran locally. It takes a `PORT` environment variable (defaults to __8080__)

```bash
# this is optional as default is set to "8080"
$ cd example
$ go run .
```

While running, the following API called are accessible to the end user.

```bash
# 1.x.y or 2.x.y will yield a response. Anything else will yield an error
$ curl --location 'http://localhost:8080/' \
--header 'Accept: application/vnd.api+json;version=1'
```

## License

adoublef-go/version is free and open-source software licensed under the [MIT License](https://www.tldrlegal.com/license/mit-license).