package grpcetcd

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"

	"github.com/tenz-io/gokit/logger"
)

type Registry struct {
	etcdClient *clientv3.Client
	path       string
	ttl        int64
	le         logger.Entry
}

func NewRegistry(
	cli *clientv3.Client,
	path string,
	ttl int64, // ttl in seconds
	le logger.Entry,
) *Registry {
	if ttl <= 0 {
		ttl = 15
	}
	if le == nil {
		le = logger.WithFields(logger.Fields{})
	}

	return &Registry{
		etcdClient: cli,
		path:       path,
		ttl:        ttl,
		le:         le,
	}
}

// Register registers the service with etcd
// addr: the address of the service, eg: 10.10.10.10:50051
// returns a revoke function to deregister the service
func (r *Registry) Register(ctx context.Context, addr string) (revoke func(context.Context), err error) {
	var (
		revokeFunc = func(_ context.Context) {}
		fullPath   = fmt.Sprintf("%s/%s", r.path, addr)
		endpoint   = endpoints.Endpoint{
			Addr: addr,
		}
		le = r.le.WithFields(logger.Fields{
			"path":     r.path,
			"Addr":     addr,
			"fullPath": fullPath,
			"endpoint": endpoint,
		})
	)

	em, err := endpoints.NewManager(r.etcdClient, r.path)
	if err != nil {
		return nil, fmt.Errorf("etcd client endpoints manager error: %w", err)
	}

	grantCtx, grantCancel := context.WithTimeout(ctx, 5*time.Second)
	defer grantCancel()

	lease, err := r.etcdClient.Grant(grantCtx, r.ttl)
	if err != nil {
		return revokeFunc, fmt.Errorf("etcd client grant error: %w", err)
	}

	le = le.WithField("leaseID", lease.ID)

	putCtx, putCancel := context.WithTimeout(ctx, 5*time.Second)
	defer putCancel()

	err = em.AddEndpoint(putCtx, fullPath, endpoint, clientv3.WithLease(lease.ID))
	if err != nil {
		return nil, fmt.Errorf("etcd client add endpoint error: %w", err)
	}

	le.Infof("etcd client add endpoint ok")

	keepAliveC, err := r.etcdClient.KeepAlive(ctx, lease.ID)
	if err != nil {
		return revokeFunc, fmt.Errorf("etcd client keep alive error: %w", err)
	}

	doneC := make(chan struct{})
	go func() {
		for {
			select {
			case <-doneC:
				le.Info("etcd client keep alive done")
				return
			case <-ctx.Done():
				le.Info("etcd client keep alive context done")
				return
			case _, ok := <-keepAliveC:
				if !ok {
					le.Warnf("etcd client keep alive channel closed")
					return
				}
			}
		}
	}()

	return func(c context.Context) {
		close(doneC)
		revokeResp, revokeErr := r.etcdClient.Revoke(c, lease.ID)
		if revokeErr != nil {
			le.WithError(revokeErr).Error("etcd client revoke error")
		} else {
			le.WithField("resp", revokeResp).Info("etcd client revoke ok")
		}
	}, nil
}
