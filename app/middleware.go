package app

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/tidwall/buntdb"
	"maunium.net/go/mautrix/id"
)

// Get authenticated user's ID
func (c *App) AuthenticatedUser(r *http.Request) *string {
	user_id, ok := r.Context().Value("user_id").(string)

	if !ok {
		return nil
	}

	return &user_id
}

// Get authenticated user's access token
func (c *App) AuthenticatedAccessToken(r *http.Request) *string {
	access_token, ok := r.Context().Value("access_token").(string)

	if !ok {
		return nil
	}

	return &access_token

}

// This ensures that request is from Synapse Homeserver
func (c *App) AuthenticateHomeserver(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		access_token, err := ExtractAccessToken(r)

		if err != nil ||
			access_token == nil ||
			*access_token != c.Config.AppService.HSAccessToken {

			log.Println("error")

			RespondWithJSON(w, &JSONResponse{
				Code: http.StatusForbidden,
				JSON: map[string]any{
					"errcode": "BAD_ACCESS_TOKEN",
					"error":   "access token invalid",
				},
			})
			return
		}

		h.ServeHTTP(w, r)

	})
}

func ReplacePathParam(path, oldValue, newValue string) string {
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		if segment == oldValue {
			segments[i] = newValue
		}
	}
	return strings.Join(segments, "/")
}

// This checks whethere a given {room_id} is actually not a room ID but the
// localpart of an alias. If it is, it resolves the alias to a room ID.
func (c *App) ValidateRoomID(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		room_id := chi.URLParam(r, "room_id")

		// missing room param
		if room_id == "" {
			RespondWithError(w, &JSONResponse{
				Code: http.StatusOK,
				JSON: map[string]any{
					"error": "room ID is required",
				},
			})
			return
		}

		is_room_id := IsValidRoomID(room_id)

		if !is_room_id {

			alias := id.NewRoomAlias(room_id, c.Config.Matrix.ServerName)

			resp, err := c.Matrix.ResolveAlias(context.Background(), alias)
			if err != nil {
				c.Log.Error().Err(err).Msg("error resolving alias")
				RespondWithJSON(w, &JSONResponse{
					Code: http.StatusForbidden,
					JSON: map[string]any{
						"errcode": "M_NOT_FOUND",
						"error":   "Room not found.",
					},
				})
				return
			}

			if resp.RoomID.String() != "" {
				// pass on the resolved room ID to next handler
				rctx := chi.RouteContext(r.Context())
				rctx.URLParams.Add("room_id", resp.RoomID.String())

				// replace the alias with the resolved room ID
				r.URL.Path = ReplacePathParam(r.URL.Path, room_id, resp.RoomID.String())
			}
		}

		h.ServeHTTP(w, r)

	})
}
func (c *App) ValidatePublicRoom(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		room_id := chi.URLParam(r, "room_id")

		var joined bool
		c.Cache.JoinedRooms.View(func(tx *buntdb.Tx) error {
			_, err := tx.Get(room_id)
			joined = err == nil
			return nil
		})

		if joined {
			h.ServeHTTP(w, r)
			return
		}

		RespondWithJSON(w, &JSONResponse{
			Code: http.StatusForbidden,
			JSON: map[string]any{
				"errcode": "M_NOT_FOUND",
				"error":   "Room not found.",
			},
		})
		return

	})
}
