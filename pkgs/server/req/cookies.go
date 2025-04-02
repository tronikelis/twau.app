package req

import "net/http"

var CookiePlayerId http.Cookie = http.Cookie{
	Name:     "player_id",
	SameSite: http.SameSiteLaxMode,
	HttpOnly: false,
	MaxAge:   1 << 31,
}
