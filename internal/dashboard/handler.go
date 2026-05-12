package dashboard

import (
	"net/http"
	"strings"

	"profile-backend/internal/contact"
	"profile-backend/internal/profile"
	"profile-backend/internal/web"
)

type Handler struct {
	contacts   contact.Store
	profiles   profile.Store
	adminToken string
}

func NewHandler(contacts contact.Store, profiles profile.Store, adminToken string) *Handler {
	return &Handler{contacts: contacts, profiles: profiles, adminToken: strings.TrimSpace(adminToken)}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	if h.adminToken == "" || strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") != h.adminToken {
		web.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	messages, err := h.contacts.List()
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not read messages")
		return
	}
	profile, err := h.profiles.Get()
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not read profile")
		return
	}
	counts := map[string]int{"new": 0, "read": 0, "archived": 0}
	for _, message := range messages {
		counts[message.Status]++
	}
	recent := messages
	if len(recent) > 5 {
		recent = recent[:5]
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{"profile": profile, "stats": map[string]any{"totalMessages": len(messages), "newMessages": counts["new"], "readMessages": counts["read"], "archived": counts["archived"]}, "recentMessages": recent})
}
