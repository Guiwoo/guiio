package guiio_handler

import (
	guiio_http "guiio/guiio_server"
	"guiio/guiio_service"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/sphynx/config"
)

type HttpHandler struct {
	server        *guiio_http.GuiioHttpServer
	bucketService BucketService
}

func NewHttpHandler(conf *config.GConfig, log *zerolog.Logger) *HttpHandler {
	server := guiio_http.NewGuiioHttpServer(conf, log)
	bucketService := guiio_service.NewStorageService()

	return &HttpHandler{
		server:        server,
		bucketService: bucketService,
	}
}

func (h *HttpHandler) Start() error {
	return h.server.Start()
}

func (h *HttpHandler) Init() {
	router := h.server.GetRouter()

	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/", h.ListBucket)
		r.Post("/", h.CreateBucket)
		r.Delete("/{bucketName}", h.DeleteBucket)
	})
}

func (h *HttpHandler) ListBucket(w http.ResponseWriter, r *http.Request) {
	h.bucketService.ListBucket(w, r)
}

func (h *HttpHandler) CreateBucket(w http.ResponseWriter, r *http.Request) {
	h.bucketService.CreateBucket(w, r)
}

func (h *HttpHandler) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	h.bucketService.DeleteBucket(w, r)
}
