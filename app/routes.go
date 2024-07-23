package app

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/hostrouter"
	"github.com/unrolled/secure"
)

func (c *App) Routes() {
	compressor := middleware.NewCompressor(5, "text/html", "text/css", "text/event-stream")
	compressor.SetEncoder("nop", func(w io.Writer, _ int) io.Writer {
		return w
	})

	// c.Router.Use(middleware.ThrottleBacklog(10, 50, time.Second*10))
	c.Router.Use(middleware.RequestID)
	c.Router.Use(middleware.RealIP)
	c.Router.Use(middleware.Logger)
	c.Router.Use(middleware.StripSlashes)
	c.Router.Use(compressor.Handler)

	c.CORS()

	hr := hostrouter.New()

	hr.Map(c.Config.App.Domain, routes(c))

	c.Router.Mount("/", hr)
}

func routes(c *App) chi.Router {
	sop := secure.Options{
		ContentSecurityPolicy: "script-src 'self' 'unsafe-eval' 'unsafe-inline' $NONCE",
		IsDevelopment:         false,
		AllowedHosts: []string{
			c.Config.App.Domain,
			"http://localhost:8080",
		},
	}

	secureMiddleware := secure.New(sop)

	r := chi.NewRouter()

	r.Route("/robots.txt", func(r chi.Router) {
		r.Get("/", c.RobotsTXT())
	})

	r.Route("/_matrix/app/v1", func(r chi.Router) {
		r.Use(c.AuthenticateHomeserver)
		r.Post("/ping", c.RespondToPing())
		r.Put("/transactions/{txnId}", c.Transactions())
	})

	r.Route("/_matrix/client/v3/rooms/{room_id}", func(r chi.Router) {
		r.Use(c.ValidateRoomID)
		r.Use(c.ValidatePublicRoom)
		r.Get("/info", c.RoomInfo())
		r.Get("/aliases", c.MatrixAPIProxy())
		r.Get("/event/*", c.MatrixAPIProxy())
		r.Get("/state", c.MatrixAPIProxy())
		r.Get("/state/*", c.MatrixAPIProxy())
		r.Get("/joined_members", c.MatrixAPIProxy())
		r.Get("/members", c.MatrixAPIProxy())
		r.Get("/timestamp_to_event", c.MatrixAPIProxy())
		r.Route("/messages", func(r chi.Router) {
			r.Get("/", c.MessagesProxy())
		})
	})

	r.Route("/_matrix/client/v1/rooms/{room_id}", func(r chi.Router) {
		r.Use(c.ValidateRoomID)
		r.Use(c.ValidatePublicRoom)
		r.Get("/hierarchy", c.MatrixAPIProxy())
		r.Get("/threads", c.MatrixAPIProxy())
		r.Get("/relations/*", c.MatrixAPIProxy())
	})

	r.Route("/publicRooms", func(r chi.Router) {
		r.Get("/", c.PublicRooms())
	})

	r.Route("/health", func(r chi.Router) {
		r.Get("/", c.Health())
	})

	r.Route("/", func(r chi.Router) {
		r.Use(secureMiddleware.Handler)
		//r.Get("/*", c.Index())
	})

	compressor := middleware.NewCompressor(5, "text/html", "text/css")
	compressor.SetEncoder("nop", func(w io.Writer, _ int) io.Writer {
		return w
	})
	r.NotFound(c.NotFound)

	return r
}

func (c *App) NotFound(w http.ResponseWriter, r *http.Request) {

	RespondWithError(w, &JSONResponse{
		Code: http.StatusNotFound,
		JSON: map[string]any{
			"errcode": "M_UNRECOGNIZED",
			"message": "Unrecognized request",
		},
	})
}

func (c *App) CORS() {
	cors := cors.New(cors.Options{
		AllowedOrigins:   c.Config.Security.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"X-PINGOTHER", "Accept", "Authorization", "Image", "Attachment", "File-Type", "Content-Type", "X-CSRF-Token", "Access-Control-Allow-Origin"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	c.Router.Use(cors.Handler)
}
