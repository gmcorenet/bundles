package handler

import "github.com/gmcorenet/framework/pkg"

type BaseHandler struct{}

func (h *BaseHandler) OK(data interface{}) *pkg.Response {
	return pkg.NewResponse(toJSON(data), 200).WithHeader("Content-Type", "application/json")
}

func (h *BaseHandler) Created(data interface{}) *pkg.Response {
	return pkg.NewResponse(toJSON(data), 201).WithHeader("Content-Type", "application/json")
}

func (h *BaseHandler) NotFound(message string) *pkg.Response {
	return pkg.NewResponse(toJSON(map[string]string{"error": message}), 404).WithHeader("Content-Type", "application/json")
}

func (h *BaseHandler) Error(message string) *pkg.Response {
	return pkg.NewResponse(toJSON(map[string]string{"error": message}), 500).WithHeader("Content-Type", "application/json")
}

func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}