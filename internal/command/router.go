package command

import (
	"redisx/internal/storage"
	"strings"
)

type Handler func(store *storage.Storage, args []string) ([]byte, error)

type Router struct {
	h map[string]Handler
}

func NewRouter() *Router {
	return &Router{h: make(map[string]Handler)}
}

func (r *Router) Register(name string, h Handler) {
	r.h[strings.ToUpper(name)] = h
}

// Handle attempts to handle the command by name. Returns (resp, handled, err).
func (r *Router) Handle(name string, store *storage.Storage, args []string) ([]byte, bool, error) {
	h, ok := r.h[strings.ToUpper(name)]
	if !ok {
		return nil, false, nil
	}
	resp, err := h(store, args)
	return resp, true, err
}
