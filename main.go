package main

import (
	"fmt"
	json "github.com/bitly/go-simplejson"
	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/etcd/client"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
	"log"
	"net/http"
	"net/url"
	"time"
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

	cfg := client.Config{
		Endpoints: []string{"http://master:4001"},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}

	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	kapi := client.NewKeysAPI(c)
	resp, err := kapi.Get(context.Background(), "/registry/services/endpoints/default/{server-name}", nil)
	if err != nil {
		log.Fatal(err)
	}
	nodeJson, _ := json.NewJson([]byte(resp.Node.Value))

	addresses, _ := nodeJson.Get("subsets").GetIndex(0).Get("addresses").Array()
	for _, address := range addresses {
		url := fmt.Sprintf("http://%s:8080", address.(map[string]interface{})["ip"].(string))

		lb.UpsertServer(ParseURI(url))
	}

	s := &http.Server{
		Addr:    ":8888",
		Handler: lb,
	}

	log.Print("Listen port 8888")

	if err := s.ListenAndServe(); err != nil {
		log.Print("Listen Error!")
	}

}
