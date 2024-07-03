package app

import (
	"context"
	"net/http"
)

func (c *App) Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		rsp := map[string]any{
			"healthy": false,
		}

		_, err := c.Matrix.Whoami(context.Background())

		if err != nil {

			rsp["error"] = err.Error()

			RespondWithJSON(w, &JSONResponse{
				Code: http.StatusOK,
				JSON: rsp,
			})
			return
		}

		rsp["healthy"] = true

		RespondWithJSON(w, &JSONResponse{
			Code: http.StatusOK,
			JSON: rsp,
		})
	}
}
