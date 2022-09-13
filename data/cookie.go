package data

import (
	"net/http"
	"strconv"

	"github.com/gorilla/securecookie"
)

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func SetSession(u *User, w http.ResponseWriter) {
	value := map[string]string{
		"login": u.Email,
		"pass":  u.Password,
		"fname": u.Fname,
		"lname": u.Lname,
		"id":    strconv.Itoa(u.Id),
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(w, cookie)
	}
}

func GetUserName(r *http.Request) User {
	var res User
	if cookie, err := r.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			res.Email = cookieValue["login"]
			res.Fname = cookieValue["fname"]
			res.Lname = cookieValue["lname"]
			res.Id, _ = strconv.Atoi(cookieValue["id"])
		}
	}
	return res
}

func ClearSession(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}
