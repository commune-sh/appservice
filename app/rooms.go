package app

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type PublicRoom struct {
	RoomID            string   `json:"room_id"`
	Name              string   `json:"name"`
	CanonicalAlias    string   `json:"canonical_alias"`
	AvatarURL         string   `json:"avatar_url"`
	Topic             string   `json:"topic"`
	JoinRule          string   `json:"join_rule"`
	HistoryVisibility string   `json:"history_visibility"`
	Children          []string `json:"children,omitempty"`
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

		/*
			cached, err := c.Cache.Rooms.Get(context.Background(), "public_rooms").Result()

			if err == nil && cached != "" {
				c.Log.Info().Msgf("Found cached public rooms")

				var data map[string]interface{}

				if err := json.Unmarshal([]byte(cached), &data); err == nil {
					RespondWithJSON(w, &JSONResponse{
						Code: http.StatusOK,
						JSON: data,
					})
					return
				}

			}
		*/

		public_rooms, err := c.GetPublicRooms()
		if err != nil {
			RespondWithJSON(w, &JSONResponse{
				Code: http.StatusOK,
				JSON: map[string]any{
					"errcode": "M_UNKNOWN",
					"error":   "Error fetching public rooms",
				},
			})
			return
		}

		/*
			go func() {
				c.Log.Info().Msgf("Caching public rooms")
				err := c.CachePublicRooms(public_rooms)
				if err != nil {
					c.Log.Error().Msgf("Couldn't marshal public rooms %v", err)
				}
			}()
		*/

		RespondWithJSON(w, &JSONResponse{
			Code: http.StatusOK,
			JSON: map[string]any{
				"rooms": public_rooms,
			},
		})
	}
}

func (c *App) GetPublicRooms() (any, error) {

	rooms, err := c.Matrix.JoinedRooms(context.Background())
	if err != nil {
		return nil, err
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
				//continue
			}

			parent := PublicRooms{
				RoomID: room_id,
				State:  state,
			}

			parents = append(parents, &parent)
		}

	}

	if len(parents) > 0 {

		rms, err := ProcessPublicRooms(parents)
		if err != nil {
			c.Log.Error().Msgf("Error processing public rooms: %v", err)
		}
		return rms, nil
	}

	return nil, nil
}

func ProcessPublicRooms(rooms []*PublicRooms) ([]PublicRoom, error) {
	processed := []PublicRoom{}
	for _, room := range rooms {

		r := PublicRoom{
			RoomID: room.RoomID.String(),
		}

		child_state := room.State[event.NewEventType("m.space.child")]
		if child_state != nil {
			for child, _ := range child_state {
				for _, ro := range rooms {
					join_rule_event := ro.State[event.NewEventType("m.room.join_rules")][""]
					if join_rule_event != nil {
						join_rule, ok := join_rule_event.Content.Raw["join_rule"].(string)
						if ok {
							if join_rule != "public" {
								continue
							}
							r.JoinRule = join_rule
						}
					}
					if ro.RoomID.String() == child {
						r.Children = append(r.Children, child)
					}
				}

			}
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

		topic_event := room.State[event.NewEventType("m.room.topic")][""]
		if topic_event != nil {
			topic, ok := topic_event.Content.Raw["topic"].(string)
			if ok {
				r.Topic = topic
			}
		}

		// hacky way to get history visibility
		var ev = event.Type{"m.room.history_visibility", 2}
		hv_event := room.State[ev][""]
		if hv_event != nil {
			hv, ok := hv_event.Content.Raw["history_visibility"].(string)
			if ok {
				r.HistoryVisibility = hv
			}
		}

		join_rule_event := room.State[event.NewEventType("m.room.join_rules")][""]
		if join_rule_event != nil {
			join_rule, ok := join_rule_event.Content.Raw["join_rule"].(string)
			if ok {

				if join_rule != "public" {
					continue
				}

				r.JoinRule = join_rule
			}
		}

		processed = append(processed, r)

	}
	return processed, nil
}

func (c *App) GetRoomInfo(room_id string) (*PublicRoom, error) {

	state, err := c.Matrix.State(context.Background(), id.RoomID(room_id))

	if err != nil {
		return nil, err
	}

	room := PublicRoom{
		RoomID: room_id,
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

	topic_event := state[event.NewEventType("m.room.topic")][""]
	if topic_event != nil {
		topic, ok := topic_event.Content.Raw["topic"].(string)
		if ok {
			room.Topic = topic
		}
	}

	return &room, nil
}

func (c *App) RoomInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		room_id := chi.URLParam(r, "room_id")

		/*
			// check if it's cached
			cached, err := c.Cache.Rooms.Get(context.Background(), room_id).Result()

			if err == nil && cached != "" {
				c.Log.Info().Msgf("Found cached room info for %v", room_id)
				var room PublicRoom
				if err := json.Unmarshal([]byte(cached), &room); err == nil {
					RespondWithJSON(w, &JSONResponse{
						Code: http.StatusOK,
						JSON: room,
					})
					return
				}

			}
		*/

		info, err := c.GetRoomInfo(room_id)

		if err != nil {
			RespondWithError(w, &JSONResponse{
				Code: http.StatusOK,
				JSON: map[string]any{
					"error": "Error fetching room state",
				},
			})
			return
		}

		/*
			go func() {
				err := c.AddRoomToCache(info)
				if err != nil {
					c.Log.Error().Msgf("Error caching room info: %v", err)
				}
			}()
		*/

		RespondWithJSON(w, &JSONResponse{
			Code: http.StatusOK,
			JSON: info,
		})
	}
}
