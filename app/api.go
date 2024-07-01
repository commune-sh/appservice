package app

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/rs/zerolog"
)

func (c *App) MatrixAPIProxy() http.HandlerFunc {

	endpoint := fmt.Sprintf("%s/", c.Config.Matrix.Homeserver)
	target, _ := url.Parse(endpoint)
	log.Println(target)

	proxy := httputil.NewSingleHostReverseProxy(target)

	return func(w http.ResponseWriter, r *http.Request) {
		c.Log.Info().Msg("Setting up Matrix API Proxy")

		c.Log.Info().
			Dict("details", zerolog.Dict().
				Str("api", fmt.Sprintf("%s %s", r.Method, r.URL.Path)),
			).Msg("Accessing Matrix API")

		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Config.AppService.AccessToken))
		w.Header().Del("Access-Control-Allow-Origin")

		proxy.ServeHTTP(w, r)
	}

}
