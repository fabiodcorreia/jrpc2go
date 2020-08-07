package jrpc2go

import (
	"net/http"
	"strings"
)

const contentTypeKey = "Content-Type"
const contentTypeValue = "application/json"

// HTTPHandleFunc it's an helper function to mediate http requests to JSON RPC and back.
func HTTPHandleFunc(m *Manager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get(contentTypeKey), contentTypeValue) {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		if r.ContentLength == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Add(contentTypeKey, contentTypeValue)

		err := m.Handle(r.Context(), r.Body, w)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte(err.Error())); err != nil {
				//TODO not sure what to do here
			}
		}
	}
}
