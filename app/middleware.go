package app

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
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

func (c *App) ValidatePublicRoom(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		room_id := chi.URLParam(r, "room_id")
		log.Println(room_id)

		rooms, err := c.Matrix.JoinedRooms(context.Background())
		if err != nil {
			c.Log.Error().Msgf("Error fetching joined rooms: %v", err)
		}

		if len(rooms.JoinedRooms) > 0 {
			for _, room := range rooms.JoinedRooms {
				if room.String() == room_id {
					h.ServeHTTP(w, r)
					return
				}
			}
		}

		RespondWithJSON(w, &JSONResponse{
			Code: http.StatusForbidden,
			JSON: map[string]any{
				"errcode": "M_NOT_IN_ROOM",
				"error":   "User not in room",
			},
		})
		return

	})
}
