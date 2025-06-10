package guiio_handler

import (
	"guiio/guiio_middleware"
	guiio_http "guiio/guiio_server/http"
	"guiio/guiio_service"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/sphynx/config"
)

type HttpHandler struct {
	server        *guiio_http.GuiioHttpServer
	bucketService guiio_service.BucketService
	log           *zerolog.Logger
}

func NewHttpHandler(conf *config.GConfig, log *zerolog.Logger) *HttpHandler {
	server := guiio_http.NewGuiioHttpServer(conf, log)
	bucketService := guiio_service.NewStorageService()

	return &HttpHandler{
		server:        server,
		bucketService: bucketService,
		log:           log,
	}
}

func (h *HttpHandler) Start() error {
	return h.server.Start()
}

func (h *HttpHandler) Init() {
	router := h.server.GetRouter()

	router.Use(guiio_middleware.HttpRequestLogger(h.log))

	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/", h.ListBucket)
		r.Post("/", h.CreateBucket)
		r.Delete("/{bucketName}", h.DeleteBucket)
	})
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
