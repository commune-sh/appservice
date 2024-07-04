package app

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func (c *App) MatrixAPIProxy() http.HandlerFunc {

	endpoint := fmt.Sprintf("%s/", c.Config.Matrix.Homeserver)
	target, _ := url.Parse(endpoint)

	proxy := httputil.NewSingleHostReverseProxy(target)

	return func(w http.ResponseWriter, r *http.Request) {

		log.Println("lol")

		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Config.AppService.AccessToken))
		w.Header().Del("Access-Control-Allow-Origin")

		proxy.ServeHTTP(w, r)
	}

}
