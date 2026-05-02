package middleware

import (
	"github.com/gmcorenet/framework/pkg"
)

type Middleware func(next pkg.HandlerFunc) pkg.HandlerFunc

func Logger(next pkg.HandlerFunc) pkg.HandlerFunc {
	return func(req *pkg.Request) *pkg.Response {
		return next(req)
	}
}

func Auth(next pkg.HandlerFunc) pkg.HandlerFunc {
	return func(req *pkg.Request) *pkg.Response {
		return next(req)
	}
}