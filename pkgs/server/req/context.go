package req

import (
	"log"
	"net/http"

	"word-amongus-game/pkgs/game_state"

	"github.com/tronikelis/maruchi"
)

type ReqContext struct {
	maruchi.ReqContext
	Rooms     game_state.Rooms
	SecretKey []byte
}

func MiddlewareReqContext(rooms game_state.Rooms, secretKey []byte) maruchi.Middleware {
	return func(ctx maruchi.ReqContext, next maruchi.Handler) {
		next(ReqContext{
			ReqContext: ctx,
			Rooms:      rooms,
			SecretKey:  secretKey,
		})
	}
}

func WithReqContext(handler func(ctx ReqContext) error) maruchi.Handler {
	return func(ctx maruchi.ReqContext) {
		if err := handler(ctx.(ReqContext)); err != nil {
			log.Println("WithReqContext_Err", "err", err)

			ctx.Writer().Header().Set("content-type", "text/plain")

			ctx.Writer().WriteHeader(http.StatusInternalServerError)
			if _, err := ctx.Writer().Write([]byte(err.Error())); err != nil {
				log.Println("WithReqContext_WriteErr", "err", err)
			}
		}
	}
}
