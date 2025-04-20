package guiio_http

import (
	"fmt"
	"guiio/guiio_middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"

	"github.com/sphynx/config"
)

type GuiioHttpServer struct {
	router *chi.Mux
	log    *zerolog.Logger
}

func NewGuiioHttpServer(conf *config.GConfig, log *zerolog.Logger) *GuiioHttpServer {
	server := &GuiioHttpServer{
		router: chi.NewRouter(),
		log:    log,
	}

	server.Init()

	return server
}

func (s *GuiioHttpServer) Init() {
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
	}))

	s.router.Use(guiio_middleware.NewLogger(s.log, config.Get[string]("server_name")))
}

func (s *GuiioHttpServer) Start() error {
	return http.ListenAndServe(fmt.Sprintf(":%d", config.Get[int]("port")), s.router)
}
