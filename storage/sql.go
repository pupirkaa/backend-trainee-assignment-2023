package storage

import (
	"assignment/domain"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Sql struct {
	dbpool *pgxpool.Pool
}

func NewSqlStorage(dbpool *pgxpool.Pool) (sql *Sql) {
	return &Sql{dbpool: dbpool}
}

func (sql *Sql) CreateSegment(ctx context.Context, name string) error {
	query := "INSERT INTO segment (name) VALUES ($1);"

	_, err := sql.dbpool.Exec(ctx, query, name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.ConstraintName == "segment_pke" {
				return domain.ErrSegmentIsAlreadyExists
			}
		}

		return fmt.Errorf("creating segment: %v", err)
	}

	return nil
}

func (sql *Sql) DeleteSegment(ctx context.Context, name string) error {
	query := "DELETE FROM segment WHERE name = $1;"

	comTag, err := sql.dbpool.Exec(ctx, query, name)
	if comTag.RowsAffected() == 0 {
		return domain.ErrSegmentNotFound
	}
	if err != nil {
		return fmt.Errorf("deleting segment: %v", err)
	}

	return nil
}

func (sql *Sql) AddUserToSegment(ctx context.Context, user int, segments []string) error {
	batch := &pgx.Batch{}
	for i := range segments {
		batch.Queue("INSERT INTO users_in_segment (user_id, segment) VALUES ($1, $2);", user, segments[i])
	}
	b := sql.dbpool.SendBatch(ctx, batch)
	defer b.Close()

	_, err := b.Exec()
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.ConstraintName == "users_in_segment_segment_fkey" {
				return domain.ErrSegmentNotFound
			}
			if pgErr.ConstraintName == "users_in_segment_user_id_segment_key" {
				return domain.ErrUserIsAlreadyHasThisSegment
			}

		}

		return fmt.Errorf("adding user to segment: %v", err)
	}
	return nil
}

func (sql *Sql) DeleteUserFromSegment(ctx context.Context, user int, segments []string) error {
	batch := &pgx.Batch{}
	for i := range segments {
		batch.Queue("DELETE FROM users_in_segment WHERE user_id = $1 AND segment = $2;", user, segments[i])
	}
	b := sql.dbpool.SendBatch(ctx, batch)
	defer b.Close()

	ct, err := b.Exec()
	if ct.RowsAffected() == 0 {
		return domain.ErrUserHaveNotThisSegment
	}
	if err != nil {
		return fmt.Errorf("deleting users: %v", err)
	}

	return nil
}

func (sql *Sql) GetUserSegments(ctx context.Context, user int) ([]string, error) {
	query := "SELECT users_in_segment.segment FROM users_in_segment WHERE user_id=$1"

	rows, err := sql.dbpool.Query(ctx, query, user)
	if err != nil {
		return nil, fmt.Errorf("querying rows: %v", err)
	}
	defer rows.Close()

	segments, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return []string{}, fmt.Errorf("collecting rows: %v", err)
	}

	return segments, nil
}
