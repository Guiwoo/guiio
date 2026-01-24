package guiio_middleware

import (
	"context"
	"guiio/guiio_util"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

func HttrRequestLogger(log *zerolog.Logger, name string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			now := time.Now()
			ctx := context.WithValue(r.Context(), guiio_util.Time, now)

			trID := guiio_util.GenerateTRID(name, now)
			ctx = context.WithValue(ctx, guiio_util.TrID, trID)

			log.Info().Ctx(ctx).Msgf("New Requset: %s", r.URL.Path)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
