package main

import (
	"bytes"
	"crypto/rand"
	"log"
	"math/big"
	"net/http"
	"time"
)

// ConsciousnessHandler writes to the page letters in random order, hopefully they end up as a helpful message
func ConsciousnessHandler(w http.ResponseWriter, r *http.Request) {
	// Listen to connection close and un-register messageChan
	notify := w.(http.CloseNotifier).CloseNotify()

	if flusher, ok := w.(http.Flusher); ok {
		//w.Header().Set("Content-Type", "text/event-stream")
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
				w.Write([]byte(letter()))
				flusher.Flush()
				log.Println("wrote")
			}
			time.Sleep(10 * time.Second)
		}

	}
	w.WriteHeader(http.StatusOK)
}

func letter() string {
	alphaNum := []string{
		"a",
		"b",
		"c",
		"d",
		"e",
		"f",
		"g",
		"h",
		"i",
		"j",
		"k",
		"l",
		"m",
		"n",
		"o",
		"p",
		"q",
		"r",
		"s",
		"t",
		"u",
		"v",
		"w",
		"x",
		"y",
		"z",
		" ",
		"0",
		"1",
		"2",
		"3",
		"4",
		"5",
		"6",
		"7",
		"8",
		"9",
		//		"9", // duplicate because rand may skip the first and last entry
	}
	c := 10
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		return "a"
	}

	x, e := rand.Int(bytes.NewReader(b), big.NewInt(int64(len(alphaNum))))
	if e != nil {
		return "a"
	}
	return alphaNum[x.Int64()]
}
