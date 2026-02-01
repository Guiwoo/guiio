package util

import (
	"context"

	"github.com/rs/zerolog"
)

// LoggerFromContext는 zerolog 로거를 컨텍스트에서 추출합니다.
// 없으면 fallback을 반환합니다.
func LoggerFromContext(ctx context.Context, fallback *zerolog.Logger) *zerolog.Logger {
	if ctx == nil {
		return fallback
	}
	if l := zerolog.Ctx(ctx); l != nil {
		return l
	}
	return fallback
}
