package guiio_handler

import (
	"fmt"
	"guiio/guiio_middleware"
	"guiio/guiio_service"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/sphynx/config"

	guiio_http "guiio/guiio_server/http"
)

type HttpHandler struct {
	conf          *config.GConfig
	log           *zerolog.Logger
	bucketService guiio_service.BucketService
}

func NewHttpHandler(conf *config.GConfig, log *zerolog.Logger) *HttpHandler {
	//Todo 밖으로 빼기
	bucketService := guiio_service.NewStorageService()

	return &HttpHandler{
		conf:          conf,
		log:           log,
		bucketService: bucketService,
	}
}

func (h *HttpHandler) Start() error {
	port := config.Get[int]("port")
	serverName := config.Get[string]("server_name")

	h.log.Info().Msgf("Server Starts %d", port)

	router := chi.NewRouter()

	router.Use(guiio_middleware.HttrRequestLogger(h.log, serverName))

	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/", h.ListBucket)
		r.Post("/", h.CreateBucket)
		r.Delete("/{bucketName}", h.DeleteBucket)
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", port), router)
}

func (h *HttpHandler) ListBucket(w http.ResponseWriter, r *http.Request) {
	ctx := guiio_http.NewChiContext(w, r)
	h.bucketService.ListBucket(ctx)
}

func (h *HttpHandler) CreateBucket(w http.ResponseWriter, r *http.Request) {
	ctx := guiio_http.NewChiContext(w, r)

	h.bucketService.CreateBucket(ctx)
}

func (h *HttpHandler) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	ctx := guiio_http.NewChiContext(w, r)
	h.bucketService.DeleteBucket(ctx)
}
