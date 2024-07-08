package app

import (
	config "commune/config"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/tidwall/buntdb"
)

type Cache struct {
	JoinedRooms *buntdb.DB
	Rooms       *redis.Client
}

func NewCache(conf *config.Config) (*Cache, error) {

	db, err := buntdb.Open(":memory:")
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Address,
		Password: conf.Redis.Password,
		DB:       conf.Redis.RoomsDB,
	})

	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(fmt.Sprintf("Could not connect to Redis: %v", err))
	}

	c := &Cache{
		JoinedRooms: db,
		Rooms:       rdb,
	}

	return c, nil
}
