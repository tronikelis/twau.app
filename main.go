package main

import (
	"net/http"

	"word-amongus-game/pkgs/controllers"
	"word-amongus-game/pkgs/server/req"

	"github.com/tronikelis/maruchi"
)

func main() {
	server := maruchi.NewServer()

	reqContext := req.NewReqContext()

	server.Middleware(func(ctx maruchi.ReqContext, next func(ctx maruchi.ReqContext)) {
		ctxBase := ctx.(maruchi.ReqContextBase)
		req.InitContext(&ctxBase, reqContext)

		next(ctxBase)
	})

	controllers.Register(server)

	if err := http.ListenAndServe("localhost:3000", server.ServeMux()); err != nil {
		panic(err)
	}
}
