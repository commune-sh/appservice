package app

import (
	"context"

	"github.com/tidwall/buntdb"
)

func (c *App) Setup() {
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
