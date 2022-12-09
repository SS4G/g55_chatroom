package db

import (
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

func RedisTestMain() {
	addr := "127.0.0.1:12000"
	opt := redis.Options{Addr: addr, DB: 0}
	redis_cli := redis.NewClient(&opt)
	res := redis_cli.Get("aa").String()
	log.Infof("redis_result=%s", res)
}
