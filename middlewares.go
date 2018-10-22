package main

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"strconv"
)

type loggerKey struct{}

func setAccessControlMiddleware(value string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", value)
			next.ServeHTTP(w, req)
		})
	}
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = strconv.Itoa(rand.Intn(10000))
		}
		logger := logrus.WithFields(logrus.Fields{
			"reqId": reqID,
		})
		logger.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
			"host":   r.Host,
		}).Info("Request begins")
		ctx := context.WithValue(r.Context(), loggerKey{}, logger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
