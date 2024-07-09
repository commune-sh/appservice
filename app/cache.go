package app

import (
	config "commune/config"
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"maunium.net/go/mautrix/id"
)

type Cache struct {
	Rooms *redis.Client
}

func NewCache(conf *config.Config) (*Cache, error) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Address,
		Password: conf.Redis.Password,
		DB:       conf.Redis.RoomsDB,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(fmt.Sprintf("Could not connect to Redis: %v", err))
	}

	c := &Cache{
		Rooms: rdb,
	}

	return c, nil
}

func (c *App) AddRoomToCache(room *PublicRoom) error {

	i, err := json.Marshal(room)
	if err != nil {
		c.Log.Error().Msgf("Couldn't marshal room info %v", err)
		return err
	}

	err = c.Cache.Rooms.Set(context.Background(), room.RoomID, string(i), 0).Err()
	if err != nil {
		c.Log.Error().Msgf("Couldn't cache room %v", err)
		return err
	}

	err = c.Cache.Rooms.Set(context.Background(), room.CanonicalAlias, room.RoomID, 0).Err()
	if err != nil {
		c.Log.Error().Msgf("Couldn't cache room %v", err)
		return err
	}

	return nil
}

func (c *App) RemoveRoomFromCache(room_id id.RoomID) error {

	c.Log.Info().Msgf("Removing room from cache: %v", room_id)

	room, err := c.GetRoomInfo(room_id.String())
	if err != nil {
		return err
	}

	err = c.Cache.Rooms.Del(context.Background(), room.CanonicalAlias).Err()
	if err != nil {
		c.Log.Error().Msgf("Couldn't remove room alias from cache %v", err)
		return err
	}

	err = c.Cache.Rooms.Del(context.Background(), room.RoomID).Err()
	if err != nil {
		c.Log.Error().Msgf("Couldn't remove room ID from cache %v", err)
		return err
	}

	return nil
}
