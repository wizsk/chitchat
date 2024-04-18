package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
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
			http.Redirect(w, r, "/signin", http.StatusMovedPermanently)
			return
		}
		puser, ok := cookie[v.Value]
		if !ok {
			http.Redirect(w, r, "/signin", http.StatusMovedPermanently)
			return
		}

		nr := r.WithContext(context.WithValue(r.Context(), "puser", puser))
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
	handler.Handle("GET /", ensureSignedin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(mustDo(os.ReadFile("templates/index.html")))
	})))

	handler.HandleFunc("GET /signin", func(w http.ResponseWriter, r *http.Request) {
		if checkCookie(r) != nil {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
			return
		}
		w.Write(mustDo(os.ReadFile("templates/signin.html")))
	})
	handler.HandleFunc("GET /signup", func(w http.ResponseWriter, r *http.Request) {
		if checkCookie(r) != nil {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
			return
		}
		w.Write(mustDo(os.ReadFile("templates/signup.html")))
	})

	handler.HandleFunc("POST /signin", signIn)
	handler.HandleFunc("POST /signup", signUp)
	handler.Handle("GET /ws", ensureSignedin(websocket.Handler(ws)))

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
	cookieFile = "/tmp/cooke.json"
	cookie     = func() map[string]*cdb.User {
		v := map[string]*cdb.User{}
		d, err := os.ReadFile(cookieFile)
		if os.IsNotExist(err) {
			return v
		}
		mustDo[*byte](nil, json.Unmarshal(d, &v))
		return v
	}()
	cookieLock sync.RWMutex
	cookieName = "__chitchat_coookie"
	db         = mustDo(cdb.InitDB("user=postgres password=pass host=localhost port=5432 sslmode=disable"))
)

func saveC(str string, u *cdb.User) {
	cookieLock.Lock()
	defer cookieLock.Unlock()
	cookie[str] = u
	if err := os.WriteFile(cookieFile, mustDo(json.Marshal(cookie)), 0770); err != nil {
		panic(err)
	}
}

func getCleanedFormValue(r *http.Request, v string) string {
	return strings.TrimSpace(r.FormValue(v))
}

func signUp(w http.ResponseWriter, r *http.Request) {
	name := getCleanedFormValue(r, "name")
	username := getCleanedFormValue(r, "username")
	password := getCleanedFormValue(r, "password")

	if name == "" || username == "" || password == "" {
		// http.Error(w, "bad input", http.StatusBadRequest)
		http.Redirect(w, r, "/signup", http.StatusMovedPermanently)
		return
	}

	user, err := db.Singup(name, username, password)
	if err != nil {
		http.Redirect(w, r, "/signup", http.StatusMovedPermanently)
		// http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// d, err := json.Marshal(user)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }
	//
	if !writeCookie(w, user) {
		return
	}
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
	//
	// w.WriteHeader(http.StatusCreated)
	// w.Write(d)
}

func signIn(w http.ResponseWriter, r *http.Request) {
	username := getCleanedFormValue(r, "username")
	password := getCleanedFormValue(r, "password")

	if username == "" || password == "" {
		// http.Error(w, "bad input", http.StatusBadRequest)
		http.Redirect(w, r, "/signin", http.StatusMovedPermanently)
		return
	}

	user, err := db.Singin(username, password)
	if err != nil {
		// http.Error(w, err.Error(), http.StatusBadRequest)
		http.Redirect(w, r, "/signin", http.StatusMovedPermanently)
		return
	}

	// d, err := json.Marshal(user)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }

	if !writeCookie(w, user) {
		return
	}
	http.Redirect(w, r, "/", http.StatusMovedPermanently)

	// w.WriteHeader(http.StatusFound)
	// w.Write(d)
}

func writeCookie(w http.ResponseWriter, user cdb.User) bool {
	buff := [32]byte{}
	if _, err := rand.Read(buff[:]); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	cv := base64.StdEncoding.EncodeToString(buff[:])
	saveC(cv, &user)
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: cv})
	return true
}

func checkCookie(r *http.Request) *cdb.User {
	v, err := r.Cookie(cookieName)
	if err != nil {
		return nil
	}
	// pointer to user
	user, ok := cookie[v.Value]
	if !ok {
		return nil
	}

	return user
}

func mustDo[T any](t T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func ws(conn *websocket.Conn) {
	defer conn.Close()
	u, ok := conn.Request().Context().Value("puser").(*cdb.User)
	if !ok {
		log.Println("ws: typecast to *User was unsuccessfull")
		return
	}

	conn.Write([]byte("Hello"))
	conn.Write(mustDo(json.Marshal(WsData{DataType: wsdt(WsDataUser), User: u})))

	for {
		// if err := websocket.Message.Receive(conn, &); err != nil {
		// 	break
		msg := WsData{}
		data := make([]byte, 1024)
		i, err := conn.Read(data)
		if err != nil {
			if _, err = conn.Write([]byte("?r")); err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(err)
			continue
		}

		// }
		// fmt.Println(msg, d)
		fmt.Println(json.Unmarshal(data[:i], &msg))
		fmt.Fprintf(conn, "hii go a msg form yaaa: %q", string(data[:i]))
	}
}
