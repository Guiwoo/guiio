package httpctx

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ChiContext struct {
	ctx context.Context
	w   http.ResponseWriter
	r   *http.Request
}

func NewChiContext(w http.ResponseWriter, r *http.Request) Context {
	return &ChiContext{ctx: r.Context(), w: w, r: r}
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

func (c *ChiContext) Context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return c.r.Context()
}

func (c *ChiContext) Request() *http.Request {
	return c.r
}

func (c *ChiContext) Stream(code int, contentType string, r io.Reader) error {
	if contentType != "" {
		c.w.Header().Set("Content-Type", contentType)
	}
	c.w.WriteHeader(code)
	_, err := io.Copy(c.w, r)
	return err
}
