package guiio_handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// ChiContext는 Chi 프레임워크를 위한 Context 구현체입니다.
type ChiContext struct {
	w http.ResponseWriter
	r *http.Request
}

// NewChiContext는 새로운 ChiContext를 생성합니다.
func NewChiContext(w http.ResponseWriter, r *http.Request) Context {
	return &ChiContext{w: w, r: r}
}

func (c *ChiContext) JSON(code int, v interface{}) error {
	c.w.Header().Set("Content-Type", "application/json")
	c.w.WriteHeader(code)
	return json.NewEncoder(c.w).Encode(v)
}

func (c *ChiContext) Bind(v interface{}) error {
	return json.NewDecoder(c.r.Body).Decode(v)
}

func (c *ChiContext) Param(name string) string {
	return chi.URLParam(c.r, name)
}

func (c *ChiContext) Query(name string) string {
	return c.r.URL.Query().Get(name)
}

func (c *ChiContext) GetHeader(name string) string {
	return c.r.Header.Get(name)
}

func (c *ChiContext) SetHeader(name, value string) {
	c.w.Header().Set(name, value)
}
