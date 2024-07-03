package app

import (
	"context"
	"net/http"
	"time"

	"github.com/tidwall/buntdb"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

func (c *App) Setup() {

	// build a cahce of rooms that appservice has joined

	rooms, err := c.Matrix.JoinedRooms(context.Background())
	if err != nil {
		c.Log.Error().Msgf("Error fetching joined rooms: %v", err)
		return
	}
	if len(rooms.JoinedRooms) > 0 {
		c.Cache.JoinedRooms.Update(func(tx *buntdb.Tx) error {
			tx.DeleteAll()
			for _, room_id := range rooms.JoinedRooms {
				//c.Log.Info().Msgf("room_id: %v", room_id)
				tx.Set(room_id.String(), "true", nil)
			}
			return nil
		})
	}

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
			c.JoinRoom(room.RoomID)
			time.Sleep(1 * time.Second)
		}
	}
}
