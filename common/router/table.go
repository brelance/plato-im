package router

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/brelance/plato/common/cache"
)

const (
	gateaWayRouterKey = "gateway_router_%d"
	TTL7D             = 7 * 24 * 60 * 60
)

type Record struct {
	Endpoint string
	ConnID   uint64
}

func Init(ctx context.Context) {
	cache.InitRedis(ctx)
}

func AddRecord(ctx context.Context, did uint64, endpoint string, connID uint64) error {
	key := fmt.Sprintf(gateaWayRouterKey, did)
	value := fmt.Sprintf("%s-%d", endpoint, connID)
	err := cache.SetString(ctx, key, value, TTL7D)
	if err != nil {
		return err
	}
	return nil
}

func DelRecord(ctx context.Context, did uint64) error {
	key := fmt.Sprintf(gateaWayRouterKey, did)
	err := cache.Del(ctx, key)
	if err != nil {
		return err
	}
	return nil
}

func QueryRecord(ctx context.Context, did uint64) (*Record, error) {
	key := fmt.Sprintf(gateaWayRouterKey, did)
	value, err := cache.GetString(ctx, key)
	if err != nil {
		return nil, err
	}

	ec := strings.Split(value, "-")
	connID, err := strconv.ParseUint(ec[1], 10, 64)
	if err != nil {
		return nil, err
	}
	return &Record{
		Endpoint: ec[0],
		ConnID:   connID,
	}, nil
}
