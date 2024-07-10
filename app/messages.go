package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
)

type cachingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (w *cachingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *cachingResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (c *App) MessagesProxy() http.HandlerFunc {

	endpoint := fmt.Sprintf("%s/", c.Config.Matrix.Homeserver)
	target, _ := url.Parse(endpoint)

	proxy := httputil.NewSingleHostReverseProxy(target)

	return func(w http.ResponseWriter, r *http.Request) {

		room_id := chi.URLParam(r, "room_id")

		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Config.AppService.AccessToken))
		w.Header().Del("Access-Control-Allow-Origin")

		cached, err := c.Cache.Messages.Get(context.Background(), room_id).Result()
		if err == nil && cached != "" {
			c.Log.Info().Msgf("Found cached messages")
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(cached), &data); err == nil {
				RespondWithJSON(w, &JSONResponse{
					Code: http.StatusOK,
					JSON: data,
				})
				w.Write([]byte(cached))

				return
			}
		}

		crw := &cachingResponseWriter{ResponseWriter: w}
		proxy.ServeHTTP(crw, r)

		if crw.statusCode == http.StatusOK {

			err := c.Cache.Messages.Set(context.Background(), room_id, crw.body.String(), 5*time.Minute).Err()
			if err != nil {
				c.Log.Error().Msgf("Couldn't cache messages %v", err)
			}

		}
	}
}
