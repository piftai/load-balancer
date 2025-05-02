package server

import (
	"encoding/json"
	"github.com/piftai/load-balancer/repository"
	"log"
	"net/http"
	"strings"
)

type ClientHandler struct {
	repository repository.Repository
}

func NewClientHandler(repository repository.Repository) *ClientHandler {
	return &ClientHandler{
		repository: repository,
	}
}

func (h *ClientHandler) Create(w http.ResponseWriter, r *http.Request) {
	var client repository.Client
	if err := json.NewDecoder(r.Body).Decode(&client); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		return
	}
	if client.ID == "" {
		if r.Header.Get("X-Client-ID") != "" {
			client.ID = r.Header.Get("X-Client-ID")
		} else {
			client.ID = strings.Split(r.RemoteAddr, ":")[0]
		}
	}
	if err := h.repository.Create(client); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		return
	}
	log.Println("New client created. ID:", client.ID)
	w.WriteHeader(http.StatusCreated)
}

func (h *ClientHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	client, err := h.repository.Get(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	if client == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *ClientHandler) Update(w http.ResponseWriter, r *http.Request) {
	var client repository.Client
	if err := json.NewDecoder(r.Body).Decode(&client); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	if err := h.repository.Update(client); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *ClientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	if err := h.repository.Delete(id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
