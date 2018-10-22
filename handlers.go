package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/aflag/rabisco-server/rabisco"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func isPlayerInRoom(ctx context.Context, logger logrus.FieldLogger, backend rabisco.Backend, playerID, roomID string) bool {
	room, err := backend.GetRoom(ctx, logger, roomID, playerID)
	if err != nil {
		logger.WithField("error", err).Error("GetRoom failed")
		return false
	}
	for _, player := range room.Players {
		if player.ID == playerID {
			return true
		}
	}
	return false
}

func makeGetRoomHandler(backend rabisco.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value(loggerKey{}).(logrus.FieldLogger)
		vars := mux.Vars(r)
		roomID := vars["id"]
		playerID := getPlayerID(r)
		if playerID == "" {
			logger.Info("unauthorized")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		room, err := backend.GetRoom(r.Context(), logger, roomID, playerID)
		if err != nil {
			if err == rabisco.ErrNotFound {
				logger.Info("Round not found")
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			} else {
				logger.WithField("error", err.Error()).Error("Unexpected error")
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		payload, err := json.Marshal(room)
		if err != nil {
			logger.WithField("error", err.Error()).Error("Unexpected error")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := w.Write(payload); err != nil {
			logger.WithField("error", err.Error()).Error("Unexpected error")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func makeCreateRoomHandler(backend rabisco.Backend) http.HandlerFunc {
	type inputT struct {
		Name string `json:"name"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value(loggerKey{}).(logrus.FieldLogger)

		playerID := getPlayerID(r)
		if playerID == "" {
			logger.Info("unauthorized")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.WithField("error", err).Error("Read error")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var input inputT
		if err := json.Unmarshal(payload, &input); err != nil {
			logger.WithField("error", err.Error()).Error("Invalid json")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		roomID := generateRandomString(6)
		if err := backend.CreateRoom(r.Context(), logger, roomID, input.Name); err != nil {
			logger.WithField("error", err.Error()).Error("Room creation failure")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// auto join created room
		if err := backend.JoinRoom(r.Context(), logger, playerID, roomID); err != nil {
			logger.WithField("error", err.Error()).Error("Room created, but autojoin failed")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Location", "/rooms/"+roomID)
		w.WriteHeader(http.StatusCreated)
	}
}

func makeJoinRoomHandler(backend rabisco.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value(loggerKey{}).(logrus.FieldLogger)

		vars := mux.Vars(r)
		roomID := vars["id"]

		playerID := getPlayerID(r)
		logger = logger.WithFields(logrus.Fields{
			"roomId":   roomID,
			"playerId": playerID,
		})
		if playerID == "" {
			logger.Info("unauthorized")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		switch err := backend.JoinRoom(r.Context(), logger, playerID, roomID); err {
		case rabisco.ErrNotFound:
			logger.Info("Room not found")
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		case nil:
			logger.Info("Joined room")
		default:
			logger.WithField("error", err.Error()).Error("Failed to join")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func makeMeHandler(backend rabisco.Backend) http.HandlerFunc {
	type outputT struct {
		PlayerID string `json:"playerId"`
		Name     string `json:"name"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value(loggerKey{}).(logrus.FieldLogger)
		playerID := getPlayerID(r)
		if playerID == "" {
			logger.Info("unauthorized")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		logger = logger.WithField("playerId", playerID)
		payload, err := json.Marshal(outputT{PlayerID: playerID, Name: playerID})
		if err != nil {
			logger.WithField("error", err.Error()).Error("Unable to marshal")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(payload); err != nil {
			logger.WithField("error", err.Error()).Error("Unable to marshal")
			return
		}
		logger.Info("Found myself")
	}
}

func makeLoginHandler(backend rabisco.Backend) http.HandlerFunc {
	type inputT struct {
		Name string `json:"name"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value(loggerKey{}).(logrus.FieldLogger)
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.WithField("error", err).Error("Read error")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logger = logger.WithField("payload", string(payload))
		var input inputT
		if err := json.Unmarshal(payload, &input); err != nil {
			logger.WithField("error", err).Info("Bad request")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(input.Name) == "" {
			logger.Info("Blank name")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		setCookie(w, input.Name)
		logger.Info("Logged in")
	}
}

func makeStartHandler(backend rabisco.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value(loggerKey{}).(logrus.FieldLogger)
		vars := mux.Vars(r)
		roomID := vars["id"]
		playerID := getPlayerID(r)
		logger = logger.WithFields(logrus.Fields{
			"playerId": playerID,
			"roomId":   roomID,
		})

		if !isPlayerInRoom(r.Context(), logger, backend, playerID, roomID) {
			logger.Warn("Player tried to start a game they didn't belong")
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		if err := backend.Start(r.Context(), logger, roomID); err != nil {
			logger.WithField("error", err).Error("Error 500")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func makePlayHandler(backend rabisco.Backend) http.HandlerFunc {
	type inputT struct {
		RoomID string        `json:"roomId"`
		Round  rabisco.Round `json:"round"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value(loggerKey{}).(logrus.FieldLogger)
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.WithField("error", err).Error("Error 500")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		playerID := getPlayerID(r)

		var input inputT
		if err := json.Unmarshal(payload, &input); err != nil {
			logger.WithField("error", err).Info("Bad json")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		logger = logger.WithFields(logrus.Fields{
			"roomId":   input.RoomID,
			"playerId": playerID,
			"round":    input.Round.String(),
		})

		if err := backend.Play(r.Context(), logger, input.RoomID, playerID, &input.Round, nil); err != nil {
			if err == rabisco.ErrNotFound {
				logger.Info("Cannot play in that room")
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			} else {
				logger.WithField("error", err).Error("Error 500")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}
	}
}

func makeNextScoreHandler(backend rabisco.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value(loggerKey{}).(logrus.FieldLogger)

		vars := mux.Vars(r)
		roomID := vars["id"]

		playerID := getPlayerID(r)
		logger = logger.WithFields(logrus.Fields{
			"roomId":   roomID,
			"playerId": playerID,
		})
		if playerID == "" {
			logger.Info("unauthorized")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if !isPlayerInRoom(r.Context(), logger, backend, playerID, roomID) {
			logger.Warn("Player tried to go next in a game they didn't belong")
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		switch err := backend.NextScore(r.Context(), logger, roomID); err {
		case rabisco.ErrNotFound:
			logger.Info("Room not found")
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		case nil:
			logger.Info("Next score")
		default:
			logger.WithField("error", err.Error()).Error("Failed to join")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
