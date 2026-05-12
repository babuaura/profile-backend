package contact

import (
	"context"
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

func (s *PostgresStore) Save(message Message) error {
	ctx, cancel := s.context()
	defer cancel()
	_, err := s.pool.Exec(ctx, `insert into contact_messages (id, name, email, budget, message, source, status, created_at) values ($1, $2, $3, $4, $5, $6, $7, $8)`,
		message.ID, message.Name, message.Email, message.Budget, message.Message, message.Source, message.Status, message.CreatedAt)
	return err
}

func (s *PostgresStore) List() ([]Message, error) {
	ctx, cancel := s.context()
	defer cancel()
	rows, err := s.pool.Query(ctx, `select id, name, email, budget, message, source, status, created_at from contact_messages order by created_at desc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByPos[Message])
}

func (s *PostgresStore) UpdateStatus(id string, status string) (Message, bool, error) {
	ctx, cancel := s.context()
	defer cancel()
	row := s.pool.QueryRow(ctx, `update contact_messages set status = $2 where id = $1 returning id, name, email, budget, message, source, status, created_at`, id, status)
	var message Message
	if err := row.Scan(&message.ID, &message.Name, &message.Email, &message.Budget, &message.Message, &message.Source, &message.Status, &message.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return Message{}, false, nil
		}
		return Message{}, false, err
	}
	return message, true, nil
}

func (s *PostgresStore) Delete(id string) (bool, error) {
	ctx, cancel := s.context()
	defer cancel()
	result, err := s.pool.Exec(ctx, `delete from contact_messages where id = $1`, id)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() > 0, nil
}

func (s *PostgresStore) context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s.timeout)
}
