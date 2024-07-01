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
	c.Log.Info().Msgf("Join response: %v", join)

	// cache the room
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

				invited := event.Content.Raw["membership"].(string) == "invite"

				if invited {
					go c.JoinRoom(event.RoomID.String())
				}

			default:
			}

		}

		RespondWithJSON(w, &JSONResponse{
			Code: http.StatusOK,
		})

	}
}
