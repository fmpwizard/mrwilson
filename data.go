package main

import (
	"crypto/rand"
	"log"
	"net/http"
)

// DataHandler generates infinite data to test speed
func DataHandler(w http.ResponseWriter, r *http.Request) {
	// Listen to connection close and un-register messageChan
	notify := w.(http.CloseNotifier).CloseNotify()

	if flusher, ok := w.(http.Flusher); ok {
		//w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=\"data.bin\"")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		for {
			select {
			case <-notify:
				log.Println("boom")
				return
			default:
				w.Write(data())
				flusher.Flush()
			}
		}

	}
	w.WriteHeader(http.StatusOK)
}

func data() []byte {

	c := 10
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		return b
	}

	return b
}
