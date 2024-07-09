package app

import (
	config "commune/config"
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"maunium.net/go/mautrix/id"

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

func (c *App) AddRoomToCache(room_id id.RoomID) error {
	c.Log.Info().Msgf("Caching joined room: %v", room_id.String())
	c.Cache.JoinedRooms.Update(func(tx *buntdb.Tx) error {
		tx.Set(room_id.String(), "true", nil)
		return nil
	})

	info, err := c.GetRoomInfo(room_id.String())
	if err != nil {
		c.Log.Error().Msgf("Could not fetch room info: %v", err)
		return err
	}

	if info != nil {

		i, err := json.Marshal(info)
		if err != nil {
			c.Log.Error().Msgf("Couldn't marshal room info %v", err)
			return err
		}

		err = c.Cache.Rooms.Set(context.Background(), room_id.String(), string(i), 0).Err()
		if err != nil {
			c.Log.Error().Msgf("Couldn't cache room %v", err)
			return err
		}
	}

	return nil
}
func (c *App) RemoveRoomFromCache(room_id id.RoomID) {
	c.Log.Info().Msgf("Removing room from cache: %v", room_id)
	c.Cache.JoinedRooms.Update(func(tx *buntdb.Tx) error {
		tx.Delete(room_id.String())
		return nil
	})
}
