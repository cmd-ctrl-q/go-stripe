package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
)

// writeJSON writes arbitrary data out as JSON
func (app *application) writeJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// check if headers is set
	if len(headers) > 0 {
		// set all headers
		for k, v := range headers[0] {
			w.Header()[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(out)

	return nil
}

// readJSON reads/decodes JSON data from request body
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	// dont handler data larger than 1mb
	maxBytes := 1048576

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// decode data of the request body into a decoder
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	// decode the decoded data into struct
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only have a single JSON value")
	}

	return nil
}

// badRequest sends a bad request error payload back to the user
func (app *application) badRequest(w http.ResponseWriter, r *http.Request, err error) error {
	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	// populate payload
	payload.Error = true
	payload.Message = err.Error()

	// write json
	out, err := json.MarshalIndent(payload, "", "\t")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(out)
	return nil
}

func (app *application) CreateDirIfNotExist(path string) error {
	const mode = 0755
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, mode)
		if err != nil {
			app.errorLog.Println(err)
			return err
		}
	}
	return nil
}
