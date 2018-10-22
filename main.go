package main

import (
	"context"
	"net/http"
	"time"

	"github.com/aflag/rabisco-server/rabisco"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func init() {
	formatter := &logrus.TextFormatter{
		FullTimestamp: true,
	}
	logrus.SetFormatter(formatter)
}

func main() {
	backend, err := rabisco.NewBackend(
		context.Background(),
		logrus.New(),
		"mongodb://127.0.0.1:27017",
	)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	r := mux.NewRouter()

	// Handlers
	r.HandleFunc("/rooms", makeCreateRoomHandler(backend)).Methods(http.MethodPost)
	r.HandleFunc("/rooms/{id}", makeGetRoomHandler(backend)).Methods(http.MethodGet)
	r.HandleFunc("/rooms/{id}/join", makeJoinRoomHandler(backend)).Methods(http.MethodPost)
	r.HandleFunc("/rooms/{id}/start", makeStartHandler(backend)).Methods(http.MethodPost)
	r.HandleFunc("/rooms/{id}/round/next", makeNextScoreHandler(backend)).Methods(http.MethodPost)
	r.HandleFunc("/play", makePlayHandler(backend)).Methods(http.MethodPost)
	r.HandleFunc("/me", makeMeHandler(backend)).Methods(http.MethodGet)
	r.HandleFunc("/login", makeLoginHandler(backend)).Methods(http.MethodPost)

	// Middlewares
	r.Use(loggerMiddleware)
	// CORS is a bit funny, it doesn't work with r.Use.
	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://90.194.89.3:8080"}),
		handlers.AllowedHeaders([]string{"Accept", "Accept-Language", "Content-Language", "Origin", "content-type"}),
		handlers.ExposedHeaders([]string{"Location", "Set-Cookie"}),
		handlers.AllowCredentials(),
	)

	srv := &http.Server{
		Handler:      cors(r),
		Addr:         "0.0.0.0:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logrus.Fatal(srv.ListenAndServe())
}
