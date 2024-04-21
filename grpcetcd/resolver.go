package grpcetcd

import (
	"context"
	"strings"
	"sync"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc/codes"
	gresolver "google.golang.org/grpc/resolver"
	"google.golang.org/grpc/status"

	"github.com/tenz-io/gokit/logger"
)

const (
	etcdBuilderScheme = "etcd"
)

type builder struct {
	c  *clientv3.Client
	le logger.Entry
}

func (b builder) Build(target gresolver.Target, cc gresolver.ClientConn, opts gresolver.BuildOptions) (gresolver.Resolver, error) {
	// Refer to https://github.com/grpc/grpc-go/blob/16d3df80f029f57cff5458f1d6da6aedbc23545d/clientconn.go#L1587-L1611
	endpoint := target.URL.Path
	if endpoint == "" {
		endpoint = target.URL.Opaque
	}
	endpoint = strings.TrimPrefix(endpoint, "/")
	r := &resolver{
		c:      b.c,
		target: endpoint,
		cc:     cc,
		le:     b.le,
	}
	r.ctx, r.cancel = context.WithCancel(context.Background())

	em, err := endpoints.NewManager(r.c, r.target)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "resolver: failed to new endpoint manager: %s", err)
	}
	r.wch, err = em.NewWatchChannel(r.ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "resolver: failed to new watch channer: %s", err)
	}

	r.wg.Add(1)
	go r.watch()
	return r, nil
}

func (b builder) Scheme() string {
	return etcdBuilderScheme
}

// NewBuilder creates a resolver builder.
func NewBuilder(client *clientv3.Client, le logger.Entry) (gresolver.Builder, error) {
	return builder{
		c:  client,
		le: le,
	}, nil
}

type resolver struct {
	c      *clientv3.Client
	target string
	cc     gresolver.ClientConn
	wch    endpoints.WatchChannel
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	le     logger.Entry
}

func (r *resolver) watch() {
	defer r.wg.Done()

	allUps := make(map[string]*endpoints.Update)
	for {
		select {
		case <-r.ctx.Done():
			return
		case ups, ok := <-r.wch:
			le := r.le.WithFields(logger.Fields{
				"target": r.target,
			})
			if !ok {
				le.Warn("resolver: watch channel closed")
				return
			}

			for _, up := range ups {
				le = le.WithFields(logger.Fields{
					"up": up,
				})
				switch up.Op {
				case endpoints.Add:
					le.Debugf("resolver: add op, endpoint")
					allUps[up.Key] = up
				case endpoints.Delete:
					le.Debugf("resolver: delete op, endpoint")
					delete(allUps, up.Key)
				}
			}

			addrs := convertToGRPCAddress(allUps)
			err := r.cc.UpdateState(gresolver.State{
				Addresses: addrs,
			})
			if err != nil {
				le.WithError(err).Warn("resolver: update state error")
				r.cc.ReportError(err)
			}
		}
	}
}

func convertToGRPCAddress(ups map[string]*endpoints.Update) []gresolver.Address {
	var addrs []gresolver.Address
	for _, up := range ups {
		addr := gresolver.Address{
			Addr:     up.Endpoint.Addr,
			Metadata: up.Endpoint.Metadata,
		}
		addrs = append(addrs, addr)
	}
	return addrs
}

// ResolveNow is a no-op here.
// It's just a hint, resolver can ignore this if it's not necessary.
func (r *resolver) ResolveNow(gresolver.ResolveNowOptions) {}

func (r *resolver) Close() {
	r.cancel()
	r.wg.Wait()
}
