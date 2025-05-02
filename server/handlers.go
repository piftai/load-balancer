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

// Helper functions for responses
func (h *ClientHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

func (h *ClientHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (h *ClientHandler) Create(w http.ResponseWriter, r *http.Request) {
	var client repository.Client
	if err := json.NewDecoder(r.Body).Decode(&client); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate client data
	if client.ID == "" { // example validation
		h.respondWithError(w, http.StatusBadRequest, "Client name is required")
		return
	}

	// Set ID if not provided
	if client.ID == "" {
		if r.Header.Get("X-Client-ID") != "" {
			client.ID = r.Header.Get("X-Client-ID")
		} else {
			client.ID = strings.Split(r.RemoteAddr, ":")[0]
		}
	}

	if err := h.repository.Create(client); err != nil {
		log.Printf("Error creating client: %v\n", err)
		h.respondWithError(w, http.StatusInternalServerError, "Error creating client")
		return
	}

	log.Printf("New client created. ID: %s\n", client.ID)
	h.respondWithJSON(w, http.StatusCreated, map[string]string{"id": client.ID})
}

func (h *ClientHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Client ID is required")
		return
	}

	client, err := h.repository.Get(id)
	if err != nil {
		log.Printf("Error getting client %s: %v\n", id, err)
		h.respondWithError(w, http.StatusInternalServerError, "Error retrieving client")
		return
	}

	if client == nil {
		h.respondWithError(w, http.StatusNotFound, "Client not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, client)
}

func (h *ClientHandler) Update(w http.ResponseWriter, r *http.Request) {
	var client repository.Client
	if err := json.NewDecoder(r.Body).Decode(&client); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if client.ID == "" {
		h.respondWithError(w, http.StatusBadRequest, "Client ID is required")
		return
	}

	if err := h.repository.Update(client); err != nil {
		log.Printf("Error updating client %s: %v\n", client.ID, err)
		h.respondWithError(w, http.StatusInternalServerError, "Error updating client")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *ClientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Client ID is required")
		return
	}

	if err := h.repository.Delete(id); err != nil {
		log.Printf("Error deleting client %s: %v\n", id, err)
		h.respondWithError(w, http.StatusInternalServerError, "Error deleting client")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
