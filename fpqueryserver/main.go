package main

import (
	"bytes"
	"encoding/json"
	"github.com/Muzzik-Dev-Group/muzzik-fingerprint"
	"io"
	"log"
	"net/http"
)

var jsonMIME = "application/json"

type HttpError struct {
	Message    string `json:"error"`
	StatusCode int
}

func (httpErr *HttpError) Error() string {
	return httpErr.Message
}

func NewHttpError(code int, message string) (httpErr *HttpError) {
	httpErr = &HttpError{message, code}
	return
}

func HandleError(res http.ResponseWriter, httpErr *HttpError) {
	res.WriteHeader(httpErr.StatusCode)
	header := res.Header()
	if header != nil {
		header.Add("Content-Type", "application/json; charset=utf-8")
	}

	err := json.NewEncoder(res).Encode(httpErr)
	if err != nil {
		HandleError(res, NewHttpError(http.StatusInternalServerError, err.Error()))
	}
}

func queryHandler(res http.ResponseWriter, req *http.Request) {
	var err error

	if req.Method != "POST" {
		httpErr := NewHttpError(http.StatusMethodNotAllowed, "only POST allowed")
		HandleError(res, httpErr)
		return
	}

	defer req.Body.Close()
	var fps []string
	var buf bytes.Buffer
	io.Copy(&buf, req.Body)
	err = json.NewDecoder(&buf).Decode(&fps)
	if err != nil {
		HandleError(res, NewHttpError(http.StatusBadRequest, err.Error()))
		return
	}

	var music *muzzikfp.Music
	music, err = muzzikfp.GetBestMatch(fps)
	if err != nil {
		HandleError(res, NewHttpError(http.StatusInternalServerError, err.Error()))
		return
	}

	err = json.NewEncoder(res).Encode(music)
	if err != nil {
		if httpErr, ok := err.(*HttpError); ok {
			HandleError(res, httpErr)
		} else {
			HandleError(res, NewHttpError(http.StatusInternalServerError, err.Error()))
		}
	}
	return
}

func main() {
	http.HandleFunc("/query", queryHandler)
	println("Server start at 8090")
	log.Fatal(http.ListenAndServe(":8090", nil))
}
