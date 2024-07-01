package app

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/tidwall/buntdb"
	"maunium.net/go/mautrix/event"
)

func (c *App) JoinRoom(room_id string) {
	join, err := c.Matrix.JoinRoom(context.Background(), room_id, "", nil)
	if err != nil {
		c.Log.Error().Msgf("Error joining room: %v", err)
	}

	// cache the room
	c.Log.Info().Msgf("Caching joined room: %v", join.RoomID.String())
	c.Cache.JoinedRooms.Update(func(tx *buntdb.Tx) error {
		tx.Set(room_id, "true", nil)
		return nil
	})
}

func (c *App) Transactions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		body, err := ioutil.ReadAll(r.Body)
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
				world_readable := event.Content.Raw["history_visibility"].(string) == "world_readable"

				if world_readable && c.Config.AppService.Rules.AutoJoin {
					go c.JoinRoom(event.RoomID.String())
				}

			case "m.room.member":

				state, ok := event.Content.Raw["membership"].(string)

				if ok {
					if state == "invite" {
						c.Log.Info().Msgf("Invited to room: %v", event.RoomID.String())
						go c.JoinRoom(event.RoomID.String())
					}
					if state == "leave" || state == "ban" {
						c.Log.Info().Msgf("Removing room from cache: %v", event.RoomID.String())
						c.Cache.JoinedRooms.Update(func(tx *buntdb.Tx) error {
							tx.Delete(event.RoomID.String())
							return nil
						})
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
