package req

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"

	"twau.app/pkgs/auth"
	"twau.app/pkgs/game_state"
	"twau.app/pkgs/random"

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

func (self ReqContext) SetCookie(cookie *http.Cookie) {
	cookie.Value = base64.StdEncoding.EncodeToString([]byte(cookie.Value))
	http.SetCookie(self.Writer(), cookie)
}

func (self ReqContext) Cookie(name string) (*http.Cookie, error) {
	cookie, err := self.Req().Cookie(name)
	if err != nil {
		return nil, err
	}

	value, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil, err
	}

	cookie.Value = string(value)
	return cookie, nil
}

func (self ReqContext) SetPlayer(name string) error {
	playerId := random.RandomB64(16)

	playerIdSigned, err := auth.SignStringB64(playerId, self.SecretKey)
	if err != nil {
		return err
	}

	playerIdCookie := CookiePlayerId
	playerIdCookie.Value = fmt.Sprintf("%s:%s", playerIdSigned, playerId)

	playerNameCookie := CookiePlayerName
	playerNameCookie.Value = name

	self.SetCookie(&playerIdCookie)
	self.SetCookie(&playerNameCookie)

	return nil
}

func (self ReqContext) ClearPlayer() {
	playerNameCookie := CookiePlayerName
	playerNameCookie.MaxAge = -1

	playerIdCookie := CookiePlayerId
	playerIdCookie.MaxAge = -1

	self.SetCookie(&playerNameCookie)
	self.SetCookie(&playerIdCookie)
}

func (self ReqContext) Player() (Player, error) {
	playerIdCookie, err := self.Cookie(CookiePlayerId.Name)
	if err != nil {
		return Player{}, err
	}

	playerNameCookie, err := self.Cookie(CookiePlayerName.Name)
	if err != nil {
		return Player{}, err
	}

	playerIdSigned1, playerId, found := strings.Cut(playerIdCookie.Value, ":")
	if !found {
		return Player{}, fmt.Errorf("invalid player id cookie")
	}

	playerIdSigned2, err := auth.SignStringB64(playerId, self.SecretKey)
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
