package personal

import "time"

type Note struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Tags      []string  `json:"tags"`
	Pinned    bool      `json:"pinned"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Reminder struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Notes     string    `json:"notes"`
	DueAt     time.Time `json:"dueAt"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Transaction struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Amount     float64   `json:"amount"`
	Category   string    `json:"category"`
	Note       string    `json:"note"`
	OccurredAt time.Time `json:"occurredAt"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type Habit struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Target        string    `json:"target"`
	Frequency     string    `json:"frequency"`
	Streak        int       `json:"streak"`
	LastCheckedAt time.Time `json:"lastCheckedAt"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type State struct {
	Notes        []Note        `json:"notes"`
	Reminders    []Reminder    `json:"reminders"`
	Transactions []Transaction `json:"transactions"`
	Habits       []Habit       `json:"habits"`
}

type NoteRequest struct {
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	Tags   []string `json:"tags"`
	Pinned bool     `json:"pinned"`
}

type ReminderRequest struct {
	Title string `json:"title"`
	Notes string `json:"notes"`
	DueAt string `json:"dueAt"`
	Done  bool   `json:"done"`
}

type TransactionRequest struct {
	Type       string  `json:"type"`
	Amount     float64 `json:"amount"`
	Category   string  `json:"category"`
	Note       string  `json:"note"`
	OccurredAt string  `json:"occurredAt"`
}

type HabitRequest struct {
	Name      string `json:"name"`
	Target    string `json:"target"`
	Frequency string `json:"frequency"`
}
