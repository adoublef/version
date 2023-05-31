package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	v1 "example/internal/http/v1"
	v2 "example/internal/http/v2"

	v "github.com/adoublef-go/version"
	"github.com/go-chi/chi/v5"
)

var PORT = os.Getenv("PORT")

func init() {
	if PORT == "" {
		PORT = "8080"
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	q := make(chan os.Signal, 1)

	signal.Notify(q, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-q
		cancel()
	}()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	v1, v2 := v1.NewService(), v2.NewService()

	mux := chi.NewMux()
	mux.Use(v.Version("vnd.api+json"))
	mux.Mount("/", v.Match(v.Map{">=1": v1, "2": v2}))

	srv := &http.Server{
		Addr:        ":" + PORT,
		Handler:     mux,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	e := make(chan error, 1)

	go func() { e <- srv.ListenAndServe() }()

	select {
	case <-ctx.Done():
		return srv.Shutdown(ctx)
	case err := <-e:
		return err
	}
}
