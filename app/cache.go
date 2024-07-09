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

func (c *App) AddRoomToCache(room *PublicRoom) error {

	i, err := json.Marshal(room)
	if err != nil {
		c.Log.Error().Msgf("Couldn't marshal room info %v", err)
		return err
	}

	err = c.Cache.Rooms.SAdd(context.Background(), "ids", room.RoomID).Err()
	if err != nil {
		c.Log.Error().Msgf("Couldn't cache room id %v", err)
		return err
	}

	err = c.Cache.Rooms.SAdd(context.Background(), "aliases", room.CanonicalAlias).Err()
	if err != nil {
		c.Log.Error().Msgf("Couldn't cache room alias %v", err)
		return err
	}

	err = c.Cache.Rooms.Set(context.Background(), room.RoomID, string(i), 0).Err()
	if err != nil {
		c.Log.Error().Msgf("Couldn't cache room %v", err)
		return err
	}

	return nil
}

func (c *App) RemoveRoomFromCache(room_id id.RoomID) error {
	c.Log.Info().Msgf("Removing room from cache: %v", room_id)
	c.Cache.JoinedRooms.Update(func(tx *buntdb.Tx) error {
		tx.Delete(room_id.String())
		return nil
	})

	err := c.Cache.Rooms.Del(context.Background(), room_id.String()).Err()
	if err != nil {
		c.Log.Error().Msgf("Couldn't cache room %v", err)
		return err
	}

	return nil
}
