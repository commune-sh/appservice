package app

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

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

func (c *App) LeaveRoom(room_id id.RoomID) error {
	_, err := c.Matrix.LeaveRoom(context.Background(), room_id)
	if err != nil {
		c.Log.Error().Msgf("Error leaving room: %v", err)
		return err
	}

	err = c.RemoveRoomFromCache(room_id)
	if err != nil {
		return err
	}

	return nil
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
		//return nil
	}

	c.Log.Info().Msgf("Joining room: %v", room_id.String())
	err = c.JoinRoom(room_id)
	if err != nil {
		return err
	}

	info, err := c.GetRoomInfo(room_id.String())
	if err != nil {
		return err
	}

	err = c.AddRoomToCache(info)
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

			go func() {
				json, err := json.Marshal(event)
				if err != nil {
					c.Log.Error().Msgf("Couldn't marshal event %v", err)
					return
				}
				err = c.CacheEvent(event.ID, json)
				if err != nil {
					c.Log.Error().Msgf("Error caching event: %v", err)
				}
			}()

			switch event.Type.Type {
			case "m.room.message":
				/*
					err := c.CacheRoomMessages(event.RoomID.String())
					if err != nil {
						c.Log.Error().Msgf("Error caching messages: %v", err)
						//http.Error(w, err.Error(), http.StatusInternalServerError)
						//return
					}
				*/
			case "m.room.history_visibility":
				state, ok := event.Content.Raw["history_visibility"].(string)

				if ok && state == "world_readable" &&
					c.Config.AppService.Rules.AutoJoin {
					err = c.ProcessRoom(event.RoomID)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}

				if ok && state != "world_readable" {
					err = c.LeaveRoom(event.RoomID)
					if err != nil {
						c.Log.Error().Msgf("Error leaving room: %v", err)
					}
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
						err := c.RemoveRoomFromCache(event.RoomID)
						if err != nil {
							http.Error(w, err.Error(), http.StatusInternalServerError)
							return
						}
					}
				}

			case "m.room.name":
				name, ok := event.Content.Raw["name"].(string)
				c.Log.Info().Msgf("New room name, updating cache value: %v", name)
				if ok {
					err := c.UpdateRoomInfoCache(event.RoomID.String())
					if err != nil {
						RespondWithError(w, &JSONResponse{
							Code: http.StatusOK,
							JSON: map[string]any{
								"error": "Error updating room info cache",
							},
						})
						return
					}
				}
			case "m.room.avatar":
				url, ok := event.Content.Raw["url"].(string)
				c.Log.Info().Msgf("New room avatar, updating cache value: %v", url)
				if ok {
					err := c.UpdateRoomInfoCache(event.RoomID.String())
					if err != nil {
						RespondWithError(w, &JSONResponse{
							Code: http.StatusOK,
							JSON: map[string]any{
								"error": "Error updating room info cache",
							},
						})
						return
					}
				}
			case "m.room.topic":
				topic, ok := event.Content.Raw["topic"].(string)
				c.Log.Info().Msgf("New room topic, updating cache value: %v", topic)
				if ok {
					err := c.UpdateRoomInfoCache(event.RoomID.String())
					if err != nil {
						RespondWithError(w, &JSONResponse{
							Code: http.StatusOK,
							JSON: map[string]any{
								"error": "Error updating room info cache",
							},
						})
						return
					}
				}
			case "m.room.canonical_alias":
				alias, ok := event.Content.Raw["alias"].(string)
				c.Log.Info().Msgf("New room alias, updating cache value: %v", alias)
				if ok {
					err := c.UpdateRoomInfoCache(event.RoomID.String())
					if err != nil {
						RespondWithError(w, &JSONResponse{
							Code: http.StatusOK,
							JSON: map[string]any{
								"error": "Error updating room info cache",
							},
						})
						return
					}
					go c.RebuildPublicRoomsCache()
				}
			default:
			}

		}

		RespondWithJSON(w, &JSONResponse{
			Code: http.StatusOK,
		})

	}
}
