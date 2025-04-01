package main

import (
	"fmt"
	"net/http"

	"word-amongus-game/pkgs/controllers"
	"word-amongus-game/pkgs/game_state"
	"word-amongus-game/pkgs/server/req"

	"github.com/gorilla/websocket"
	"github.com/tronikelis/maruchi"
)

func main() {
	wsUpgrader := websocket.Upgrader{}

	server := maruchi.NewServer()

	states := game_state.NewStates()

	server.Middleware(func(ctx maruchi.ReqContext, next func(ctx maruchi.ReqContext)) {
		ctxBase := ctx.(maruchi.ReqContextBase)
		req.InitContext(&ctxBase, states)

		next(ctxBase)
	})

	server.Route("", "/ws", func(ctx maruchi.ReqContext) {
		ws, err := wsUpgrader.Upgrade(ctx.Writer(), ctx.Req(), nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer ws.Close()

		for {
			mt, message, err := ws.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return
			}

			if err := ws.WriteMessage(mt, message); err != nil {
				fmt.Println(err)
			}
		}
	})

	controllers.Register(server)

	if err := http.ListenAndServe("localhost:3000", server.ServeMux()); err != nil {
		panic(err)
	}
}
