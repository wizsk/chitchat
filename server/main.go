package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	cdb "chithat/db"

	"golang.org/x/net/websocket"
	_ "golang.org/x/net/websocket"
)

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (wr *wrappedWriter) WriteHeader(statusCode int) {
	wr.ResponseWriter.WriteHeader(statusCode)
	wr.statusCode = statusCode
}

func (wr *wrappedWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := wr.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("not a hijacker?")
	}
	return h.Hijack()
}

func ensureSignedin(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v, err := r.Cookie(cookieName)
		if err != nil {
			http.Redirect(w, r, "/sinin", http.StatusMovedPermanently)
			return
		}
		user, ok := cookie[v.Value]
		if !ok {
			http.Redirect(w, r, "/sinin", http.StatusMovedPermanently)
			return
		}

		nr := r.WithContext(context.WithValue(r.Context(), "user", user))
		h.ServeHTTP(w, nr)
	})
}

func logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wr := wrappedWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		h.ServeHTTP(w, r)
		log.Println(wr.statusCode, r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	handler := http.NewServeMux()

	handler.HandleFunc("POST /singin", singIn)
	handler.HandleFunc("POST /singup", singUp)
	handler.Handle("/ws", ensureSignedin(websocket.Handler(ws)))

	addr := ":8001"
	server := http.Server{
		Addr:    addr,
		Handler: logger(handler),
	}

	// defer db.Close()
	// defer fmt.Println("closed")

	log.Println("listening http://localhost" + addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("ListenAndServe: " + err.Error())
	}

}

var (
	// peralal safe korte hobe
	cookie     = map[string]cdb.User{}
	cookieName = "__chitchat_coookie"
	db         = mustDo(cdb.InitDB("user=postgres password=pass host=localhost port=5432 sslmode=disable"))
)

func getCleanedFormValue(r *http.Request, v string) string {
	return strings.TrimSpace(r.FormValue(v))
}

func singUp(w http.ResponseWriter, r *http.Request) {
	name := getCleanedFormValue(r, "name")
	userName := getCleanedFormValue(r, "user_name")
	password := getCleanedFormValue(r, "password")

	if name == "" || userName == "" || password == "" {
		http.Error(w, "bad input", http.StatusBadRequest)
		return
	}

	user, err := db.Singup(name, userName, password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !writeCookie(w, user) {
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(d)
}

func singIn(w http.ResponseWriter, r *http.Request) {
	userName := getCleanedFormValue(r, "user_name")
	password := getCleanedFormValue(r, "password")

	if userName == "" || password == "" {
		http.Error(w, "bad input", http.StatusBadRequest)
		return
	}

	user, err := db.Singin(userName, password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !writeCookie(w, user) {
		return
	}

	w.WriteHeader(http.StatusFound)
	w.Write(d)
}

func writeCookie(w http.ResponseWriter, user cdb.User) bool {
	buff := [32]byte{}
	if _, err := rand.Read(buff[:]); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	cv := base64.StdEncoding.EncodeToString(buff[:])
	cookie[cv] = user
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: cv})
	return true
}

func mustDo[T any](t T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func ws(conn *websocket.Conn) {
}
