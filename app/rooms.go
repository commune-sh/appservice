package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type PublicRoom struct {
	RoomID            string   `json:"room_id"`
	Type              string   `json:"type,omitempty"`
	OriginServerTS    int64    `json:"origin_server_ts"`
	RoomType          string   `json:"room_type,omitempty"`
	Name              string   `json:"name,omitempty"`
	CanonicalAlias    string   `json:"canonical_alias"`
	CommuneAlias      string   `json:"commune_alias,omitempty"`
	Sender            string   `json:"sender,omitempty"`
	AvatarURL         string   `json:"avatar_url,omitempty"`
	BannerURL         string   `json:"banner_url,omitempty"`
	Topic             string   `json:"topic,omitempty"`
	JoinRule          string   `json:"join_rule"`
	HistoryVisibility string   `json:"history_visibility"`
	Children          []string `json:"children,omitempty"`
	Parents           []string `json:"parents,omitempty"`
	Settings          any      `json:"settings,omitempty"`
	RoomCategories    any      `json:"room_categories,omitempty"`
	IsBridge          bool     `json:"is_bridge"`
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

		if c.Config.Cache.PublicRooms.Enabled {

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

		}

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

		if c.Config.Cache.PublicRooms.Enabled {
			go func() {
				c.Log.Info().Msgf("Caching public rooms")
				err := c.CachePublicRooms(public_rooms)
				if err != nil {
					c.Log.Error().Msgf("Couldn't marshal public rooms %v", err)
				}
			}()
		}

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

		bridge := CouldBeBridge(room.State)
		r.IsBridge = bridge

		child_state := room.State[event.NewEventType("m.space.child")]
		if child_state != nil {
			for child, ev := range child_state {

				// don't list child spaces with empty content
				via := ev.Content.Raw["via"]
				if via == nil {
					continue
				}

				for _, ro := range rooms {
					join_rule_event := ro.State[event.NewEventType("m.room.join_rules")][""]
					if join_rule_event != nil {
						join_rule, ok := join_rule_event.Content.Raw["join_rule"].(string)
						if ok {
							if join_rule != "public" {
								//continue
							}
							if ro.RoomID.String() == child {
								r.Children = append(r.Children, child)
							}
						}
					}
				}

			}
		}

		room_type_event := room.State[event.NewEventType("m.room.create")][""]
		if room_type_event != nil {
			room_type, ok := room_type_event.Content.Raw["type"].(string)
			if ok {
				r.Type = room_type
			}
			r.Sender = room_type_event.Sender.String()
			r.OriginServerTS = room_type_event.Timestamp
		}

		name_event := room.State[event.NewEventType("m.room.name")][""]
		if name_event != nil {
			name, ok := name_event.Content.Raw["name"].(string)
			if ok {
				if strings.Contains(name, "[⛓️]") {
					continue
				}
				r.Name = name
			}
		}

		ev := event.Type{"commune.room.name", 2}
		name_event = room.State[ev][""]
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
		ev = event.Type{"m.room.history_visibility", 2}
		hv_event := room.State[ev][""]
		if hv_event != nil {
			hv, ok := hv_event.Content.Raw["history_visibility"].(string)
			if ok {
				r.HistoryVisibility = hv
			}
		}

		// hacky way to get history visibility
		ev = event.Type{"commune.room.banner", 2}
		banner_event := room.State[ev][""]
		if banner_event != nil {
			banner, ok := banner_event.Content.Raw["url"].(string)
			if ok {
				r.BannerURL = banner
			}
		}

		ev = event.Type{"commune.room.type", 2}
		rt_event := room.State[ev][""]
		if rt_event != nil {
			rtv, ok := rt_event.Content.Raw["type"].(string)
			if ok {
				r.RoomType = rtv
			}
		}

		/*
				ev = event.Type{"commune.room.alias", 2}
				ca_event := room.State[ev][""]
				if ca_event != nil {
					calias, ok := ca_event.Content.Raw["alias"].(string)
					if ok {
						r.CommuneAlias = calias
					}
				}

					ev = event.Type{"commune.room.settings", 2}
					settings_event := room.State[ev][""]
					if settings_event != nil {
						r.Settings = settings_event.Content.Raw["settings"]
					}

			ev = event.Type{"commune.room.categories", 2}
			rc_event := room.State[ev][""]
			if rc_event != nil {
				r.RoomCategories = rc_event.Content.Raw["categories"]
			}

		*/

		join_rule_event := room.State[event.NewEventType("m.room.join_rules")][""]
		if join_rule_event != nil {
			join_rule, ok := join_rule_event.Content.Raw["join_rule"].(string)
			if ok {

				/*
					if join_rule != "public" {
						continue
					}
				*/

				r.JoinRule = join_rule
			}
		}

		processed = append(processed, r)

	}
	return processed, nil
}

