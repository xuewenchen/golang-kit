package etcd

import (
	"golang-kit/log"
	"strings"

	etcd2 "github.com/coreos/etcd/client"
	"google.golang.org/grpc/naming"
)

type resolver struct {
	serviceName string
}

func NewResolver(serviceName string) *resolver {
	return &resolver{serviceName: serviceName}
}

func (re *resolver) Resolve(target string) (w naming.Watcher, err error) {
	if re.serviceName == "" {
		log.Error("no service name provided")
		return
	}

	EtcClient, err := etcd2.New(etcd2.Config{
		Endpoints: strings.Split(target, ","),
	})
	if err != nil {
		log.Error("creat etcd2 client failed: %s", err.Error())
		return
	}

	return &watcher{re: re, api: etcd2.NewKeysAPI(EtcClient)}, nil
}
