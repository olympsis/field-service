package db

import (
	"log"

	"github.com/gomodule/redigo/redis"
)

const (
	FIELD_KEY = "fields"
)

type RedisContext struct {
	pool *redis.Pool
}

func (r *RedisContext) Get() redis.Conn {
	return r.pool.Get()
}

func (r *RedisContext) Ping() (string, error) {
	conn := r.Get()
	defer conn.Close()
	return redis.String(conn.Do("PING"))
}

func (r *RedisContext) GeoAdd(lat float64, long float64, v string) (int, error) {
	conn := r.Get()
	defer conn.Close()
	return redis.Int(conn.Do("GEOADD", FIELD_KEY, lat, long, v))
}

func (r *RedisContext) RemoveIndex(key string) (int, error) {
	conn := r.Get()
	defer conn.Close()
	return redis.Int(conn.Do("ZREM", FIELD_KEY, key))
}

func (r *RedisContext) GeoRadius(lat float64, long float64, rad float64) ([]string, error) {
	conn := r.Get()
	defer conn.Close()
	values, err := redis.Strings(conn.Do("GEORADIUS", FIELD_KEY, long, lat, rad, "mi", "ASC"))
	if err != nil {
		return []string{}, err
	}
	return values, nil
}

func (r *RedisContext) waitForRedis() {
	conn := r.Get()
	defer conn.Close()
	for {
		log.Println("Pinging Redis ...")
		pong, err := r.Ping()
		if err == nil {
			log.Printf("Redis says %s", pong)
			break
		}
	}
}

func MakeRedisContext() *RedisContext {
	pool := &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", "192.168.1.225:6379")
			if err != nil {
				log.Fatal(err.Error())
			}
			return conn, err
		},
	}
	ctx := &RedisContext{pool}
	ctx.waitForRedis()
	return ctx
}
