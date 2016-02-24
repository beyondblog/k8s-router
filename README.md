# k8s-router
Simple HTTP router for Kubernetes

### 描述

这是一个简单基于Kubernetes运行pod的http路由,功能类似kube-proxy(但是没有它那么强大) 目前只能转发http请求(暂不支持tcp和udp)
实现原理是 从etcd中读取kubernetes中指定serverName的endpoint ip地址(watcher 改变),然后自己监听一个端口给外部做负载

使用它的好处是可以根据具体的场合去灵活的自定义负载方案(例如:往往容器启动后需要一些时间才能正常提供服务此时 k8s默认方案会负载过来这时候可能就会有类似504的错误),而且有一个简单的日志能看到请求具体负载到哪一个节点上

目前未在生产环境进行测试仅供学习参考

### 如何使用
```
$ ./k8s-router -h
NAME:
   k8s-router - Simple HTTP router for Kubernetes

USAGE:
   k8s-router [global options] command [command options] [arguments...]

VERSION:
   0.0.1

COMMANDS:
   help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --port, -p "8888"                            proxy listen port
   --service_name, -s                           proxy kubernetes serviceName
   --service_port, --sp "8080"                  proxy kubernetes service port
   --etcd_service, --es "http://master:4001"    kubernetes etcd service
   --log, -l "router.log"                       logs
   --help, -h                                   show help
   --version, -v                                print the version
```

### 手动编译

得有个go lang环境然后clone 这个项目

```
$ GOPATH=`godep path`:$GOPATH
$ go build
```

### 参考项目

[0] [https://github.com/vulcand/oxy](https://github.com/vulcand/oxy)
