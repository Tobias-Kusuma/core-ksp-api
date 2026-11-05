package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var Ctx_bg = context.Background()

var Redis *redis.Client

func init() {
	InitRedisClient()
}

// InitRedisClient menginisialisasi atau memperbarui client Redis berdasarkan environment variables
func InitRedisClient() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		host := os.Getenv("REDIS_HOST")
		if host == "" {
			host = "redis"
		}

		port := os.Getenv("REDIS_PORT")
		if port == "" {
			port = "6379"
		}

		addr = fmt.Sprintf("%s:%s", host, port)
	}

	password := os.Getenv("REDIS_PASSWORD")

	dbNum := 0
	if dbEnv := os.Getenv("REDIS_DB"); dbEnv != "" {
		if val, err := strconv.Atoi(dbEnv); err == nil {
			dbNum = val
		}
	}

	if Redis != nil {
		_ = Redis.Close()
	}

	Redis = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       dbNum,
	})
}

// PingRedis mengecek koneksi ke Redis dengan batas waktu timeout
func PingRedis(timeout time.Duration) error {
	if Redis == nil {
		InitRedisClient()
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return Redis.Ping(ctx).Err()
}

// CloseRedis menutup koneksi client Redis
func CloseRedis() error {
	if Redis != nil {
		return Redis.Close()
	}
	return nil
}

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
