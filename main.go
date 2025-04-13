package main

import (
	"fmt"
	"net/http"

	"word-amongus-game/pkgs/controllers"
	"word-amongus-game/pkgs/server/req"

	"github.com/tronikelis/maruchi"
)

func main() {
	server := maruchi.NewServer()

	server.Group("").
		Middleware(func(ctx maruchi.ReqContext, next maruchi.Handler) {
			// if production {
			// 	ctx.Header().Set("cache-control", "public, max-age=31536000")
			// } else {
			ctx.Writer().Header().Set("cache-control", "no-cache, no-store, must-revalidate")
			// }
			next(ctx)
		}).
		Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	reqContext := req.NewReqContext()

	server.Middleware(func(ctx maruchi.ReqContext, next func(ctx maruchi.ReqContext)) {
		ctxBase := ctx.(maruchi.ReqContextBase)
		req.InitContext(&ctxBase, reqContext)

		next(ctxBase)
	})

	controllers.Register(server)

	errChan := make(chan error)
	go func() {
		errChan <- http.ListenAndServe("localhost:3000", server.ServeMux())
	}()

	fmt.Println("listening on 3000")

	panic(<-errChan)
}
