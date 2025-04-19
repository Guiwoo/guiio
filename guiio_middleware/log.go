package guiio_middleware

import (
	"context"
	"guiio/guiio_util"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

func NewLogger(l *zerolog.Logger, serverName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			trId := guiio_util.GenerateTRID(serverName)
			ctx := context.WithValue(r.Context(), guiio_util.TrID, trId)
			ctx = context.WithValue(ctx, guiio_util.Time, time.Now())

			myCtx := context.WithValue(context.Background(), guiio_util.TrID, trId)
			l.Info().Ctx(myCtx).Msg("Request started")

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
