package personal

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
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

func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		web.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	state, err := h.store.Summary()
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not read personal workspace")
		return
	}
	income, expense := moneyTotals(state.Transactions)
	web.WriteJSON(w, http.StatusOK, map[string]any{
		"notesCount":        len(state.Notes),
		"openReminders":     countOpenReminders(state.Reminders),
		"transactionsCount": len(state.Transactions),
		"habitsCount":       len(state.Habits),
		"checkedToday":      countCheckedToday(state.Habits),
		"income":            income,
		"expense":           expense,
		"balance":           income - expense,
		"recentNotes":       takeNotes(state.Notes, 3),
		"upcomingReminders": takeReminders(state.Reminders, 4),
		"habits":            state.Habits,
	})
}

func (h *Handler) ListNotes(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		web.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	notes, err := h.store.ListNotes()
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not read notes")
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{"notes": notes})
}

func (h *Handler) CreateNote(w http.ResponseWriter, r *http.Request) {
	var input NoteRequest
	if !h.decodeAuthorized(w, r, &input) {
		return
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Body = strings.TrimSpace(input.Body)
	if input.Title == "" {
		web.WriteError(w, http.StatusBadRequest, "Title is required")
		return
	}
	if len(input.Title) > 160 || len(input.Body) > 5000 {
		web.WriteError(w, http.StatusBadRequest, "Note is too long")
		return
	}
	now := time.Now().UTC()
	note := Note{ID: newID(), Title: input.Title, Body: input.Body, Tags: cleanTags(input.Tags), Pinned: input.Pinned, CreatedAt: now, UpdatedAt: now}
	note, err := h.store.CreateNote(note)
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not save note")
		return
	}
	web.WriteJSON(w, http.StatusCreated, note)
}

func (h *Handler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	h.deleteByID(w, r, "note", h.store.DeleteNote)
}

func (h *Handler) ListReminders(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		web.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	reminders, err := h.store.ListReminders()
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not read reminders")
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{"reminders": reminders})
}

func (h *Handler) CreateReminder(w http.ResponseWriter, r *http.Request) {
	var input ReminderRequest
	if !h.decodeAuthorized(w, r, &input) {
		return
	}
	reminder, ok := reminderFromRequest(input)
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Title and valid dueAt are required")
		return
	}
	reminder, err := h.store.CreateReminder(reminder)
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not save reminder")
		return
	}
	web.WriteJSON(w, http.StatusCreated, reminder)
}

func (h *Handler) ToggleReminder(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		web.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	reminder, found, err := h.store.ToggleReminder(r.PathValue("id"))
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not save reminder")
		return
	}
	if !found {
		web.WriteError(w, http.StatusNotFound, "Reminder not found")
		return
	}
	web.WriteJSON(w, http.StatusOK, reminder)
}

func (h *Handler) DeleteReminder(w http.ResponseWriter, r *http.Request) {
	h.deleteByID(w, r, "reminder", h.store.DeleteReminder)
}

func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		web.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	transactions, err := h.store.ListTransactions()
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not read transactions")
		return
	}
	income, expense := moneyTotals(transactions)
	web.WriteJSON(w, http.StatusOK, map[string]any{"transactions": transactions, "income": income, "expense": expense, "balance": income - expense})
}

func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var input TransactionRequest
	if !h.decodeAuthorized(w, r, &input) {
		return
	}
	transaction, ok := transactionFromRequest(input)
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Valid type, amount, category, and occurredAt are required")
		return
	}
	transaction, err := h.store.CreateTransaction(transaction)
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not save transaction")
		return
	}
	web.WriteJSON(w, http.StatusCreated, transaction)
}

func (h *Handler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	h.deleteByID(w, r, "transaction", h.store.DeleteTransaction)
}

func (h *Handler) ListHabits(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		web.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	habits, err := h.store.ListHabits()
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not read habits")
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{"habits": habits})
}

func (h *Handler) CreateHabit(w http.ResponseWriter, r *http.Request) {
	var input HabitRequest
	if !h.decodeAuthorized(w, r, &input) {
		return
	}
	habit, ok := habitFromRequest(input)
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Name and target are required")
		return
	}
	habit, err := h.store.CreateHabit(habit)
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not save habit")
		return
	}
	web.WriteJSON(w, http.StatusCreated, habit)
}

