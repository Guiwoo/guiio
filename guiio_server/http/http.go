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
	log.Info().Msgf("HttpServer Conf %+v", conf)
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

func (s *GuiioHttpServer) GetRouter() *chi.Mux {
	return s.router
}

func (s *GuiioHttpServer) Start() error {
	port := config.Get[int]("port")
	s.log.Info().Msgf("Server Starts %d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), s.router)
}
