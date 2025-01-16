package dicovery

import (
	"context"
	"sync"

	"github.com/brelance/plato/common/config"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type ServiceDiscovery struct {
	cli clientv3.Client
	// why we should add lock in here
	lock sync.Mutex
	ctx  *context.Context
}

func NewServiceDiscovery(ctx *context.Context) *ServiceDiscovery {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   config.GetEndpointsForDiscovery(),
		DialTimeout: config.GetDTimeoutForDiscovery(),
	})

	if err != nil {
		panic(err)
	}

	return &ServiceDiscovery{
		cli: *cli,
		ctx: ctx,
	}
}

func (s *ServiceDiscovery) WathchService(prefix string, set, del func(key, value string)) error {
	resp, err := s.cli.Get(*s.ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}

	// init
	for _, ev := range resp.Kvs {
		set(string(ev.Key), string(ev.Value))
	}

	s.watch(prefix, resp.Header.Revision+1, set, del)
	return nil
}

// why we need revision in here
func (s *ServiceDiscovery) watch(prefix string, rev int64, set, del func(key, value string)) {
	rch := s.cli.Watch(*s.ctx, prefix, clientv3.WithRev(rev))
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				set(string(ev.Kv.Key), string(ev.Kv.Value))
			case mvccpb.DELETE:
				del(string(ev.Kv.Key), string(ev.Kv.Value))
			}
		}
	}
}

func (s *ServiceDiscovery) Close() error {
	return s.cli.Close()
}