type RoomInfo struct {
	RoomID         string `json:"room_id"`
	Name           string `json:"name,omitempty"`
	CanonicalAlias string `json:"canonical_alias"`
	AvatarURL      string `json:"avatar_url,omitempty"`
	BannerURL      string `json:"banner_url,omitempty"`
	Topic          string `json:"topic,omitempty"`
}

type RoomInfoOptions struct {
	RoomID      string
	ChildRoomID string
	EventID     string
}

func (c *App) GetRoomInfo(r *RoomInfoOptions) (*RoomInfo, error) {

	state, err := c.Matrix.State(context.Background(), id.RoomID(r.RoomID))

	if err != nil {
		return nil, err
	}

	room := RoomInfo{
		RoomID: r.RoomID,
	}

	name_event := state[event.NewEventType("m.room.name")][""]
	if name_event != nil {
		name, ok := name_event.Content.Raw["name"].(string)
		if ok {
			room.Name = name
		}
	}

	ev := event.Type{"commune.room.name", 2}
	name_event = state[ev][""]
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

	ev = event.Type{"commune.room.banner", 2}
	banner_event := state[ev][""]
	if banner_event != nil {
		banner, ok := banner_event.Content.Raw["url"].(string)
		if ok {
			room.BannerURL = banner
		}
	}

	return &room, nil
}

func (c *App) RoomInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		room_id := chi.URLParam(r, "room_id")

		child_room := r.URL.Query().Get("room")

		event_id := r.URL.Query().Get("event")

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
		o := &RoomInfoOptions{
			RoomID: room_id,
		}

		info, err := c.GetRoomInfo(o)

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

		resp := map[string]any{
			"info": info,
		}

		room := RoomInfo{}

		if child_room != "" {

			hierarchy, err := c.Matrix.Hierarchy(context.Background(), id.RoomID(room_id), nil)

			if err != nil {
				c.Log.Error().Msgf("Error fetching room hierarchy: %v", err)
			}

			if hierarchy != nil {
				//c.Log.Info().Msgf("Room hierarchy: %v", hierarchy)

				for _, child := range hierarchy.Rooms {
					slug := Slugify(child.Name)
					if slug == child_room || child.RoomID.String() == child_room {

						room = RoomInfo{
							RoomID: child.RoomID.String(),
							Name:   child.Name,
							Topic:  child.Topic,
						}

						if child.AvatarURL.FileID != "" {
							room.AvatarURL = fmt.Sprintf("mxc://%s/%s", child.AvatarURL.Homeserver, child.AvatarURL.FileID)
						}

						resp["room"] = room
					}
				}
			}
		}

		if event_id != "" && room.RoomID != "" {

			event, err := c.Matrix.GetEvent(context.Background(), id.RoomID(room.RoomID), id.EventID(event_id))
			if err != nil {
				c.Log.Error().Msgf("Error fetching event: %v", err)
			}
			if event != nil {
				c.Log.Info().Msgf("Event: %v", event)
				resp["event"] = event
			}
			if event != nil {
				profile, err := c.Matrix.GetProfile(context.Background(), event.Sender)
				if err != nil {
					c.Log.Error().Msgf("Error fetching profile: %v", err)
				}
				if profile != nil {
					resp["sender"] = profile
				}
			}
		}

		RespondWithJSON(w, &JSONResponse{
			Code: http.StatusOK,
			JSON: resp,
		})
	}
}
