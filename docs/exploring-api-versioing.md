# Exploring RESTful APIs & Versioning

## Background

The premise of this article is to share my story, implementing REST API versioning with the header. This does read like a lot of jargon, but not to worry as we are going to break this journey down & hopefully at the end you will have learnt something new.

## Introduction

For new developers, the pursuit to gaining a job in the profession may start with building web services. After all, in this day and age, so much of the world is connected through the web, it makes for a logical choice. However, even this can lead to very broad learning, and knowing what is important to sacrifice your time with, can be hard. 

Through much trail and error, I have found there are a few common topics that show up in tutorial blogs/videos that the developers feel are important to understand. One of these, which the remainder of this article will discuss, is [API versioning](https://www.baeldung.com/rest-versioning).

## What is API?

To build some context, it makes sense to introduce what an API is. We can use the example of a music streaming app (__Application__) that you, the end user downloads on their device. This will communicate with a service (__Programme__) that I, the developer, has running on some server in the world. This is done  through a set of rules and instructions (__Interface__) enabling the user to request data, like their favourite playlist in a safe & predictable manner. 

In short we have explained an API.

## So what is API versioning?

If we suggest that an API is just a set of rules and instructions, then versioning can be seen as a contract. Whereby in agreeing to the contract, both parties can be sure that everyone is using the same set of rules and instructions. When I, the developer, want to introduce a new rule, which doesn't conform the previous contract, a new one is created, to which you, the user are allows to upgrade to. This [article](https://www.baeldung.com/rest-versioning#contract) goes a bit more technical its explanation. This article will mainly focus on the implantations.

> Note from this point, we will be exploring some code.

## Implementing Options

There are a few approaches to  in which we can achieve this so that the user is not blind to the version of API that they are interacting with:

- __URI__ version the URI space using version indicators

```
<!-- URL -->
===>
GET http://foo.app/v1/users/1 HTTP/1.1
<===
HTTP/1.1 200 OK
{
    "name": "John Smith"
}
```

- __Media Type__ version the Representation of the Resource

```
===>
GET http://bar.app/users/3 HTTP/1.1
Accept: application/vnd.api+json;version=2
<===
HTTP/1.1 200 OK
Content-Type: application/vnd.api+json;version=2
{
    "name": "John Smith"
}
```

We can take a look at a comparison between the two options and see they both can return the same data. There is not necessarily a right or wrong to which you decide to use. The former option is generally easier to implement and often what intermediate tutorials will encourage to budding developers. This article though will take a dive into the latter option.

<!-- We can take a look at a comparison between the two options. -->

When I was first introduced to this option, what appealed to me was that the URI was cleaner. 


My language of choice for developing web servers has been [Go](https://go.dev). A future article may go into detail why I enjoy using it so much, but for now we will look to versioning using this language.

## Approaching the problem

When I task myself with a project I first map out the what the high-level interface could look like. In this case I wanted to be able to do something like this:

```go
func main() {
    mux := chi.NewMux()
    // this middleware will grab the version from the request somehow
    mux.Use(version.Version("vnd.api+json")) 
    // this will match the handler with the version of the request
    mux.Handle("/users", version.Match(version.Map{"^1": apiV1(), "^2": apiV2()}))
    /* do something with the server mux */
}
```

A couple things to note. 

1. The middleware sent to the `mux.Use` method takes a parameter, this is because the media type does not _have to be_ `vnd.api+json` in fact, when researching, many applications include the their name, so parametrizing this variable gives greater flexibility. 

1. The `mux.Match` handler takes a map who's keys are adhere to [semantic versioning](https://semver.org/). More on this later.

To start I went searching if a package already existed that I could use & of course there was. I turned to a package called [kataras/versioning](https://github.com/kataras/versioning). You can see that their API is not too far off what I was looking to achieve. 

```go
router := http.NewServeMux()
router.Handle("/", versioning.NewMatcher(versioning.Map{
    // v1Handler is a handler of yours that will be executed only on version 1.
    "1.0":               v1Handler, 
    ">= 2, < 3":         v2Handler,
    versioning.NotFound: http.NotFoundHandler(),
}))
```

This is great, so why what now? Well my favourite method of learning is taking a look at the source code. With Go this activity is rewardingly easy. 

Firstly, I found that they use [hashicorp/go-version](https://github.com/hashicorp/go-version) for parsing versions (i.e __v1.3.4__), and verifying versions against a set of constraints (i.e. __>=1.2__). Unfortunately, their library doesn't support the [caret (^)](https://docs.npmjs.com/cli/v6/using-npm/semver#caret-ranges-123-025-004) syntax. This in the context for versioning means that a constraint `^3.x.y` which is equivalent to `>= 3.0.0, < 4.0.0`. Though this is equivalent, I am a lazy developer and so going for the slightly simpler syntax is favourable.

Secondly I took a dive into their `NewMatcher` function. They make an internal call to a another [public function]((https://github.com/kataras/versioning/blob/master/version.go#L55)) `GetVersion` in which most of the logic occurs. Before closing off the function by calling the next handler's `ServeHTTP` method.

Stepping into the `GetVersion` function, it's clear that this was designed to accommodate various use cases:

- [First](https://github.com/kataras/versioning/blob/master/version.go#L57-L61), checking if it has been added to the request context via middleware up the stack


- [Next](https://github.com/kataras/versioning/blob/master/version.go#L64-L66), from the `Accept-Version` header. 

- [Finally](https://github.com/kataras/versioning/blob/master/version.go#L69-L92), if neither of these options worked, it will check the `Accept` header. Failing this check it returns a `not found` variable

As the `Accept` header is generally seen as the preferred header, given it is standard to the spec, I didn't want to give the flexibility. Designing libraries for mass consumption is hard. All developers think differently, have different opinions and garnering support often means accommodating all potential use-cases regardless of correctness.

Stepping back into the `NewMatcher` function to [line 50](https://github.com/kataras/versioning/blob/master/versioning.go#L50) we see first use of the external package that will parse the requested version into valid semver syntax. 

Moving our way down to [line 56](https://github.com/kataras/versioning/blob/master/versioning.go#L56) we see ourselves comparing the version with the map that was passed as an argument before closing off the function body by calling the next handler.

```go
func NewMatcher(versions Map) http.Handler {
	constraintsHandlers, notFoundHandler := buildConstraints(versions)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		versionString := GetVersion(r)
		if versionString == NotFound {
			notFoundHandler.ServeHTTP(w, r)
			return
		}

		ver, err := version.NewVersion(versionString)
		if err != nil {
			notFoundHandler.ServeHTTP(w, r)
			return
		}

		for _, ch := range constraintsHandlers {
			if ch.constraints.Check(ver) {
				w.Header().Set("X-API-Version", ver.String())
				ch.handler.ServeHTTP(w, r)
				return
			}
		}

		notFoundHandler.ServeHTTP(w, r)
	})
}
```

> This is the full function, extremely clean and easy to follow. 

There is a lot of great patterns to be taken from this. Bear in mind, this was last touched 4years ago, and still behaves how the original author expects. 

With this then lets' get started deconstructing and creating our own library. 

## Building my solution

With the context defined, let's tackle the first thing which is my `Version` middleware handler. Many frameworks (regardless of language) work with the idea of passing down the request through one or many functions, all of which perform their own task. Those familiar with functional programming will already know this pattern called [function composition](https://en.wikipedia.org/wiki/Function_composition#:~:text=In%20mathematics%2C%20function%20composition%20is,the%20function%20f%20to%20x.) as this concepts stems from mathematics. The benefits are that each function is specialised to do one job, extremely well. More of this can be understood from this [article](https://medium.com/@pragyan88/writing-middleware-composition-and-currying-elegance-in-javascript-8b15c98a541b#:~:text=Middlewares%20are%20usually%20applied%20as,made%20up%20of%20pure%20functions.).

Walking through my first of two functions, wew only check the `Accept` header which is passed to an internal function called `parseVersion`. We also pass a variable `vendor` as previously stated, this offers extra flexibility to it's usage. A parse fail returns a 406 error else we add it to the context and move down the stack.

```go
func Vendor(vendor string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accept := r.Header.Get("Accept")

			ver, err := parseVersion(accept, vendor)
			if err != nil {
				// not acceptable
				http.Error(w, "not acceptable", http.StatusNotAcceptable)
				return
			}

			ctx := context.WithValue(r.Context(), apiVersion, ver)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
```

We now step into the `parseVersion` which reads as:

```go
func parseVersion(mediaTyp, vendor string) (*version.Version, error) {
	mediaTyp, params, err := mime.ParseMediaType(mediaTyp)
	if err != nil {
		return nil, err
	}

	if !strings.Contains(mediaTyp, vendor) {
		return nil, errors.New("not a valid media type")
	}

	v, ok := params["version"]
	if !ok {
		return nil, errors.New("no version")
	}

	return version.NewVersion(v)
}
```

The main difference between this and what the [kataras/versioning](https://github.com/kataras/versioning/blob/master/version.go#L69-L92) is doing is that I am making use of the `mime.ParseMediaType` function built into the standard library rather that attempting to roll your own parsing logic. It is usually recommended to use as much of the standard library as possible and to me this was such a case. I next check that the correct __media type__ and the __version__ parameter exists before stepping up. Again, the [kataras/versioning](https://github.com/kataras/versioning) doesn't make both checks which I do feel is unfortunate given the __media type__ is as important to validate as the __version__.

The last line in the function is the where I use an external package to handle with semantic versioning. Rather than using the __Hashicorp__ package I will be using [Masterminds/semver](https://github.com/Masterminds/semver) package instead. There both do the same job, both share similar amounts of stars. However the advantage I get is that looks to be more recently maintained as well as support for the __caret syntax__. 

With this function out the way, we move onto the second function, the `Match` handler. This is another small function which looks like

```go
func Match(cm Map) http.Handler {
	cs, err := build(cm)
	if err != nil {
		panic(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ver, ok := r.Context().Value(apiVersion).(*version.Version)
		if !ok {
			// bad request, not really
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		h, ok := match(cs, ver)
		if !ok {
			// not found
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		// This should always be valid
		major := strconv.Itoa(int(ver.Major()))
		w.Header().Set("X-API-Version", major)

		h.ServeHTTP(w, r)
	})
}
```

Once again, this is not very different but the single responsibility of the function is easy to deduce from the handler. All we do is match the version to a corresponding handler. I do some funky business when setting the header as I currently only want to show the major version that the API corresponds to. Another design decision I made was my use of panic. My rationale here is that if we fail to build then our app can't work, so there isn't a strong case of graceful handling here.

That pretty much is it. Which is good, my goal is to learn how complicated it would be to implement this verse __URI versioning__, and whether the added complexities are worth it. Answering that question is up for debate ultimately.

## Where do we go from here?

More learning. What I shares is quite a basic use of the `Accept` header. Let's take a look at an example from [Postman](https://web.postman.co/):

```md
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7
```

This is a list of accepted media, some basic types such as `text/html` some not so basic like `application-signed-exchange`. The part that I am most intrigued in, is the `q` factor you see in a few of them. This essentially the preference for that given media type. It's a lot nerdier than that and something I am digging my feet into implementing.

Thank you for enjoying my journey, it introduced me to many different areas of api development that I had previously not been aware of. I would also like to develop on this, so open to suggestions and improvements. 

You can read the full code here.