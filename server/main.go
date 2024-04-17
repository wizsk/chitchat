package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	cdb "chithat/db"

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
	handler := http.NewServeMux()

	handler.HandleFunc("POST /singin", singIn)
	handler.HandleFunc("POST /singup", singUp)
	handler.HandleFunc("/ws", ws)

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

func ws(w http.ResponseWriter, r *http.Request) {
}
