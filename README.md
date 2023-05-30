# Rest APIs

Building a restful api.

We will be using the following format 

## Running Locally

To run the application go into the `bin` directory and run the following command

```bash
# this is optional as default is set to "8080"
$ export PORT=[port number]
$ /web
```

While running, the following API called are accessible to the end user.

```http
@hostname = localhost
@port = 8080
@host = {{hostname}}:{{port}}
@contentType = application/json
@mediaType = application/vnd.api+json
@apiVersion = 1.0

GET http://{{host}}
Content-Type: {{contentType}}
Accept: {{mediaType}}; version={{apiVersion}}
```

## Resources

1. https://github.com/Masterminds/semver

1. https://github.com/hashicorp/go-version