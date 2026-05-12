package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"profile-backend/internal/web"
)

type Handler struct {
	adminToken string
	fcmKey     string
	client     *http.Client
}

func NewHandler(adminToken string, fcmKey string) *Handler {
	return &Handler{adminToken: strings.TrimSpace(adminToken), fcmKey: strings.TrimSpace(fcmKey), client: &http.Client{}}
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		web.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{
		"provider":   "firebase-fcm",
		"configured": h.fcmKey != "",
		"features":   []string{"task reminders", "bill reminders", "daily summary", "weekly review", "overspending alerts", "document expiry alerts"},
	})
}

func (h *Handler) SendTest(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		web.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var input struct {
		Token string `json:"token"`
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		web.WriteError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if h.fcmKey == "" {
		web.WriteJSON(w, http.StatusAccepted, map[string]any{"queued": false, "message": "FCM_SERVER_KEY is not configured"})
		return
	}
	if strings.TrimSpace(input.Token) == "" {
		web.WriteError(w, http.StatusBadRequest, "Token is required")
		return
	}
	if input.Title == "" {
		input.Title = "Babu Personal OS"
	}
	if input.Body == "" {
		input.Body = "Notifications are connected."
	}
	if err := h.sendLegacy(input.Token, input.Title, input.Body); err != nil {
		web.WriteError(w, http.StatusBadGateway, err.Error())
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{"sent": true})
}

func (h *Handler) sendLegacy(token, title, body string) error {
	payload := map[string]any{
		"to": token,
		"notification": map[string]string{
			"title": title,
			"body":  body,
		},
	}
	encoded, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, "https://fcm.googleapis.com/fcm/send", bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "key="+h.fcmKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("FCM request failed with status %d", resp.StatusCode)
	}
	return nil
}

func (h *Handler) authorized(r *http.Request) bool {
	return h.adminToken != "" && strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") == h.adminToken
}
