package main

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

func GuiLogger(l *zerolog.Logger, serverName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			trId := GenerateTRID(serverName)
			ctx := context.WithValue(r.Context(), TrID, trId)
			ctx = context.WithValue(ctx, Time, time.Now())

			myCtx := context.WithValue(context.Background(), TrID, trId)
			l.Info().Ctx(myCtx).Msg("Request started")

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
