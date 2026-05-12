package profile

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"

	"profile-backend/internal/web"
)

type Handler struct {
	store      Store
	adminToken string
}

func NewHandler(store Store, adminToken string) *Handler {
	return &Handler{store: store, adminToken: strings.TrimSpace(adminToken)}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	profile, err := h.store.Get()
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not load profile")
		return
	}
	web.WriteJSON(w, http.StatusOK, profile)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		web.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var input Profile
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		web.WriteError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	profile, ok := cleanProfile(input)
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Name, title, valid email, website, and summary are required")
		return
	}
	profile, err := h.store.Save(profile)
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not save profile")
		return
	}
	web.WriteJSON(w, http.StatusOK, profile)
}

func (h *Handler) authorized(r *http.Request) bool {
	return h.adminToken != "" && strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") == h.adminToken
}

func cleanProfile(input Profile) (Profile, bool) {
	input.Name = strings.TrimSpace(input.Name)
	input.Title = strings.TrimSpace(input.Title)
	input.Location = strings.TrimSpace(input.Location)
	input.Email = strings.TrimSpace(input.Email)
	input.Website = strings.TrimSpace(input.Website)
	input.Summary = strings.TrimSpace(input.Summary)
	if input.Name == "" || input.Title == "" || input.Email == "" || input.Website == "" || input.Summary == "" {
		return Profile{}, false
	}
	if len(input.Name) > 120 || len(input.Title) > 160 || len(input.Location) > 120 || len(input.Summary) > 1200 {
		return Profile{}, false
	}
	if _, err := mail.ParseAddress(input.Email); err != nil {
		return Profile{}, false
	}
	if !strings.HasPrefix(input.Website, "https://") && !strings.HasPrefix(input.Website, "http://") {
		return Profile{}, false
	}
	input.Highlights = cleanMetrics(input.Highlights)
	input.Links = cleanLinks(input.Links)
	input.FocusAreas = cleanStrings(input.FocusAreas, 12, 80)
	return input, true
}

func cleanMetrics(items []Metric) []Metric {
	out := make([]Metric, 0, len(items))
	for _, item := range items {
		label := strings.TrimSpace(item.Label)
		value := strings.TrimSpace(item.Value)
		if label != "" && value != "" && len(label) <= 80 && len(value) <= 80 {
			out = append(out, Metric{Label: label, Value: value})
		}
	}
	return out
}

func cleanLinks(items []ProfileLink) []ProfileLink {
	out := make([]ProfileLink, 0, len(items))
	for _, item := range items {
		label := strings.TrimSpace(item.Label)
		url := strings.TrimSpace(item.URL)
		if label != "" && len(label) <= 80 && (strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "mailto:")) {
			out = append(out, ProfileLink{Label: label, URL: url})
		}
	}
	return out
}

func cleanStrings(items []string, limit int, maxLen int) []string {
	out := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" || len(value) > maxLen || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
		if len(out) == limit {
			break
		}
	}
	return out
}
