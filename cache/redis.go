package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

var Ctx_bg = context.Background()

var Redis = redis.NewClient(&redis.Options{
	Addr:     "redis:6379",
	Password: "",
	DB:       0,
})

// set item ke cache;
// input parms key, lalu instance of struct/map sebagai value, dan time to live;
func RedisSet[T any](key string, value T, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return Redis.Set(Ctx_bg, key, data, ttl).Err()
}

// ambil item dari cache;
// input parms berupa key;
// output berupa data/value, isFound (true/false), dan error;
func RedisGet[T any](key string) (T, bool, error) {
	var result T

	data, err := Redis.Get(Ctx_bg, key).Bytes()

	if err == redis.Nil {
		return result, false, nil
	}

	if err != nil {
		return result, false, err
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return result, false, err
	}

	return result, true, nil
}

// hapus item dari cache;
func RedisDel(key string) error {
	return Redis.Del(Ctx_bg, key).Err()
}
