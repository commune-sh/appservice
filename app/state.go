package app

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
)

func (c *App) StateProxy() http.HandlerFunc {

	endpoint := fmt.Sprintf("%s/", c.Config.Matrix.Homeserver)
	target, _ := url.Parse(endpoint)

	proxy := httputil.NewSingleHostReverseProxy(target)

	return func(w http.ResponseWriter, r *http.Request) {

		room_id := chi.URLParam(r, "room_id")
		/*
			from := r.URL.Query().Get("from")
			to := r.URL.Query().Get("to")
		*/

		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Config.AppService.AccessToken))
		w.Header().Del("Access-Control-Allow-Origin")

		cached, err := c.Cache.State.Get(context.Background(), room_id).Result()
		if err == nil && cached != "" {
			c.Log.Info().Msgf("Found cached state for %v", room_id)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Write([]byte(cached))
			return
		}

		crw := &CachingResponseWriter{ResponseWriter: w}
		proxy.ServeHTTP(crw, r)

		if crw.statusCode == http.StatusOK {
			// cache state
			err := c.Cache.State.Set(context.Background(), room_id, crw.body.String(), 60*time.Minute).Err()
			if err != nil {
				c.Log.Error().Msgf("Couldn't cache state %v", err)
			} else {
				c.Log.Info().Msgf("Cached state for room %v", room_id)
			}
		}
	}
}
