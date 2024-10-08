package app

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
)

type CachingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (w *CachingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *CachingResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (c *App) MessagesProxy() http.HandlerFunc {

	endpoint := fmt.Sprintf("%s/", c.Config.Matrix.Homeserver)
	target, _ := url.Parse(endpoint)

	proxy := httputil.NewSingleHostReverseProxy(target)

	return func(w http.ResponseWriter, r *http.Request) {

		room_id := chi.URLParam(r, "room_id")
		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")

		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Config.AppService.AccessToken))
		w.Header().Del("Access-Control-Allow-Origin")

		// return cached messages if no query params
		if from == "" && to == "" && c.Config.Cache.Messages.Enabled {
			cached, err := c.Cache.Messages.Get(context.Background(), room_id).Result()
			if err == nil && cached != "" {
				c.Log.Info().Msgf("Found cached messages")
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Write([]byte(cached))
				return
			}
		}

		crw := &CachingResponseWriter{ResponseWriter: w}
		proxy.ServeHTTP(crw, r)

		if crw.statusCode == http.StatusOK {
			// cache messages
			if from == "" && to == "" && c.Config.Cache.Messages.Enabled {

				ttl := c.Config.Cache.Messages.ExpireAfter
				if ttl == 0 {
					c.Log.Info().Msg("No TTL in config, using default value: 3600")
					ttl = 3600
				}

				expire := time.Duration(ttl) * time.Second

				err := c.Cache.Messages.Set(context.Background(), room_id, crw.body.String(), expire).Err()
				if err != nil {
					c.Log.Error().Msgf("Couldn't cache messages %v", err)
				}
			}
		}
	}
}
