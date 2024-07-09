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

func (c *App) JoinRoom(room_id id.RoomID) error {
	_, err := c.Matrix.JoinRoom(context.Background(), room_id.String(), "", nil)
	if err != nil {
		c.Log.Error().Msgf("Error joining room: %v", err)
		return err
	}
	return nil
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

func (c *App) ProcessRoom(room_id id.RoomID) error {

	state, err := c.Matrix.State(context.Background(), room_id)
	if err != nil {
		c.Log.Error().Msgf("Error fetching state: %v", err)
		return err
	}

	has_children := event.NewEventType("m.space.child")
	has_parent := event.NewEventType("m.space.parent")

	is_parent_space := len(state[has_children]) > 0
	is_child_space := len(state[has_parent]) > 0

	if !is_parent_space || is_child_space {
	}

	err = c.JoinRoom(room_id)
	if err != nil {
		return err
	}
	err = c.AddRoomToCache(room_id)
	if err != nil {
		return err
	}

	return nil
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
					c.Log.Info().Msgf("Autojoining room: %v", event.RoomID.String())
					err = c.ProcessRoom(event.RoomID)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}

				if ok && state != "world_readable" {
					c.LeaveRoom(event.RoomID)
				}

			case "m.room.member":

				state, ok := event.Content.Raw["membership"].(string)

				if ok {
					if state == "invite" {
						c.Log.Info().Msgf("Invited to room: %v", event.RoomID.String())
						err = c.ProcessRoom(event.RoomID)
						if err != nil {
							http.Error(w, err.Error(), http.StatusInternalServerError)
							return
						}
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
