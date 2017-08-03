package cracker

import (
	"fmt"
	"net/http"
)

const (
	HeadError    = 500
	HeadOK       = 200
	HeadData     = 201
	HeadHeart    = 202
	HeadQuit     = 203
	HeadNotFound = 404
)

func WriteHTTPError(w http.ResponseWriter, message string) {
	w.WriteHeader(HeadError)
	fmt.Fprintf(w, "%s", message)
}

func WriteNotFoundError(w http.ResponseWriter, message string) {
	w.WriteHeader(HeadNotFound)
	fmt.Fprintf(w, "%s", message)
}

func WriteHTTPOK(w http.ResponseWriter, data string) {
	w.WriteHeader(HeadOK)
	fmt.Fprintf(w, "%s", data)
}

func WriteHTTPData(w http.ResponseWriter, data []byte) {
	w.WriteHeader(HeadData)
	w.Write(data)
}

func WriteHTTPQuit(w http.ResponseWriter, data string) {
	w.WriteHeader(HeadQuit)
	fmt.Fprintf(w, "%s", data)
}

func WriteHTTPHeart(w http.ResponseWriter, data string) {
	w.WriteHeader(HeadHeart)
	fmt.Fprintf(w, "%s", data)
}
