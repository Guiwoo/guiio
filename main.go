package main

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/sphynx/logger"
)

//go:embed banner.txt
var banner []byte

func main() {
	fmt.Println(string(banner))
	mlog := logger.New()

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
	}))
	r.Use(GuiLogger(mlog, "guiIo"))
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

	mlog.Info().Msg("Start Server")
	if err := http.ListenAndServe(":8080", r); err != nil {
		mlog.Panic().Err(err).Msg("Failed to start server")
	}
}
