package db

import "github.com/vera-byte/vera_im_lib/pkg/redis"

func NewRedis(addr string, password string) *redis.Conn {
	return redis.New(addr, password)
}
