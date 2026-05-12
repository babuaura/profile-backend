package personal

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool    *pgxpool.Pool
	timeout time.Duration
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool, timeout: 5 * time.Second}
}

func (s *PostgresStore) Summary() (State, error) {
	notes, err := s.ListNotes()
	if err != nil {
		return State{}, err
	}
	reminders, err := s.ListReminders()
	if err != nil {
		return State{}, err
	}
	transactions, err := s.ListTransactions()
	if err != nil {
		return State{}, err
	}
	habits, err := s.ListHabits()
	if err != nil {
		return State{}, err
	}
	return State{Notes: notes, Reminders: reminders, Transactions: transactions, Habits: habits}, nil
}

func (s *PostgresStore) ListNotes() ([]Note, error) {
	ctx, cancel := s.context()
	defer cancel()
	rows, err := s.pool.Query(ctx, `select id, title, body, tags, pinned, created_at, updated_at from personal_notes order by pinned desc, updated_at desc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Note, error) {
		var note Note
		var tags []byte
		if err := row.Scan(&note.ID, &note.Title, &note.Body, &tags, &note.Pinned, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return Note{}, err
		}
		_ = json.Unmarshal(tags, &note.Tags)
		return note, nil
	})
}

func (s *PostgresStore) CreateNote(note Note) (Note, error) {
	ctx, cancel := s.context()
	defer cancel()
	tags, _ := json.Marshal(note.Tags)
	_, err := s.pool.Exec(ctx, `insert into personal_notes (id, title, body, tags, pinned, created_at, updated_at) values ($1, $2, $3, $4, $5, $6, $7)`,
		note.ID, note.Title, note.Body, tags, note.Pinned, note.CreatedAt, note.UpdatedAt)
	return note, err
}

func (s *PostgresStore) DeleteNote(id string) (bool, error) {
	return s.deleteByID(`delete from personal_notes where id = $1`, id)
}

func (s *PostgresStore) ListReminders() ([]Reminder, error) {
	ctx, cancel := s.context()
	defer cancel()
	rows, err := s.pool.Query(ctx, `select id, title, notes, due_at, done, created_at, updated_at from personal_reminders order by done asc, due_at asc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByPos[Reminder])
}

func (s *PostgresStore) CreateReminder(reminder Reminder) (Reminder, error) {
	ctx, cancel := s.context()
	defer cancel()
	_, err := s.pool.Exec(ctx, `insert into personal_reminders (id, title, notes, due_at, done, created_at, updated_at) values ($1, $2, $3, $4, $5, $6, $7)`,
		reminder.ID, reminder.Title, reminder.Notes, reminder.DueAt, reminder.Done, reminder.CreatedAt, reminder.UpdatedAt)
	return reminder, err
}

func (s *PostgresStore) ToggleReminder(id string) (Reminder, bool, error) {
	ctx, cancel := s.context()
	defer cancel()
	row := s.pool.QueryRow(ctx, `update personal_reminders set done = not done, updated_at = $2 where id = $1 returning id, title, notes, due_at, done, created_at, updated_at`, id, time.Now().UTC())
	var reminder Reminder
	if err := row.Scan(&reminder.ID, &reminder.Title, &reminder.Notes, &reminder.DueAt, &reminder.Done, &reminder.CreatedAt, &reminder.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return Reminder{}, false, nil
		}
		return Reminder{}, false, err
	}
	return reminder, true, nil
}

func (s *PostgresStore) DeleteReminder(id string) (bool, error) {
	return s.deleteByID(`delete from personal_reminders where id = $1`, id)
}

func (s *PostgresStore) ListTransactions() ([]Transaction, error) {
	ctx, cancel := s.context()
	defer cancel()
	rows, err := s.pool.Query(ctx, `select id, type, amount, category, note, occurred_at, created_at, updated_at from personal_transactions order by occurred_at desc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByPos[Transaction])
}

func (s *PostgresStore) CreateTransaction(transaction Transaction) (Transaction, error) {
	ctx, cancel := s.context()
	defer cancel()
	_, err := s.pool.Exec(ctx, `insert into personal_transactions (id, type, amount, category, note, occurred_at, created_at, updated_at) values ($1, $2, $3, $4, $5, $6, $7, $8)`,
		transaction.ID, transaction.Type, transaction.Amount, transaction.Category, transaction.Note, transaction.OccurredAt, transaction.CreatedAt, transaction.UpdatedAt)
	return transaction, err
}

func (s *PostgresStore) DeleteTransaction(id string) (bool, error) {
	return s.deleteByID(`delete from personal_transactions where id = $1`, id)
}

func (s *PostgresStore) ListHabits() ([]Habit, error) {
	ctx, cancel := s.context()
	defer cancel()
	rows, err := s.pool.Query(ctx, `select id, name, target, frequency, streak, last_checked_at, created_at, updated_at from personal_habits order by streak desc, updated_at desc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByPos[Habit])
}

func (s *PostgresStore) CreateHabit(habit Habit) (Habit, error) {
	ctx, cancel := s.context()
	defer cancel()
	_, err := s.pool.Exec(ctx, `insert into personal_habits (id, name, target, frequency, streak, last_checked_at, created_at, updated_at) values ($1, $2, $3, $4, $5, $6, $7, $8)`,
		habit.ID, habit.Name, habit.Target, habit.Frequency, habit.Streak, habit.LastCheckedAt, habit.CreatedAt, habit.UpdatedAt)
	return habit, err
}

func (s *PostgresStore) CheckInHabit(id string) (Habit, bool, error) {
	ctx, cancel := s.context()
	defer cancel()
	row := s.pool.QueryRow(ctx, `select id, name, target, frequency, streak, last_checked_at, created_at, updated_at from personal_habits where id = $1`, id)
	var habit Habit
	if err := row.Scan(&habit.ID, &habit.Name, &habit.Target, &habit.Frequency, &habit.Streak, &habit.LastCheckedAt, &habit.CreatedAt, &habit.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return Habit{}, false, nil
		}
		return Habit{}, false, err
	}
	habit = checkedInHabit(habit, time.Now().UTC())
	_, err := s.pool.Exec(ctx, `update personal_habits set streak = $2, last_checked_at = $3, updated_at = $4 where id = $1`,
		habit.ID, habit.Streak, habit.LastCheckedAt, habit.UpdatedAt)
	return habit, true, err
}

func (s *PostgresStore) DeleteHabit(id string) (bool, error) {
	return s.deleteByID(`delete from personal_habits where id = $1`, id)
}

func (s *PostgresStore) deleteByID(query string, id string) (bool, error) {
	ctx, cancel := s.context()
	defer cancel()
	result, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() > 0, nil
}

func (s *PostgresStore) context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s.timeout)
}
