package req

import (
	"net/http"
)

var (
	CookiePlayerId = http.Cookie{
		Path:     "/",
		Name:     "player_id",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   1 << 31,
	}
	CookiePlayerName = http.Cookie{
		Path:     "/",
		Name:     "player_name",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   1 << 31,
	}
	CookieRoomPassword = http.Cookie{
		Path:     "/",
		Name:     "room_password",
		SameSite: http.SameSiteLaxMode,
	}
)
