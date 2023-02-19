package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func main() {
	//new redis client for test
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6380",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	//set key
	rdb.Set(context.Background(), "pidb", "foo", 0)

	val, err := rdb.Get(context.Background(), "pidb").Result()
	if err != nil {
		panic(err)
	}

	fmt.Println("pidb", val)
}
