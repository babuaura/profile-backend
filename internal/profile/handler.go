package profile

import (
	"net/http"
	"profile-backend/internal/web"
)

type Handler struct{ store Store }

func NewHandler(store Store) *Handler { return &Handler{store: store} }
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	profile, err := h.store.Get()
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not load profile")
		return
	}
	web.WriteJSON(w, http.StatusOK, profile)
}
