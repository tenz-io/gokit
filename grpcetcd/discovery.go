package grpcetcd

import (
	"context"
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
)

const (
	EtcdBuilderScheme = "etcd"
)

type Discovery struct {
	etcdClient *clientv3.Client
	target     string // service target in etcd
	dialOpts   []grpc.DialOption
}

func NewDiscovery(
	cli *clientv3.Client,
	path string, // path in etcd, e.g: /services/imagesearch-rank
	dialOpts ...grpc.DialOption,
) *Discovery {
	return &Discovery{
		etcdClient: cli,
		target:     fmt.Sprintf("%s://%s", EtcdBuilderScheme, path),
		dialOpts:   dialOpts,
	}
}

// Dial connects to the target service
func (d *Discovery) Dial(ctx context.Context) (cc *grpc.ClientConn, err error) {
	etcdResolver, err := resolver.NewBuilder(d.etcdClient)
	d.dialOpts = append(d.dialOpts, grpc.WithResolvers(etcdResolver))
	d.dialOpts = append(d.dialOpts, grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`))
	d.dialOpts = append(d.dialOpts, grpc.WithBlock())
	d.dialOpts = append(d.dialOpts, grpc.WithInsecure())
	return grpc.DialContext(ctx, d.target, d.dialOpts...)
}
