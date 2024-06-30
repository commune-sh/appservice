package app

import (
	"log"
	"net/http"
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
