package etcd

import (
	"context"
	"fmt"
	etcd2 "github.com/coreos/etcd/client"
	"golang-kit/log"
	xip "golang-kit/net/ip"
	"strings"
	"time"
)

var (
	Prefix     = "micro"
	EtcClient  etcd2.Client
	serviceKey string
	interval   = time.Second * 10
	ttl        = 15
	stopSignal = make(chan bool, 1)
)

func Register(name string, host string, port int, target string) (err error) {
	// 获得本机内网ip
	if host == "127.0.0.1" || host == "0.0.0.0" {
		host = xip.InternalIP()
	}
	serviceValue := fmt.Sprintf("%s:%d", host, port)
	serviceKey = fmt.Sprintf("/%s/%s/%s", Prefix, name, serviceValue)
	EtcClient, err = etcd2.New(etcd2.Config{
		Endpoints: strings.Split(target, ","),
	})
	if err != nil {
		log.Error("create etcd2 client failed: %v", err.Error())
		return
	}
	keysAPI := etcd2.NewKeysAPI(EtcClient)
	go func() {
		ticker := time.NewTicker(interval)
		setOptions := &etcd2.SetOptions{TTL: time.Second * time.Duration(ttl), Refresh: true, PrevExist: etcd2.PrevExist}
		for {
			_, err := keysAPI.Get(context.Background(), serviceKey, &etcd2.GetOptions{Recursive: true})
			if err != nil {
				if etcd2.IsKeyNotFound(err) {
					if _, err := keysAPI.Set(context.Background(), serviceKey, serviceValue, &etcd2.SetOptions{TTL: time.Second * time.Duration(ttl)}); err != nil {
						log.Error("set service %s with ttl to etcd2 failed: %s", name, err.Error())
					}
				} else {
					log.Error("set service %s with ttl to etcd2 failed: %s", name, err.Error())
				}
			} else {
				if _, err := keysAPI.Set(context.Background(), serviceKey, "", setOptions); err != nil {
					log.Error("refresh service %s with ttl to etcd2 failed: %s", name, err.Error())
				}
			}
			select {
			case <-stopSignal:
				return
			case <-ticker.C:
			}
		}
	}()

	return nil
}

func UnRegister() error {
	stopSignal <- true
	stopSignal = make(chan bool, 1)
	_, err := etcd2.NewKeysAPI(EtcClient).Delete(context.Background(), serviceKey, &etcd2.DeleteOptions{Recursive: true})
	if err != nil {
		log.Error("deregister '%s' failed: %s", serviceKey, err.Error())
	}
	return err
}
