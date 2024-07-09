package app

import (
	"context"
	"net/http"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

func (c *App) Setup() {
	// build a cache of rooms that appservice has joined
	rooms, err := c.Matrix.JoinedRooms(context.Background())
	if err != nil {
		c.Log.Error().Msgf("Error fetching joined rooms: %v", err)
		return
	}

	if len(rooms.JoinedRooms) > 0 {
		for _, room_id := range rooms.JoinedRooms {

			info, err := c.GetRoomInfo(room_id.String())
			if err != nil {
				c.Log.Error().Msgf("Error fetching room info: %v", err)
			}

			if info != nil {
				c.AddRoomToCache(info)
				if err != nil {
					c.Log.Error().Msgf("Error adding room to cache: %v", err)
				}
			}

		}
	}

	c.Log.Info().Msg("Rebuilding public rooms cache")
	c.RebuildPublicRoomsCache()
}

func (c *App) JoinPublicRooms() {

	rooms, err := c.Matrix.JoinedRooms(context.Background())
	if err != nil {
		c.Log.Error().Msgf("Error fetching joined rooms: %v", err)
		return
	}
	if len(rooms.JoinedRooms) == 0 {
		return
	}

	// look up all public rooms and join them.

	var publicRooms struct {
		Chunk                  []event.Event `json:"chunk"`
		TotalRoomCountEstimate int           `json:"total_room_count_estimate"`
	}

	type filter struct {
		RoomTypes []string `json:"room_types"`
	}

	type jsr struct {
		Filter filter `json:"filter"`
		Limit  int    `json:"limit"`
	}

	var req jsr = jsr{
		Filter: filter{
			RoomTypes: []string{"m.space"},
		},
		Limit: 1000,
	}

	_, err = c.Matrix.MakeFullRequest(context.Background(), mautrix.FullRequest{
		Method:       http.MethodPost,
		URL:          c.Matrix.BuildClientURL("v3", "publicRooms"),
		RequestJSON:  &req,
		ResponseJSON: &publicRooms,
	})

	if err != nil {
		c.Log.Error().Msgf("Error fetching public rooms: %v", err)
		return
	}

	c.Log.Info().Msgf("Total public rooms: %v", len(publicRooms.Chunk))

	if len(publicRooms.Chunk) > 0 {
		for _, room := range publicRooms.Chunk {
			already_joined := Contains(rooms.JoinedRooms, room.RoomID)

			if already_joined {
				c.Log.Info().Msgf("Already joined: %v", room.RoomID)
				continue
			}
			c.Log.Info().Msgf("Joining: %v", room.RoomID)
			err = c.JoinRoom(room.RoomID)
			if err != nil {
				c.Log.Error().Msgf("Couldn't join room: %v", err)
				return
			}
			time.Sleep(1 * time.Second)
		}
	}
}
