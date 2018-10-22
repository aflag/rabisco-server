package main

import (
	"net/http"
	"strings"
)

func getPlayerID(r *http.Request) string {
	// TODO: encrypt this
	if playerID, err := r.Cookie("playerId"); err != nil {
		return ""
	} else {
		return strings.TrimSpace(playerID.Value)
	}
}

func setCookie(w http.ResponseWriter, playerID string) {
	http.SetCookie(w, &http.Cookie{Name: "playerId", Value: playerID})
}
