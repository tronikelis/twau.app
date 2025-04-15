package req

import "net/http"

var (
	CookiePlayerId http.Cookie = http.Cookie{
		Path:     "/",
		Name:     "player_id",
		SameSite: http.SameSiteLaxMode,
		HttpOnly: false,
		MaxAge:   1 << 31,
	}

	CookiePlayerName http.Cookie = http.Cookie{
		Path:     "/",
		Name:     "player_name",
		SameSite: http.SameSiteLaxMode,
		HttpOnly: false,
		MaxAge:   1 << 31,
	}
)
