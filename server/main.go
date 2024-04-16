package main

import (
	"log"
	"net/http"
	"time"

	_ "golang.org/x/net/websocket"
)

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wr := wrappedWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		h.ServeHTTP(&wr, r)

		log.Println(wr.statusCode, r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {

	x := http.NewServeMux()
	logger(wrappedWriter)
	x.Handle("/", logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hwll"))
	})))

	y := http.NewServeMux()
	y.Handle("/bar/", http.StripPrefix("/bar", x))

	z := http.NewServeMux()
	z.Handle("/fo/", http.StripPrefix("/fo", x))

	err := http.ListenAndServe(":8001", z)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
