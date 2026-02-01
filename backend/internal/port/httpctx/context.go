package httpctx

import (
	"context"
	"io"
	"net/http"
)

// Context는 HTTP 요청 컨텍스트를 추상화한 인터페이스입니다.
type Context interface {
	// JSON은 JSON 응답을 전송합니다.
	JSON(code int, v interface{}) error

	// Bind는 요청 본문을 구조체에 바인딩합니다.
	Bind(v interface{}) error

	// Param은 URL 파라미터를 반환합니다.
	Param(name string) string

	// Query는 쿼리 파라미터를 반환합니다.
	Query(name string) string

	// GetHeader는 요청 헤더를 반환합니다.
	GetHeader(name string) string

	// SetHeader는 응답 헤더를 설정합니다.
	SetHeader(name, value string)

	// Context는 표준 context.Context를 반환합니다.
	Context() context.Context

	// Request는 원본 *http.Request를 반환합니다.
	Request() *http.Request

	// Stream은 스트림 응답을 전송합니다.
	Stream(code int, contentType string, r io.Reader) error
}
