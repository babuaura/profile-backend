package profile

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

func (s *PostgresStore) Get() (Profile, error) {
	ctx, cancel := s.context()
	defer cancel()
	var data []byte
	err := s.pool.QueryRow(ctx, `select data from profiles where id = $1`, ownerProfileID).Scan(&data)
	if err == nil {
		var profile Profile
		if err := json.Unmarshal(data, &profile); err != nil {
			return Profile{}, err
		}
		return profile, nil
	}
	if err != pgx.ErrNoRows {
		return Profile{}, err
	}
	profile := defaultProfile()
	encoded, err := json.Marshal(profile)
	if err != nil {
		return Profile{}, err
	}
	_, err = s.pool.Exec(ctx, `insert into profiles (id, data) values ($1, $2) on conflict (id) do nothing`, ownerProfileID, encoded)
	return profile, err
}

func (s *PostgresStore) context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s.timeout)
}
