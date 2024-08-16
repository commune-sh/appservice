package app

import (
	config "commune/config"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"maunium.net/go/mautrix/id"
)

type Cache struct {
	Rooms    *redis.Client
	Events   *redis.Client
	Messages *redis.Client
	State    *redis.Client
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

	edb := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Address,
		Password: conf.Redis.Password,
		DB:       conf.Redis.EventsDB,
	})

	_, err = edb.Ping(context.Background()).Result()
	if err != nil {
		panic(fmt.Sprintf("Could not connect to Redis: %v", err))
	}

	mdb := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Address,
		Password: conf.Redis.Password,
		DB:       conf.Redis.MessagesDB,
	})

	_, err = mdb.Ping(context.Background()).Result()
	if err != nil {
		panic(fmt.Sprintf("Could not connect to Redis: %v", err))
	}

	sdb := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Address,
		Password: conf.Redis.Password,
		DB:       conf.Redis.StateDB,
	})

	_, err = sdb.Ping(context.Background()).Result()
	if err != nil {
		panic(fmt.Sprintf("Could not connect to Redis: %v", err))
	}

	c := &Cache{
		Rooms:    rdb,
		Events:   edb,
		Messages: mdb,
		State:    sdb,
	}

	return c, nil
}

func (c *App) CacheEvent(event_id id.EventID, event any) error {
	err := c.Cache.Events.Set(context.Background(), event_id.String(), event, 0).Err()
	if err != nil {
		c.Log.Error().Msgf("Couldn't cache room %v", err)
		return err
	}
	return nil
}

func (c *App) AddRoomToCache(room *RoomInfo) error {

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

	//go c.RebuildPublicRoomsCache()

	return nil
}

func (c *App) RemoveRoomFromCache(room_id id.RoomID) error {

	c.Log.Info().Msgf("Removing room from cache: %v", room_id)

	room, err := c.GetRoomInfo(&RoomInfoOptions{
		RoomID: room_id.String(),
	})
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

	go c.RebuildPublicRoomsCache()

	return nil
}

func (c *App) CachePublicRooms(public_rooms any) error {

	json, err := json.Marshal(public_rooms)
	if err != nil {
		c.Log.Error().Msgf("Couldn't marshal public rooms %v", err)
		return err
	}
	ttl := c.Config.Cache.PublicRooms.ExpireAfter
	if ttl == 0 {
		c.Log.Info().Msg("No TTL in config, using default value: 3600")
		ttl = 3600
	}

	expire := time.Duration(ttl) * time.Second

	err = c.Cache.Rooms.Set(context.Background(), "public_rooms", json, expire).Err()
	if err != nil {
		c.Log.Error().Msgf("Couldn't cache public rooms %v", err)
		return err
	}

	return nil
}

func (c *App) RebuildPublicRoomsCache() error {
	public_rooms, err := c.GetPublicRooms()
	if err != nil {
		return err
	}

	json, err := json.Marshal(public_rooms)
	if err != nil {
		c.Log.Error().Msgf("Couldn't marshal public rooms %v", err)
		return err
	}

	err = c.Cache.Rooms.Set(context.Background(), "public_rooms", json, 0).Err()
	if err != nil {
		c.Log.Error().Msgf("Couldn't cache public rooms %v", err)
		return err
	}

	return nil
}

func (c *App) UpdateRoomInfoCache(room_id string) error {
	info, err := c.GetRoomInfo(&RoomInfoOptions{
		RoomID: room_id,
	})

	if err != nil {
		return err
	}

	err = c.AddRoomToCache(info)
	if err != nil {
		c.Log.Error().Msgf("Error caching room info: %v", err)
		return err
	}

	return nil
}

func (c *App) CacheRoomMessages(room_id string) error {
	c.Log.Info().Msgf("Caching messages for room: %v", room_id)
	messages, err := c.Matrix.Messages(context.Background(), id.RoomID(room_id), "", "", 'b', nil, 100)
	if err != nil {
		c.Log.Error().Msgf("Error fetching messages: %v", err)
		return err
	}

	json, err := json.Marshal(messages)
	if err != nil {
		c.Log.Error().Msgf("Couldn't marshal messages %v", err)
		return err
	}

	err = c.Cache.Messages.Set(context.Background(), room_id, json, 60*time.Minute).Err()
	if err != nil {
		c.Log.Error().Msgf("Couldn't cache messages %v", err)
		return err
	}
	return nil
}
