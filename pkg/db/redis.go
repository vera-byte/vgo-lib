package db

import "github.com/vera-byte/vgo-lib/pkg/redis"

func NewRedis(addr string, password string) *redis.Conn {
	return redis.New(addr, password)
}
