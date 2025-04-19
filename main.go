package main

import (
	_ "embed"
	"fmt"
	"net/http"

	"guiio/guiio_middleware"
	"guiio/guiio_util"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"github.com/sphynx/logger"
)

var (
	//go:embed banner.txt
	banner []byte
	Mlog   *zerolog.Logger
)

func main() {
	guiio_util.ServerInfo(banner)
	Mlog = logger.New()

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
	}))
	r.Use(guiio_middleware.NewLogger(Mlog, "guiio"))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})

	r.Route("/api/bucket", func(r chi.Router) {
		//todo trid middleware for all the root rotues before middlewaer
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello, World!")
		})
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello, World!")
		})
	})

	Mlog.Info().Msg("Start Server")
	if err := http.ListenAndServe(":8080", r); err != nil {
		Mlog.Panic().Err(err).Msg("Failed to start server")
	}
}
