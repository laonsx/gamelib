package graceful

import (
	"log"
	"net/http"
)

func ListenAndServe(addr string, handler http.Handler) {

	err := http.ListenAndServe(addr, handler)
	if err != nil {

		log.Println("[error] Start service failed", "listen[", addr, "] error[", err, "]")
	}
}
