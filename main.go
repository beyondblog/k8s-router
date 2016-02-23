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
	"sync"
)

func ParseURI(uri string) *url.URL {
	out, err := url.ParseRequestURI(uri)
	if err != nil {
		panic(err)
	}
	return out
}

type endpointResponse struct {
	resp string
	err  error
}

func UpsertServer(lb *roundrobin.RoundRobin, endpoints string, servicePort int) {
	nodeJson, _ := json.NewJson([]byte(endpoints))

	subsets := nodeJson.Get("subsets").GetIndex(0)

	if subsets.Interface() != nil {

		addresses, _ := subsets.Get("addresses").Array()
		for _, address := range addresses {
			url := fmt.Sprintf("http://%s:%d", address.(map[string]interface{})["ip"].(string), servicePort)
			fmt.Println("upsertServer: " + url)

			lb.UpsertServer(ParseURI(url))
		}
	}
}

func StartProxy(kubernetes *k8s.K8s, port int, serviceName string, servicePort int) {

	fwd, _ := forward.New()
	lb, _ := roundrobin.New(fwd)
	var mutex = &sync.Mutex{}

	endpoints, err := kubernetes.GetNodeEnpoints(serviceName)
	if err != nil {
		log.Print("serverName not found!")
	}

	UpsertServer(lb, endpoints, servicePort)

	addr := fmt.Sprintf(":%d", port)

	epchan := make(chan endpointResponse, 1)

	go func(kubernetes *k8s.K8s) {
		for {
			fmt.Println("watcher endpoints ...")
			endpoints, err := kubernetes.WatcherEndpoints(serviceName)
			fmt.Println("endpoints:" + endpoints)
			epchan <- endpointResponse{resp: endpoints, err: err}
		}
	}(kubernetes)

	go func() {
		s := &http.Server{
			Addr:    addr,
			Handler: lb,
		}

		log.Print("Listen " + addr)

		if err := s.ListenAndServe(); err != nil {
			log.Print("Listen Error!")
		}
	}()

	for {
		select {
		case rtresp := <-epchan:
			resp := rtresp.resp
			mutex.Lock()
			UpsertServer(lb, resp, servicePort)
			mutex.Unlock()
		}
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
