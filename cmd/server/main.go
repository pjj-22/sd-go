package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

func main() {

	mux := http.NewServeMux()
	var (
		mu    sync.Mutex
		calls int
	)

	mux.HandleFunc("POST /hello", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		n := calls
		mu.Unlock()
		log.Printf("call %d from %s", n, r.Header.Get("X-User-ID"))
		fmt.Fprintln(w, "hello world")
	})
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}