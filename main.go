package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"word-amongus-game/pkgs/game_state"
	"word-amongus-game/pkgs/server/controllers"
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

	secretKeyStr := os.Getenv("SECRET_KEY")
	if secretKeyStr == "" {
		log.Fatal("empty SECRET_KEY")
	}

	server.Middleware(req.MiddlewareReqContext(game_state.NewRooms(), []byte(secretKeyStr)))

	controllers.Register(server)

	errChan := make(chan error)
	go func() {
		errChan <- http.ListenAndServe("localhost:3000", server.ServeMux())
	}()

	fmt.Println("listening on 3000")

	panic(<-errChan)
}
