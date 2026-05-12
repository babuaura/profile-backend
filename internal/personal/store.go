package personal

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type Store interface {
	Summary() (State, error)
	ListNotes() ([]Note, error)
	CreateNote(Note) (Note, error)
	DeleteNote(string) (bool, error)
	ListReminders() ([]Reminder, error)
	CreateReminder(Reminder) (Reminder, error)
	ToggleReminder(string) (Reminder, bool, error)
	DeleteReminder(string) (bool, error)
	ListTransactions() ([]Transaction, error)
	CreateTransaction(Transaction) (Transaction, error)
	DeleteTransaction(string) (bool, error)
	ListHabits() ([]Habit, error)
	CreateHabit(Habit) (Habit, error)
	CheckInHabit(string) (Habit, bool, error)
	DeleteHabit(string) (bool, error)
}

type FileStore struct {
	path string
	mu   sync.Mutex
}

func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

func (s *FileStore) Summary() (State, error) {
	return s.get()
}

func (s *FileStore) ListNotes() ([]Note, error) {
	state, err := s.get()
	return state.Notes, err
}

func (s *FileStore) CreateNote(note Note) (Note, error) {
	state, err := s.get()
	if err != nil {
		return Note{}, err
	}
	state.Notes = append(state.Notes, note)
	return note, s.save(state)
}

func (s *FileStore) DeleteNote(id string) (bool, error) {
	return s.deleteByID(func(state *State) bool {
		next := state.Notes[:0]
		found := false
		for _, item := range state.Notes {
			if item.ID == id {
				found = true
				continue
			}
			next = append(next, item)
		}
		state.Notes = next
		return found
	})
}

func (s *FileStore) ListReminders() ([]Reminder, error) {
	state, err := s.get()
	return state.Reminders, err
}

func (s *FileStore) CreateReminder(reminder Reminder) (Reminder, error) {
	state, err := s.get()
	if err != nil {
		return Reminder{}, err
	}
	state.Reminders = append(state.Reminders, reminder)
	return reminder, s.save(state)
}

func (s *FileStore) ToggleReminder(id string) (Reminder, bool, error) {
	state, err := s.get()
	if err != nil {
		return Reminder{}, false, err
	}
	for i := range state.Reminders {
		if state.Reminders[i].ID == id {
			state.Reminders[i].Done = !state.Reminders[i].Done
			state.Reminders[i].UpdatedAt = nowUTC()
			if err := s.save(state); err != nil {
				return Reminder{}, false, err
			}
			return state.Reminders[i], true, nil
		}
	}
	return Reminder{}, false, nil
}

func (s *FileStore) DeleteReminder(id string) (bool, error) {
	return s.deleteByID(func(state *State) bool {
		next := state.Reminders[:0]
		found := false
		for _, item := range state.Reminders {
			if item.ID == id {
				found = true
				continue
			}
			next = append(next, item)
		}
		state.Reminders = next
		return found
	})
}

func (s *FileStore) ListTransactions() ([]Transaction, error) {
	state, err := s.get()
	return state.Transactions, err
}

func (s *FileStore) CreateTransaction(transaction Transaction) (Transaction, error) {
	state, err := s.get()
	if err != nil {
		return Transaction{}, err
	}
	state.Transactions = append(state.Transactions, transaction)
	return transaction, s.save(state)
}

func (s *FileStore) DeleteTransaction(id string) (bool, error) {
	return s.deleteByID(func(state *State) bool {
		next := state.Transactions[:0]
		found := false
		for _, item := range state.Transactions {
			if item.ID == id {
				found = true
				continue
			}
			next = append(next, item)
		}
		state.Transactions = next
		return found
	})
}

func (s *FileStore) ListHabits() ([]Habit, error) {
	state, err := s.get()
	return state.Habits, err
}

func (s *FileStore) CreateHabit(habit Habit) (Habit, error) {
	state, err := s.get()
	if err != nil {
		return Habit{}, err
	}
	state.Habits = append(state.Habits, habit)
	return habit, s.save(state)
}

func (s *FileStore) CheckInHabit(id string) (Habit, bool, error) {
	state, err := s.get()
	if err != nil {
		return Habit{}, false, err
	}
	for i := range state.Habits {
		if state.Habits[i].ID == id {
			state.Habits[i] = checkedInHabit(state.Habits[i], nowUTC())
			if err := s.save(state); err != nil {
				return Habit{}, false, err
			}
			return state.Habits[i], true, nil
		}
	}
	return Habit{}, false, nil
}

func (s *FileStore) DeleteHabit(id string) (bool, error) {
	return s.deleteByID(func(state *State) bool {
		next := state.Habits[:0]
		found := false
		for _, item := range state.Habits {
			if item.ID == id {
				found = true
				continue
			}
			next = append(next, item)
		}
		state.Habits = next
		return found
	})
}

func (s *FileStore) Get() (State, error) {
	return s.get()
}

func (s *FileStore) Save(state State) error {
	return s.save(state)
}

func (s *FileStore) get() (State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return State{}, nil
	}
	if err != nil {
		return State{}, err
	}
	defer file.Close()

	var state State
	if err := json.NewDecoder(file).Decode(&state); err != nil {
		return State{}, err
	}
	sortState(&state)
	return state, nil
}

func (s *FileStore) save(state State) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sortState(&state)
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(file).Encode(state); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *FileStore) deleteByID(remove func(*State) bool) (bool, error) {
	state, err := s.get()
	if err != nil {
		return false, err
	}
	if !remove(&state) {
		return false, nil
	}
	return true, s.save(state)
}

func sortState(state *State) {
	sort.Slice(state.Notes, func(i, j int) bool {
		if state.Notes[i].Pinned != state.Notes[j].Pinned {
			return state.Notes[i].Pinned
		}
		return state.Notes[i].UpdatedAt.After(state.Notes[j].UpdatedAt)
	})
	sort.Slice(state.Reminders, func(i, j int) bool {
		if state.Reminders[i].Done != state.Reminders[j].Done {
			return !state.Reminders[i].Done
		}
		return state.Reminders[i].DueAt.Before(state.Reminders[j].DueAt)
	})
	sort.Slice(state.Transactions, func(i, j int) bool {
		return state.Transactions[i].OccurredAt.After(state.Transactions[j].OccurredAt)
	})
	sort.Slice(state.Habits, func(i, j int) bool {
		if state.Habits[i].Streak != state.Habits[j].Streak {
			return state.Habits[i].Streak > state.Habits[j].Streak
		}
		return state.Habits[i].UpdatedAt.After(state.Habits[j].UpdatedAt)
	})
}

func nowUTC() time.Time {
	return time.Now().UTC()
}

func checkedInHabit(habit Habit, now time.Time) Habit {
	if sameUTCDate(habit.LastCheckedAt, now) {
		habit.UpdatedAt = now
		return habit
	}
	if sameUTCDate(habit.LastCheckedAt, now.AddDate(0, 0, -1)) {
		habit.Streak++
	} else {
		habit.Streak = 1
	}
	habit.LastCheckedAt = now
	habit.UpdatedAt = now
	return habit
}

func sameUTCDate(a, b time.Time) bool {
	if a.IsZero() || b.IsZero() {
		return false
	}
	aa := a.UTC()
	bb := b.UTC()
	return aa.Year() == bb.Year() && aa.YearDay() == bb.YearDay()
}
