package k8s

import (
	"github.com/beyondblog/k8s-router/Godeps/_workspace/src/github.com/coreos/etcd/client"
	"github.com/beyondblog/k8s-router/Godeps/_workspace/src/golang.org/x/net/context"
	"time"
)

type K8s struct {
	etcdKeysApi client.KeysAPI
}

func New(etcdService string) (*K8s, error) {
	kubernetes := &K8s{}

	cfg := client.Config{
		Endpoints: []string{etcdService},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}

	c, err := client.New(cfg)
	if err != nil {
		return nil, err
	}

	kapi := client.NewKeysAPI(c)
	kubernetes.etcdKeysApi = kapi

	return kubernetes, nil

}

func (k *K8s) GetNodeEnpoints(serviceName string) (string, error) {

	resp, err := k.etcdKeysApi.Get(context.Background(), "/registry/services/endpoints/default/"+serviceName, nil)
	if err != nil {
		return "", err
	}

	return resp.Node.Value, nil
}

func (k *K8s) WatcherEndpoints(serviceName string) (string, error) {

	watcher := k.etcdKeysApi.Watcher("/registry/services/endpoints/default/"+serviceName, &client.WatcherOptions{})

	resp, err := watcher.Next(context.Background())
	if err != nil {
		return "", err
	}

	return resp.Node.Value, nil
}
