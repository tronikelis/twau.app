package req

import (
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"strings"

	"word-amongus-game/pkgs/auth"
	"word-amongus-game/pkgs/game_state"
	"word-amongus-game/pkgs/random"

	"github.com/tronikelis/maruchi"
)

type ReqContext struct {
	maruchi.ReqContext
	Rooms     game_state.Rooms
	SecretKey []byte
}

type Player struct {
	Id   string
	Name string
}

func (self ReqContext) SetPlayer(name string) error {
	playerId, err := random.RandomHex(random.LengthPlayerId)

	playerIdSigned, err := auth.SignStringHex(playerId, self.SecretKey)
	if err != nil {
		return err
	}

	playerIdCookie := CookiePlayerId
	playerIdCookie.Value = fmt.Sprintf("%s:%s", playerIdSigned, playerId)

	playerNameCookie := CookiePlayerName
	playerNameCookie.Value = name

	http.SetCookie(self.Writer(), &playerIdCookie)
	http.SetCookie(self.Writer(), &playerNameCookie)

	return nil
}

func (self ReqContext) Player() (Player, error) {
	playerIdCookie, err := self.Req().Cookie(CookiePlayerId.Name)
	if err != nil {
		return Player{}, err
	}

	playerNameCookie, err := self.Req().Cookie(CookiePlayerName.Name)
	if err != nil {
		return Player{}, err
	}

	playerIdSigned1, playerId, found := strings.Cut(playerIdCookie.Value, ":")
	if !found {
		return Player{}, fmt.Errorf("invalid player id cookie")
	}

	playerIdSigned2, err := auth.SignStringHex(playerId, self.SecretKey)
	if err != nil {
		return Player{}, err
	}

	if subtle.ConstantTimeCompare([]byte(playerIdSigned1), []byte(playerIdSigned2)) != 1 {
		return Player{}, fmt.Errorf("unauthorized player id cookie")
	}

	return Player{
		Id:   playerId,
		Name: playerNameCookie.Value,
	}, nil
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
