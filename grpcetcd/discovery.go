package grpcetcd

import (
	"context"
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	"github.com/tenz-io/gokit/logger"
)

type Discovery struct {
	etcdClient *clientv3.Client
	target     string // service target in etcd
	dialOpts   []grpc.DialOption
	le         logger.Entry
}

func NewDiscovery(
	cli *clientv3.Client,
	path string, // path in etcd, e.g: /services/imagesearch-rank
	le logger.Entry,
	dialOpts ...grpc.DialOption,
) *Discovery {
	if le == nil {
		le = logger.WithFields(logger.Fields{})
	}

	return &Discovery{
		etcdClient: cli,
		target:     fmt.Sprintf("%s:///%s", etcdBuilderScheme, path),
		dialOpts:   dialOpts,
		le:         le,
	}
}

// Dial connects to the target service
func (d *Discovery) Dial(ctx context.Context) (cc *grpc.ClientConn, err error) {
	etcdResolver, err := NewBuilder(d.etcdClient, d.le)
	d.dialOpts = append(d.dialOpts, grpc.WithResolvers(etcdResolver))
	d.dialOpts = append(d.dialOpts, grpc.WithBlock())
	d.dialOpts = append(d.dialOpts, grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`))
	d.dialOpts = append(d.dialOpts, grpc.WithInsecure())
	return grpc.DialContext(ctx, d.target, d.dialOpts...)
}
