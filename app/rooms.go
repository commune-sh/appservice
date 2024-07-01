package app

import (
	"context"
	"net/http"

	"maunium.net/go/mautrix/event"
)

type PublicRoom struct {
	RoomID         string `json:"room_id"`
	Name           string `json:"name"`
	CanonicalAlias string `json:"canonical_alias"`
	AvatarURL      string `json:"avatar_url"`
}

type Rooms struct {
	rooms []PublicRoom `json:"rooms"`
}

func (c *App) PublicRooms() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		rooms, err := c.Matrix.JoinedRooms(context.Background())
		if err != nil {
			RespondWithJSON(w, &JSONResponse{
				Code: http.StatusOK,
				JSON: map[string]any{
					"errcode": "M_UNKNOWN",
					"error":   "Error fetching rooms",
				},
			})
		}

		pr := []PublicRoom{}

		if len(rooms.JoinedRooms) > 0 {

			for _, room_id := range rooms.JoinedRooms {

				state, err := c.Matrix.State(context.Background(), room_id)
				if err != nil {
					c.Log.Error().Msgf("Error fetching state: %v", err)
				}

				has_children := event.NewEventType("m.space.child")
				has_parent := event.NewEventType("m.space.parent")

				is_parent_space := len(state[has_children]) > 0
				is_child_space := len(state[has_parent]) > 0

				if !is_parent_space || is_child_space {
					continue
				}

				room := PublicRoom{
					RoomID: room_id.String(),
				}

				name_event := state[event.NewEventType("m.room.name")][""]
				if name_event != nil {
					name, ok := name_event.Content.Raw["name"].(string)
					if ok {
						room.Name = name
					}
				}

				alias_event := state[event.NewEventType("m.room.canonical_alias")][""]
				if alias_event != nil {
					alias, ok := alias_event.Content.Raw["alias"].(string)
					if ok {
						room.CanonicalAlias = alias
					}
				}

				avatar_event := state[event.NewEventType("m.room.avatar")][""]
				if avatar_event != nil {
					avatar, ok := avatar_event.Content.Raw["url"].(string)
					if ok {
						room.AvatarURL = avatar
					}
				}

				pr = append(pr, room)

			}
		}

		RespondWithJSON(w, &JSONResponse{
			Code: http.StatusOK,
			JSON: map[string]any{
				"rooms": pr,
			},
		})
	}
}
