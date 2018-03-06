package etcd

import (
	"fmt"
	"strings"

	etcd2 "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc/naming"
)

type watcher struct {
	re            *resolver
	api           etcd2.KeysAPI
	isInitialized bool
}

func (w *watcher) Close() {
}

func (w *watcher) Next() ([]*naming.Update, error) {
	dir := fmt.Sprintf("/%s/%s/", Prefix, w.re.serviceName)
	if !w.isInitialized {
		resp, err := w.api.Get(context.Background(), dir, &etcd2.GetOptions{Recursive: true})
		w.isInitialized = true
		if err == nil {
			addrs := extractAddrs(resp)
			if l := len(addrs); l != 0 {
				updates := make([]*naming.Update, l)
				for i := range addrs {
					updates[i] = &naming.Update{Op: naming.Add, Addr: addrs[i]}
				}
				return updates, nil
			}
		}
	}
	etcdWatcher := w.api.Watcher(dir, &etcd2.WatcherOptions{Recursive: true})
	for {
		resp, err := etcdWatcher.Next(context.Background())
		if err == nil {
			switch resp.Action {
			case "set":
				return []*naming.Update{{Op: naming.Add, Addr: resp.Node.Value}}, nil
			case "delete", "expire":
				return []*naming.Update{{Op: naming.Delete, Addr: strings.TrimPrefix(resp.Node.Key, dir)}}, nil // not using PrevNode because it may nil
			}
		}
	}
}

func extractAddrs(resp *etcd2.Response) []string {
	addrs := []string{}
	if resp == nil || resp.Node == nil {
		return addrs
	}
	for i := range resp.Node.Nodes {
		if v := resp.Node.Nodes[i].Value; v != "" {
			addrs = append(addrs, v)
		}
	}
	return addrs
}
