package req

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"

	"word-amongus-game/pkgs/auth"
	"word-amongus-game/pkgs/random"
)

var (
	CookiePlayerId http.Cookie = http.Cookie{
		Path:     "/",
		Name:     "player_id",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   1 << 31,
	}
	CookiePlayerName http.Cookie = http.Cookie{
		Path:     "/",
		Name:     "player_name",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   1 << 31,
	}
)

type PlayerCookies struct {
	Id   *http.Cookie
	Name *http.Cookie
}

type Player struct {
	Id   string
	Name string
}

func GetPlayerFromCookies(req *http.Request, key []byte) (Player, error) {
	playerIdCookie, err := req.Cookie(CookiePlayerId.Name)
	if err != nil {
		return Player{}, err
	}

	playerNameCookie, err := req.Cookie(CookiePlayerName.Name)
	if err != nil {
		return Player{}, err
	}

	playerIdSigned1, playerId, found := strings.Cut(playerIdCookie.Value, ":")
	if !found {
		return Player{}, fmt.Errorf("invalid player id cookie")
	}

	playerIdSigned2, err := auth.SignStringHex(playerId, key)
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

func NewPlayerCookies(playerName string, key []byte) (PlayerCookies, error) {
	playerId, err := random.RandomHex(random.LengthPlayerId)

	playerIdSigned, err := auth.SignStringHex(playerId, key)
	if err != nil {
		return PlayerCookies{}, err
	}

	playerIdCookie := CookiePlayerId
	playerIdCookie.Value = fmt.Sprintf("%s:%s", playerIdSigned, playerId)

	playerNameCookie := CookiePlayerName
	playerNameCookie.Value = playerName

	return PlayerCookies{
		Id:   &playerIdCookie,
		Name: &playerNameCookie,
	}, nil
}
