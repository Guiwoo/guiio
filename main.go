package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/sphynx/logger"
)

type key string

const (
	TRID key = "trid"
	Time key = "time"
)

func GuiLogger(l *zerolog.Logger, serverName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			trId := GenerateTRID(serverName)
			ctx := context.WithValue(r.Context(), "trid", trId)
			ctx = context.WithValue(ctx, Time, time.Now())

			myCtx := context.WithValue(context.Background(), "trid", trId)
			l.Info().Ctx(myCtx).Msg("Request started")

			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}

func GenerateTRID(serverName string) string {
	return fmt.Sprintf("%s-%d", serverName, time.Now().UnixNano())
}

func main() {

	mlog := logger.New()

	r := chi.NewRouter()
	r.Use(GuiLogger(mlog, "guiIo"))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})

	if err := http.ListenAndServe(":8080", r); err != nil {
		mlog.Panic().Err(err).Msg("Failed to start server")
	}
}
