package main

import (
	"fmt"
	json "github.com/beyondblog/k8s-router/Godeps/_workspace/src/github.com/bitly/go-simplejson"
	"github.com/beyondblog/k8s-router/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/beyondblog/k8s-router/Godeps/_workspace/src/github.com/vulcand/oxy/forward"
	"github.com/beyondblog/k8s-router/Godeps/_workspace/src/github.com/vulcand/oxy/roundrobin"
	"github.com/beyondblog/k8s-router/k8s"
	"log"
	"net/http"
	"net/url"
	"os"
)

func ParseURI(uri string) *url.URL {
	out, err := url.ParseRequestURI(uri)
	if err != nil {
		panic(err)
	}
	return out
}

func StartProxy(kubernetes *k8s.K8s, port int, serviceName string, servicePort int) {

	fwd, _ := forward.New()
	lb, _ := roundrobin.New(fwd)

	endpoints, err := kubernetes.GetNodeEnpoints(serviceName)
	if err != nil {
		log.Print("serverName not found!")
	}

	nodeJson, _ := json.NewJson([]byte(endpoints))

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
		etcdService string
	)

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:        "port, p",
			Value:       8888,
			Usage:       "proxy listen port",
			Destination: &port,
		},
		cli.StringFlag{
			Name:        "service_name, s",
			Usage:       "proxy kubernetes serviceName",
			Destination: &serviceName,
		},
		cli.IntFlag{
			Name:        "service_port, sp",
			Value:       8080,
			Usage:       "proxy kubernetes service port",
			Destination: &servicePort,
		},
		cli.StringFlag{
			Name:        "etcd_service, es",
			Value:       "http://master:4001",
			Usage:       "kubernetes etcd service",
			Destination: &etcdService,
		},
	}

	app.Action = func(c *cli.Context) {
		if len(serviceName) == 0 {
			log.Fatal("serverName is require!")
		}

		kubernetes, err := k8s.New(etcdService)
		if err != nil {
			log.Fatal(" init k8s error ")
		}

		StartProxy(kubernetes, port, serviceName, servicePort)
	}

	app.Run(os.Args)
}