func (h *Handler) CheckInHabit(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		web.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	habit, found, err := h.store.CheckInHabit(r.PathValue("id"))
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not check in habit")
		return
	}
	if !found {
		web.WriteError(w, http.StatusNotFound, "Habit not found")
		return
	}
	web.WriteJSON(w, http.StatusOK, habit)
}

func (h *Handler) DeleteHabit(w http.ResponseWriter, r *http.Request) {
	h.deleteByID(w, r, "habit", h.store.DeleteHabit)
}

func (h *Handler) decodeAuthorized(w http.ResponseWriter, r *http.Request, input any) bool {
	if !h.authorized(r) {
		web.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return false
	}
	if err := json.NewDecoder(r.Body).Decode(input); err != nil {
		web.WriteError(w, http.StatusBadRequest, "Invalid JSON body")
		return false
	}
	return true
}

func (h *Handler) deleteByID(w http.ResponseWriter, r *http.Request, label string, remove func(string) (bool, error)) {
	if !h.authorized(r) {
		web.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	found, err := remove(r.PathValue("id"))
	if err != nil {
		web.WriteError(w, http.StatusInternalServerError, "Could not delete "+label)
		return
	}
	if !found {
		web.WriteError(w, http.StatusNotFound, titleCase(label)+" not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) authorized(r *http.Request) bool {
	return h.adminToken != "" && strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") == h.adminToken
}

func reminderFromRequest(input ReminderRequest) (Reminder, bool) {
	title := strings.TrimSpace(input.Title)
	dueAt, err := time.Parse(time.RFC3339, strings.TrimSpace(input.DueAt))
	if title == "" || err != nil {
		return Reminder{}, false
	}
	now := time.Now().UTC()
	return Reminder{ID: newID(), Title: title, Notes: strings.TrimSpace(input.Notes), DueAt: dueAt.UTC(), Done: input.Done, CreatedAt: now, UpdatedAt: now}, true
}

func transactionFromRequest(input TransactionRequest) (Transaction, bool) {
	kind := strings.ToLower(strings.TrimSpace(input.Type))
	category := strings.TrimSpace(input.Category)
	occurredAt, err := time.Parse(time.RFC3339, strings.TrimSpace(input.OccurredAt))
	if (kind != "income" && kind != "expense") || input.Amount <= 0 || category == "" || err != nil {
		return Transaction{}, false
	}
	now := time.Now().UTC()
	return Transaction{ID: newID(), Type: kind, Amount: input.Amount, Category: category, Note: strings.TrimSpace(input.Note), OccurredAt: occurredAt.UTC(), CreatedAt: now, UpdatedAt: now}, true
}

func habitFromRequest(input HabitRequest) (Habit, bool) {
	name := strings.TrimSpace(input.Name)
	target := strings.TrimSpace(input.Target)
	frequency := strings.ToLower(strings.TrimSpace(input.Frequency))
	if frequency == "" {
		frequency = "daily"
	}
	if name == "" || target == "" || len(name) > 120 || len(target) > 240 {
		return Habit{}, false
	}
	now := time.Now().UTC()
	return Habit{ID: newID(), Name: name, Target: target, Frequency: frequency, CreatedAt: now, UpdatedAt: now}, true
}

func cleanTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	seen := map[string]bool{}
	for _, tag := range tags {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if tag == "" || seen[tag] {
			continue
		}
		seen[tag] = true
		out = append(out, tag)
	}
	return out
}

func moneyTotals(transactions []Transaction) (float64, float64) {
	var income, expense float64
	for _, transaction := range transactions {
		if transaction.Type == "income" {
			income += transaction.Amount
		} else {
			expense += transaction.Amount
		}
	}
	return income, expense
}

func countOpenReminders(reminders []Reminder) int {
	count := 0
	for _, reminder := range reminders {
		if !reminder.Done {
			count++
		}
	}
	return count
}

func countCheckedToday(habits []Habit) int {
	count := 0
	now := time.Now().UTC()
	for _, habit := range habits {
		if sameUTCDate(habit.LastCheckedAt, now) {
			count++
		}
	}
	return count
}

func takeNotes(notes []Note, count int) []Note {
	if len(notes) < count {
		return notes
	}
	return notes[:count]
}

func takeReminders(reminders []Reminder, count int) []Reminder {
	open := make([]Reminder, 0, len(reminders))
	for _, reminder := range reminders {
		if !reminder.Done {
			open = append(open, reminder)
		}
	}
	if len(open) < count {
		return open
	}
	return open[:count]
}

func titleCase(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func newID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(bytes)
}
