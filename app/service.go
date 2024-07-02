package app

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/tidwall/buntdb"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func (c *App) JoinRoom(room_id string) {
	_, err := c.Matrix.JoinRoom(context.Background(), room_id, "", nil)
	if err != nil {
		c.Log.Error().Msgf("Error joining room: %v", err)
	}
}

func (c *App) AddRoomToCache(room_id string) {
	c.Log.Info().Msgf("Caching joined room: %v", room_id)
	c.Cache.JoinedRooms.Update(func(tx *buntdb.Tx) error {
		tx.Set(room_id, "true", nil)
		return nil
	})
}

func (c *App) LeaveRoom(room_id id.RoomID) {
	_, err := c.Matrix.LeaveRoom(context.Background(), room_id)
	if err != nil {
		c.Log.Error().Msgf("Error joining room: %v", err)
	}

	c.RemoveRoomFromCache(room_id)
}

func (c *App) RemoveRoomFromCache(room_id id.RoomID) {
	c.Log.Info().Msgf("Removing room from cache: %v", room_id)
	c.Cache.JoinedRooms.Update(func(tx *buntdb.Tx) error {
		tx.Delete(room_id.String())
		return nil
	})
}

func (c *App) Transactions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var events struct {
			Events []event.Event `json:"events"`
		}
		if err := json.Unmarshal(body, &events); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, event := range events.Events {

			switch event.Type.Type {
			case "m.room.history_visibility":
				state, ok := event.Content.Raw["history_visibility"].(string)

				if ok && state == "world_readable" &&
					c.Config.AppService.Rules.AutoJoin {
					c.JoinRoom(event.RoomID.String())
					c.AddRoomToCache(event.RoomID.String())
				}

				if ok && state != "world_readable" {
					c.LeaveRoom(event.RoomID)
				}

			case "m.room.member":

				state, ok := event.Content.Raw["membership"].(string)

				if ok {
					if state == "invite" {
						c.Log.Info().Msgf("Invited to room: %v", event.RoomID.String())
						c.JoinRoom(event.RoomID.String())
						c.AddRoomToCache(event.RoomID.String())
					}
					if state == "leave" || state == "ban" {
						c.RemoveRoomFromCache(event.RoomID)
					}
				}

			default:
			}

		}

		RespondWithJSON(w, &JSONResponse{
			Code: http.StatusOK,
		})

	}
}
