package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"profile-backend/internal/personal"
	"profile-backend/internal/web"
)

type Handler struct {
	store      personal.Store
	client     *Client
	adminToken string
}

func NewHandler(store personal.Store, client *Client, adminToken string) *Handler {
	return &Handler{store: store, client: client, adminToken: strings.TrimSpace(adminToken)}
}

func (h *Handler) DailyBriefing(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		web.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	state, err := h.store.Summary()
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not read personal workspace")
		return
	}
	prompt := buildDailyPrompt(state)
	text, provider, err := h.client.Complete(r.Context(), prompt)
	if err != nil {
		text = localBriefing(state)
		provider = "local"
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{
		"provider": provider,
		"briefing": text,
	})
}

func (h *Handler) NoteSummary(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		web.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var input struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		web.WriteError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	body := strings.TrimSpace(input.Body)
	if body == "" {
		web.WriteError(w, http.StatusBadRequest, "Body is required")
		return
	}
	prompt := fmt.Sprintf("Summarize this note in 2 short bullets and suggest 3 tags.\nTitle: %s\nNote: %s", input.Title, body)
	text, provider, err := h.client.Complete(r.Context(), prompt)
	if err != nil {
		text = fallbackSummary(body)
		provider = "local"
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{"provider": provider, "summary": text})
}

func (h *Handler) authorized(r *http.Request) bool {
	return h.adminToken != "" && strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") == h.adminToken
}

type Client struct {
	Provider string
	APIKey   string
	Model    string
	HTTP     *http.Client
}

func NewClient(provider, apiKey, model string) *Client {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		provider = "local"
	}
	if model == "" {
		if provider == "groq" {
			model = "llama-3.1-8b-instant"
		} else {
			model = "gemini-1.5-flash"
		}
	}
	return &Client{Provider: provider, APIKey: strings.TrimSpace(apiKey), Model: model, HTTP: &http.Client{Timeout: 20 * time.Second}}
}

func (c *Client) Complete(ctx context.Context, prompt string) (string, string, error) {
	if c.APIKey == "" || c.Provider == "local" {
		return "", "local", fmt.Errorf("AI key not configured")
	}
	switch c.Provider {
	case "gemini":
		return c.completeGemini(ctx, prompt)
	case "groq":
		return c.completeGroq(ctx, prompt)
	default:
		return "", c.Provider, fmt.Errorf("unsupported AI_PROVIDER %q", c.Provider)
	}
}

func (c *Client) completeGemini(ctx context.Context, prompt string) (string, string, error) {
	payload := map[string]any{
		"contents": []map[string]any{{"parts": []map[string]string{{"text": prompt}}}},
	}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.Model, c.APIKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", "gemini", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", "gemini", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", "gemini", fmt.Errorf("gemini failed: %s", strings.TrimSpace(string(data)))
	}
	var output struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
		return "", "gemini", err
	}
	if len(output.Candidates) == 0 || len(output.Candidates[0].Content.Parts) == 0 {
		return "", "gemini", fmt.Errorf("empty gemini response")
	}
	return output.Candidates[0].Content.Parts[0].Text, "gemini", nil
}

func (c *Client) completeGroq(ctx context.Context, prompt string) (string, string, error) {
	payload := map[string]any{
		"model": c.Model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a concise personal AI operating system assistant."},
			{"role": "user", "content": prompt},
		},
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", "groq", err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", "groq", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", "groq", fmt.Errorf("groq failed: %s", strings.TrimSpace(string(data)))
	}
	var output struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
		return "", "groq", err
	}
	if len(output.Choices) == 0 {
		return "", "groq", fmt.Errorf("empty groq response")
	}
	return output.Choices[0].Message.Content, "groq", nil
}

func buildDailyPrompt(state personal.State) string {
	return fmt.Sprintf("Create a daily briefing for my personal OS. Notes: %d, reminders: %d, transactions: %d, habits: %d. Mention priorities, pending reminders, expense alerts, habit streaks, and next actions.",
		len(state.Notes), len(state.Reminders), len(state.Transactions), len(state.Habits))
}

func localBriefing(state personal.State) string {
	open := 0
	for _, reminder := range state.Reminders {
		if !reminder.Done {
			open++
		}
	}
	var expense float64
	for _, transaction := range state.Transactions {
		if transaction.Type == "expense" {
			expense += transaction.Amount
		}
	}
	return fmt.Sprintf("Today you have %d open task(s), %d note(s), %d habit(s), and Rs %.0f tracked expenses. Review urgent tasks first, complete one routine check-in, tag new notes, and check budget alerts.", open, len(state.Notes), len(state.Habits), expense)
}

func fallbackSummary(body string) string {
	words := strings.Fields(body)
	if len(words) > 24 {
		words = words[:24]
	}
	return "Summary: " + strings.Join(words, " ")
}
