package contact

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"profile-backend/internal/web"
)

type Handler struct {
	store      Store
	adminToken string
}

func NewHandler(store Store, adminToken string) *Handler {
	return &Handler{store: store, adminToken: strings.TrimSpace(adminToken)}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		web.WriteError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	input.Budget = strings.TrimSpace(input.Budget)
	input.Message = strings.TrimSpace(input.Message)
	input.Source = strings.TrimSpace(input.Source)
	if input.Source == "" {
		input.Source = "portfolio"
	}
	if input.Name == "" || input.Email == "" || input.Message == "" {
		web.WriteError(w, http.StatusBadRequest, "Name, email, and message are required")
		return
	}
	if len(input.Name) > 120 || len(input.Email) > 160 || len(input.Budget) > 80 || len(input.Message) > 3000 {
		web.WriteError(w, http.StatusBadRequest, "Submitted message is too long")
		return
	}
	if _, err := mail.ParseAddress(input.Email); err != nil {
		web.WriteError(w, http.StatusBadRequest, "Email address is invalid")
		return
	}
	message := Message{ID: newID(), Name: input.Name, Email: input.Email, Budget: input.Budget, Message: input.Message, Source: input.Source, Status: "new", CreatedAt: time.Now().UTC()}
	if err := h.store.Save(message); err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not save message")
		return
	}
	web.WriteJSON(w, http.StatusCreated, map[string]string{"id": message.ID, "message": "Message received"})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		web.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	messages, err := h.store.List()
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not read messages")
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{"messages": messages})
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		web.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var input StatusRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		web.WriteError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	status := strings.TrimSpace(input.Status)
	if status != "new" && status != "read" && status != "archived" {
		web.WriteError(w, http.StatusBadRequest, "Status must be new, read, or archived")
		return
	}
	message, found, err := h.store.UpdateStatus(r.PathValue("id"), status)
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not update message")
		return
	}
	if !found {
		web.WriteError(w, http.StatusNotFound, "Message not found")
		return
	}
	web.WriteJSON(w, http.StatusOK, message)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		web.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	found, err := h.store.Delete(r.PathValue("id"))
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not delete message")
		return
	}
	if !found {
		web.WriteError(w, http.StatusNotFound, "Message not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) authorized(r *http.Request) bool {
	return h.adminToken != "" && strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") == h.adminToken
}
func newID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(bytes)
}
