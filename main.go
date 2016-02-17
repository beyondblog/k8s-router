package main

import (
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
	"net/http"
	"net/url"
)

func ParseURI(uri string) *url.URL {
	out, err := url.ParseRequestURI(uri)
	if err != nil {
		panic(err)
	}
	return out
}

func main() {
	fwd, _ := forward.New()
	lb, _ := roundrobin.New(fwd)

	lb.UpsertServer(ParseURI("http://ip:port"))

	s := &http.Server{
		Addr:    ":8080",
		Handler: lb,
	}
	s.ListenAndServe()
}
