package main

import (
	"context"
	"flag"
	"net/http"
	"os"
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
	// clear command line (someone very naughty is adding all these extra
	// test.* flags)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	mongoURL := flag.String("mongo-url", "mongodb://127.0.0.1:27017", "mongo server url")
	allowOrigin := flag.String("allow-origin", "http://localhost:8000", "allow connections from this origin")
	bindAddr := flag.String("bind-to", "localhost:8000", "bind to this ip and port")

	flag.Parse()

	backend, err := rabisco.NewBackend(
		context.Background(),
		logrus.New(),
		*mongoURL,
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
		handlers.AllowedOrigins([]string{*allowOrigin}),
		handlers.AllowedHeaders([]string{"Accept", "Accept-Language", "Content-Language", "Origin", "content-type"}),
		handlers.ExposedHeaders([]string{"Location", "Set-Cookie"}),
		handlers.AllowCredentials(),
	)

	srv := &http.Server{
		Handler:      cors(r),
		Addr:         *bindAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logrus.Fatal(srv.ListenAndServe())
}
