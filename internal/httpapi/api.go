// Package httpapi exposes the todo domain as a small JSON API.
package httpapi

import (
	"embed"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/ludanortmun/peco/internal/core"
)

//go:embed openapi.json swagger.html
var docs embed.FS

func New(client *core.PecoClient) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /todos", func(w http.ResponseWriter, r *http.Request) {
		includeDone := false
		if r.URL.Query().Has("includeDone") {
			var err error
			includeDone, err = strconv.ParseBool(r.URL.Query().Get("includeDone"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		items, err := client.List(includeDone)
		if err != nil {
			domainError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	})
	mux.HandleFunc("GET /todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := todoID(w, r)
		if !ok {
			return
		}
		item, err := client.Get(id)
		if err != nil {
			domainError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	})
	mux.HandleFunc("POST /todos", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Content string     `json:"content"`
			Due     *time.Time `json:"due"`
			State   core.State `json:"state"`
		}
		if !decode(w, r, &input) {
			return
		}
		if !validState(input.State) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		item := core.NewTodo(input.Content)
		item.Due, item.State = input.Due, input.State
		saved, err := client.Save(item)
		if err != nil {
			domainError(w, err)
			return
		}
		w.Header().Set("Location", "/todos/"+strconv.FormatUint(uint64(saved.Id), 10))
		writeJSON(w, http.StatusCreated, saved)
	})
	mux.HandleFunc("PUT /todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := todoID(w, r)
		if !ok {
			return
		}
		item, err := client.Get(id)
		if err != nil {
			domainError(w, err)
			return
		}
		var input struct {
			Id        uint        `json:"id"`
			Content   string      `json:"content"`
			State     *core.State `json:"state"`
			CreatedAt time.Time   `json:"createdAt"`
			Due       *time.Time  `json:"due"`
		}
		if !decode(w, r, &input) {
			return
		}
		if input.Id != item.Id || !input.CreatedAt.Equal(item.CreatedAt) || input.State == nil || !validState(*input.State) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		item.Content, item.Due, item.State = input.Content, input.Due, *input.State
		saved, err := client.Save(item)
		if err != nil {
			domainError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, saved)
	})
	mux.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, r *http.Request) {
		data, _ := docs.ReadFile("openapi.json")
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})
	mux.HandleFunc("GET /docs/{$}", func(w http.ResponseWriter, r *http.Request) {
		data, _ := docs.ReadFile("swagger.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})
	return mux
}

func validState(state core.State) bool { return state == core.Pending || state == core.Done }

func todoID(w http.ResponseWriter, r *http.Request) (uint, bool) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, strconv.IntSize)
	if err != nil || id == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return 0, false
	}
	return uint(id), true
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		w.WriteHeader(http.StatusBadRequest)
		return false
	}
	return true
}

func domainError(w http.ResponseWriter, err error) {
	switch err.Error() {
	case core.ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
	case core.ErrInvalidInput:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}
