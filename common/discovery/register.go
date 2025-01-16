package dicovery

import (
	"context"

	"github.com/brelance/plato/common/config"
	"github.com/brelance/plato/common/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type ServiceRegister struct {
	cli           *clientv3.Client
	leaseID       clientv3.LeaseID
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	key           string
	val           string
	ctx           *context.Context
}

func NewServiceRegister(ctx *context.Context, key string, endportinfo *EndpointInfo, lease int64) (*ServiceRegister, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   config.GetEndpointsForDiscovery(),
		DialTimeout: config.GetDTimeoutForDiscovery(),
	})

	if err != nil {
		// logger
		return nil, err
	}

	s := &ServiceRegister{
		cli: cli,
		key: key,
		val: endportinfo.Marshal(),
		ctx: ctx,
	}

	err = s.putKeyWithLease(lease)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *ServiceRegister) putKeyWithLease(lease int64) error {
	// set the time of register
	resp, err := s.cli.Grant(*s.ctx, lease)
	if err != nil {
		return err
	}

	_, err = s.cli.Put(*s.ctx, s.key, s.val, clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}

	leaseRespChan, err := s.cli.KeepAlive(*s.ctx, resp.ID)
	if err != nil {
		return err
	}
	s.leaseID = resp.ID
	s.keepAliveChan = leaseRespChan
	return nil
}

func (s *ServiceRegister) UpdateValue(val *EndpointInfo) error {
	value := val.Marshal()

	_, err := s.cli.Put(*s.ctx, s.key, value, clientv3.WithLease(s.leaseID))
	if err != nil {
		return err
	}
	s.val = value
	// print log info here
	logger.Logger.
		Info().
		Msgf("ServiceRegister.updateValue leaseID=%d Put key=%s,val=%s, success!", s.leaseID, s.key, s.val)
	return nil
}

func (s *ServiceRegister) ListenLeaseRespChan() {
	for leaseKeepResp := range s.keepAliveChan {
		logger.Logger.
			Info().
			Msgf("lease success leaseID:%d, Put key:%s,val:%s reps:+%v",
				s.leaseID, s.key, s.val, leaseKeepResp)
	}

	logger.Logger.
		Info().
		Msgf("lease failed !!!  leaseID:%d, Put key:%s,val:%s", s.leaseID, s.key, s.val)
}

func (s *ServiceRegister) Close() error {
	_, err := s.cli.Revoke(*s.ctx, s.leaseID)
	if err != nil {
		return err
	}
	logger.Logger.
		Info().
		Msgf("lease close !!!  leaseID:%d, Put key:%s,val:%s  success!", s.leaseID, s.key, s.val)

	return nil
}
