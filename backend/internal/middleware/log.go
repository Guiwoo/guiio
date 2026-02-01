package middleware

import (
	"context"
	"guiio/backend/internal/util"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

func HttrRequestLogger(log *zerolog.Logger, name string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			now := time.Now()
			ctx := context.WithValue(r.Context(), util.Time, now)

			trID := util.GenerateTRID(name, now)
			ctx = context.WithValue(ctx, util.TrID, trID)

			reqLogger := log.With().Str("trid", trID).Time("req_time", now).Logger()
			ctx = reqLogger.WithContext(ctx)

			reqLogger.Info().Ctx(ctx).Msgf("New Request: %s", r.URL.Path)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
