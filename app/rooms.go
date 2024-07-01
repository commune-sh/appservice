package app

import (
	"context"
	"net/http"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
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

type PublicRooms struct {
	RoomID id.RoomID
	State  mautrix.RoomStateMap
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

		parents := []*PublicRooms{}

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

				parent := PublicRooms{
					RoomID: room_id,
					State:  state,
				}

				parents = append(parents, &parent)
			}

		}

		resp := map[string]any{}

		if len(parents) > 0 {
			rms, err := ProcessPublicRooms(parents)
			if err != nil {
				c.Log.Error().Msgf("Error processing public rooms: %v", err)
			}
			c.Log.Info().Msgf("Public rooms: %v", rms)
			resp["chunk"] = rms
			resp["total_room_count_estimate"] = len(parents)
		}

		//c.Log.Info().Msgf("Public rooms: %v", parents)

		RespondWithJSON(w, &JSONResponse{
			Code: http.StatusOK,
			JSON: resp,
		})
	}
}

func ProcessPublicRooms(rooms []*PublicRooms) ([]PublicRoom, error) {
	processed := []PublicRoom{}
	for _, room := range rooms {

		r := PublicRoom{
			RoomID: room.RoomID.String(),
		}

		name_event := room.State[event.NewEventType("m.room.name")][""]
		if name_event != nil {
			name, ok := name_event.Content.Raw["name"].(string)
			if ok {
				r.Name = name
			}
		}

		alias_event := room.State[event.NewEventType("m.room.canonical_alias")][""]
		if alias_event != nil {
			alias, ok := alias_event.Content.Raw["alias"].(string)
			if ok {
				r.CanonicalAlias = alias
			}
		}

		avatar_event := room.State[event.NewEventType("m.room.avatar")][""]
		if avatar_event != nil {
			avatar, ok := avatar_event.Content.Raw["url"].(string)
			if ok {
				r.AvatarURL = avatar
			}
		}

		processed = append(processed, r)

	}
	return processed, nil
}
