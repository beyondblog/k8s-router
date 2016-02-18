package main

import (
	"fmt"
	json "github.com/bitly/go-simplejson"
	"github.com/codegangsta/cli"
	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/etcd/client"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

func ParseURI(uri string) *url.URL {
	out, err := url.ParseRequestURI(uri)
	if err != nil {
		panic(err)
	}
	return out
}

func StartProxy(port int, serviceName string, servicePort int) {

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
	resp, err := kapi.Get(context.Background(), "/registry/services/endpoints/default/"+serviceName, nil)
	if err != nil {
		log.Print("serverName not found!")
		log.Fatal(err)
	}
	nodeJson, _ := json.NewJson([]byte(resp.Node.Value))

	addresses, _ := nodeJson.Get("subsets").GetIndex(0).Get("addresses").Array()
	for _, address := range addresses {
		url := fmt.Sprintf("http://%s:%d", address.(map[string]interface{})["ip"].(string), servicePort)

		lb.UpsertServer(ParseURI(url))
	}

	addr := fmt.Sprintf(":%d", port)

	s := &http.Server{
		Addr:    addr,
		Handler: lb,
	}

	log.Print("Listen " + addr)

	if err := s.ListenAndServe(); err != nil {
		log.Print("Listen Error!")
	}

}

func main() {

	app := cli.NewApp()
	app.Name = "k8s-router"
	app.Usage = "Simple HTTP router for Kubernetes"
	app.Version = "0.0.1"

	var (
		port        int
		serviceName string
		servicePort int
	)

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:        "port, p",
			Value:       8888,
			Usage:       "proxy listen port",
			Destination: &port,
		},
		cli.StringFlag{
			Name:        "serviceName, s",
			Usage:       "proxy kubernetes serviceName",
			Destination: &serviceName,
		},
		cli.IntFlag{
			Name:        "servicePort, sp",
			Value:       8080,
			Usage:       "proxy kubernetes service port",
			Destination: &servicePort,
		},
	}

	app.Action = func(c *cli.Context) {
		if len(serviceName) == 0 {
			log.Fatal("serverName is require!")
		}
		StartProxy(port, serviceName, servicePort)
	}

	app.Run(os.Args)
}
