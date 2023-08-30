//go:build integration

package storage

import (
	"assignment/domain"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest"
)

var pgPool *pgxpool.Pool

func TestMain(m *testing.M) {
	d, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct Docker connection pool: %v", err)
	}

	if err := d.Client.Ping(); err != nil {
		log.Fatalf("Could not connect to Docker: %v", err)
	}

	pg, err := d.Run("postgres", "15.4", []string{"POSTGRES_PASSWORD=test"})
	if err != nil {
		log.Fatalf("Could not run postgres: %v", err)
	}

	if err := d.Retry(func() (err error) {
		pgPool, err = pgxpool.New(
			context.Background(),
			fmt.Sprintf("host=localhost port=%s user=postgres password=test", pg.GetPort("5432/tcp")),
		)
		if err != nil {
			return err
		}

		return pgPool.Ping(context.Background())
	}); err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	code := m.Run()

	if err := d.Purge(pg); err != nil {
		log.Printf("Could not purge postgres: %v", err)
	}

	os.Exit(code)
}

func TestSql_CreateSegment(t *testing.T) {
	ctx := context.Background()
	storage := NewSqlStorage(pgPool)

	if err := storage.InitDb(ctx); err != nil {
		t.Fatalf("Could not init database: %v", err)
	}
	t.Cleanup(func() {
		_, err := pgPool.Exec(ctx, "DELETE FROM segment;")
		if err != nil {
			t.Errorf("Failed to cleanup table \"segment\": %v", err)
		}
	})

	t.Run("given empty database create new segment successfully", func(t *testing.T) {
		if err := storage.CreateSegment(ctx, "TEST_SEGMENT"); err != nil {
			t.Fatalf("Expected to create new segment in an empty database, but got error: %v", err)
		}

		dbRes, err := pgPool.Query(ctx, "SELECT * FROM segment;")
		if err != nil {
			t.Fatalf("Could not query segment from database: %v", err)
		}

		rows, err := pgx.CollectRows(dbRes, pgx.RowToMap)
		if err != nil {
			t.Fatalf("Could not scan segments from database: %v", err)
		}

		if len(rows) == 0 {
			t.Fatalf("Expected to query newly created segment, but got error: %v", err)
		}

		if len(rows) > 1 {
			t.Errorf(
				"Expected empty database to contain only one row after CreateSegment, but got %d",
				len(rows),
			)
		}

		row := rows[0]
		segment, ok := row["name"]
		if !ok {
			t.Errorf("Expected segment to contain field \"name\"")
		} else if segment != "TEST_SEGMENT" {
			t.Errorf(
				"Expected newly created segment's names to match\n"+
					"\tExpected=\"TEST_SEGMENT\"\n"+
					"\tGot=\"%s\"",
				segment,
			)
		}

		if t.Failed() {
			segments, err := json.MarshalIndent(rows, "\t", "  ")
			if err != nil {
				t.Fatalf("Could not marshal the segments that we've got: %v", err)
			}

			t.Logf("Got segments: %s", string(segments))
		}
	})

	t.Run("given database with a single segment, when creating segment with the same name, expect error", func(t *testing.T) {
		err := storage.CreateSegment(ctx, "TEST_SEGMENT")
		if err == nil {
			t.Fatal("Expected to have an error, but got nil")
		}

		if !errors.Is(err, domain.ErrSegmentAlreadyExists) {
			t.Errorf(
				"Expected to have error domain.ErrSegmentAlreadyExists, but instead got:\n"+
					"\tType=%[1]T,\n"+
					"\tErr=\"%[1]s\"",
				err,
			)
		}
	})
}

func TestSql_DeleteSegment(t *testing.T) {
	ctx := context.Background()
	storage := NewSqlStorage(pgPool)

	if err := storage.InitDb(ctx); err != nil {
		t.Fatalf("Could not init database: %v", err)
	}

	t.Run("given database with one test segment", func(t *testing.T) {
		if err := storage.CreateSegment(ctx, "TEST_SEGMENT"); err != nil {
			t.Fatalf("Could not create test segment: %v", err)
		}

		if err := storage.DeleteSegment(ctx, "TEST_SEGMENT"); err != nil {
			t.Fatalf("Expected to delete test segment, but got error: %v", err)
		}

		dbRes, err := pgPool.Query(ctx, "SELECT * FROM segment;")
		if err != nil {
			t.Fatalf("Could not query segment from database: %v", err)
		}

		rows, err := pgx.CollectRows(dbRes, pgx.RowToMap)
		if err != nil {
			t.Fatalf("Could not scan segments from database: %v", err)
		}

		if len(rows) >= 1 {
			t.Errorf(
				"Expected empty database, but got %d",
				len(rows),
			)
		}

		if t.Failed() {
			segments, err := json.MarshalIndent(rows, "\t", "  ")
			if err != nil {
				t.Fatalf("Could not marshal the segments that we've got: %v", err)
			}

			t.Logf("Got segments: %s", string(segments))
		}
	})

	t.Run("given database with test segment, when deleting a different segment, expect to not delete the test segment",
		func(t *testing.T) {
			if err := storage.CreateSegment(ctx, "TEST_SEGMENT"); err != nil {
				t.Fatalf("Could not create test segment: %v", err)
			}

			err := storage.DeleteSegment(ctx, "DIFFERENT_SEGMENT")
			if err == nil {
				t.Errorf("Expected to have an error, but got nil")
			} else if !errors.Is(err, domain.ErrSegmentNotFound) {
				t.Errorf(
					"Expected to have error domain.ErrSegmentNotFound, but instead got:\n"+
						"\tType=%[1]T,\n"+
						"\tErr=\"%[1]s\"",
					err,
				)
			}

			dbRes, err := pgPool.Query(ctx, "SELECT * FROM segment;")
			if err != nil {
				t.Fatalf("Could not query segment from database: %v", err)
			}

			rows, err := pgx.CollectRows(dbRes, pgx.RowToMap)
			if err != nil {
				t.Fatalf("Could not scan segments from database: %v", err)
			}

			if len(rows) == 0 {
				t.Fatal(
					"Expected original row to remain intact, but got no segments from db",
				)
			}

			if len(rows) > 1 {
				t.Errorf(
					"Expected no new rows in database, but got %d",
					len(rows),
				)
			}

			row := rows[0]
			segment, ok := row["name"]
			if !ok {
				t.Errorf("Expected segment to contain field \"name\"")
			} else if segment != "TEST_SEGMENT" {
				t.Errorf(
					"Expected original row to remain its name\n"+
						"\tExpected=\"TEST_SEGMENT\"\n"+
						"\tGot=\"%s\"",
					segment,
				)
			}

			if t.Failed() {
				segments, err := json.MarshalIndent(rows, "\t", "  ")
				if err != nil {
					t.Fatalf("Could not marshal the segments that we've got: %v", err)
				}

				t.Logf("Got segments: %s", string(segments))
			}
		})
}

func TestSql_AddUserToSegment(t *testing.T) {
	ctx := context.Background()
	storage := NewSqlStorage(pgPool)

	if err := storage.InitDb(ctx); err != nil {
		t.Fatalf("Could not init database: %v", err)
	}

	t.Run("given database with one test segment, when adding this test segment to user id 1000", func(t *testing.T) {
		_, err := pgPool.Exec(ctx, "DELETE FROM users_in_segment; DELETE FROM segment;")
		if err != nil {
			t.Fatalf("Could not clean database: %v", err)
		}

		if err := storage.CreateSegment(ctx, "TEST_SEGMENT"); err != nil {
			t.Fatalf("Could not create test segment: %v", err)
		}

		if err := storage.AddUserToSegment(ctx, 1000, []string{"TEST_SEGMENT"}); err != nil {
			t.Fatalf("Expected to add test segment to user, but got error: %v", err)
		}

		dbRes, err := pgPool.Query(ctx, "SELECT * FROM users_in_segment;")
		if err != nil {
			t.Fatalf("Could not query segment from database: %v", err)
		}

		rows, err := pgx.CollectRows(dbRes, pgx.RowToMap)
		if err != nil {
			t.Fatalf("Could not scan segments from database: %v", err)
		}

		if len(rows) == 0 {
			t.Fatal(
				"Expected user to have test segment, but user got no segments",
			)
		}

		if len(rows) > 1 {
			t.Errorf(
				"Expected to have only one row, but have %d",
				len(rows),
			)
		}

		row := rows[0]
		segment, ok := row["segment"]
		if !ok {
			t.Errorf("Expected segment to contain field \"name\"")
		} else if segment != "TEST_SEGMENT" {
			t.Errorf(
				"Expected original row to remain its name\n"+
					"\tExpected=\"TEST_SEGMENT\"\n"+
					"\tGot=\"%s\"",
				segment,
			)
		}

		if t.Failed() {
			segments, err := json.MarshalIndent(rows, "\t", "  ")
			if err != nil {
				t.Fatalf("Could not marshal the segments that we've got: %v", err)
			}

			t.Logf("Got segments: %s", string(segments))
		}
	})

	t.Run("given database with test segment, when adding a different segment to a user with id=1000, expect to have ErrSegmentNotFound",
		func(t *testing.T) {
			_, err := pgPool.Exec(ctx, "DELETE FROM users_in_segment; DELETE FROM segment;")
			if err != nil {
				t.Fatalf("Could not clean database: %v", err)
			}

			if err := storage.CreateSegment(ctx, "TEST_SEGMENT"); err != nil {
				t.Fatalf("Could not create test segment: %v", err)
			}

			err = storage.AddUserToSegment(ctx, 1000, []string{"DIFFERENT_SEGMENT"})
			if err == nil {
				t.Errorf("Expected to have an error, but got nil")
			} else if !errors.Is(err, domain.ErrSegmentNotFound) {
				t.Errorf(
					"Expected to have error domain.ErrSegmentNotFound, but instead got:\n"+
						"\tType=%[1]T,\n"+
						"\tErr=\"%[1]s\"",
					err,
				)
			}

			dbRes, err := pgPool.Query(ctx, "SELECT * FROM users_in_segment;")
			if err != nil {
				t.Fatalf("Could not query segment from database: %v", err)
			}

			rows, err := pgx.CollectRows(dbRes, pgx.RowToMap)
			if err != nil {
				t.Fatalf("Could not scan segments from database: %v", err)
			}

			if len(rows) >= 1 {
				t.Errorf(
					"Expected no new rows in database, but got %d",
					len(rows),
				)
			}

			if t.Failed() {
				segments, err := json.MarshalIndent(rows, "\t", "  ")
				if err != nil {
					t.Fatalf("Could not marshal the segments that we've got: %v", err)
				}

				t.Logf("Got segments: %s", string(segments))
			}
		})

	t.Run("given database with test segment and user with id 1000 has test segment, when adding a test segment to a user, expect to haveErrUserIsAlreadyHasThisSegment",
		func(t *testing.T) {
			_, err := pgPool.Exec(ctx, "DELETE FROM users_in_segment; DELETE FROM segment;")
			if err != nil {
				t.Fatalf("Could not clean database: %v", err)
			}

			if err := storage.CreateSegment(ctx, "TEST_SEGMENT"); err != nil {
				t.Fatalf("Could not create test segment: %v", err)
			}

			if err := storage.AddUserToSegment(ctx, 1000, []string{"TEST_SEGMENT"}); err != nil {
				t.Fatalf("Could not add test segment to a user: %v", err)
			}

			err = storage.AddUserToSegment(ctx, 1000, []string{"TEST_SEGMENT"})
			if err == nil {
				t.Errorf("Expected to have an error, but got nil")
			} else if !errors.Is(err, domain.ErrUserIsAlreadyHasThisSegment) {
				t.Errorf(
					"Expected to have error domain.ErrUserIsAlreadyHasThisSegment, but instead got:\n"+
						"\tType=%[1]T,\n"+
						"\tErr=\"%[1]s\"",
					err,
				)
			}

			dbRes, err := pgPool.Query(ctx, "SELECT * FROM users_in_segment;")
			if err != nil {
				t.Fatalf("Could not query segment from database: %v", err)
			}

			rows, err := pgx.CollectRows(dbRes, pgx.RowToMap)
			if err != nil {
				t.Fatalf("Could not scan segments from database: %v", err)
			}

			if len(rows) == 0 {
				t.Fatal(
					"Expected user to have test segment, but user got no segments",
				)
			}

			if len(rows) > 1 {
				t.Errorf(
					"Expected no new rows in database, but got %d",
					len(rows),
				)
			}

			row := rows[0]
			segment, ok := row["segment"]
			if !ok {
				t.Errorf("Expected segment to contain field \"name\"")
			} else if segment != "TEST_SEGMENT" {
				t.Errorf(
					"Expected original row to remain its name\n"+
						"\tExpected=\"TEST_SEGMENT\"\n"+
						"\tGot=\"%s\"",
					segment,
				)
			}

			if t.Failed() {
				segments, err := json.MarshalIndent(rows, "\t", "  ")
				if err != nil {
					t.Fatalf("Could not marshal the segments that we've got: %v", err)
				}

				t.Logf("Got segments: %s", string(segments))
			}
		})
}

func TestSql_DeleteUserFromSegment(t *testing.T) {
	ctx := context.Background()
	storage := NewSqlStorage(pgPool)

	if err := storage.InitDb(ctx); err != nil {
		t.Fatalf("Could not init database: %v", err)
	}

	t.Run("given database with one test segment and user has this segment, when deleting this test segment from the user", func(t *testing.T) {
		_, err := pgPool.Exec(ctx, "DELETE FROM users_in_segment; DELETE FROM segment;")
		if err != nil {
			t.Fatalf("Could not clean database: %v", err)
		}

		if err := storage.CreateSegment(ctx, "TEST_SEGMENT"); err != nil {
			t.Fatalf("Could not create test segment: %v", err)
		}

		if err := storage.AddUserToSegment(ctx, 1000, []string{"TEST_SEGMENT"}); err != nil {
			t.Fatalf("Expected to add test segment to user, but got error: %v", err)
		}

		if err := storage.DeleteUserFromSegment(ctx, 1000, []string{"TEST_SEGMENT"}); err != nil {
			t.Errorf("Expected to delete test segment to user, but got error: %v", err)
		}

		dbRes, err := pgPool.Query(ctx, "SELECT * FROM users_in_segment;")
		if err != nil {
			t.Fatalf("Could not query segment from database: %v", err)
		}

		rows, err := pgx.CollectRows(dbRes, pgx.RowToMap)
		if err != nil {
			t.Fatalf("Could not scan segments from database: %v", err)
		}

		if len(rows) >= 1 {
			t.Fatal(
				"Expected user not to have test segment, but user got some segments",
			)
		}

		if t.Failed() {
			segments, err := json.MarshalIndent(rows, "\t", "  ")
			if err != nil {
				t.Fatalf("Could not marshal the segments that we've got: %v", err)
			}

			t.Logf("Got segments: %s", string(segments))
		}
	})

	t.Run("given database with one test segment and user has this segment, when deleting different segment from the user and expecting ErrUserHaveNotThisSegment",
		func(t *testing.T) {
			_, err := pgPool.Exec(ctx, "DELETE FROM users_in_segment; DELETE FROM segment;")
			if err != nil {
				t.Fatalf("Could not clean database: %v", err)
			}

			if err := storage.CreateSegment(ctx, "TEST_SEGMENT"); err != nil {
				t.Fatalf("Could not create test segment: %v", err)
			}

			if err := storage.AddUserToSegment(ctx, 1000, []string{"TEST_SEGMENT"}); err != nil {
				t.Fatalf("Expected to add test segment to user, but got error: %v", err)
			}

			err = storage.DeleteUserFromSegment(ctx, 1000, []string{"DIFFERENT_SEGMENT"})
			if err == nil {
				t.Errorf("Expected to have an error, but got nil")
			} else if !errors.Is(err, domain.ErrUserHaveNotThisSegment) {
				t.Errorf(
					"Expected to have error domain.ErrUserHaveNotThisSegment, but instead got:\n"+
						"\tType=%[1]T,\n"+
						"\tErr=\"%[1]s\"",
					err,
				)
			}

			dbRes, err := pgPool.Query(ctx, "SELECT * FROM users_in_segment;")
			if err != nil {
				t.Fatalf("Could not query segment from database: %v", err)
			}

			rows, err := pgx.CollectRows(dbRes, pgx.RowToMap)
			if err != nil {
				t.Fatalf("Could not scan segments from database: %v", err)
			}

			if len(rows) > 1 {
				t.Errorf(
					"Expected no new rows in database, but got %d",
					len(rows),
				)
			}

			if len(rows) == 0 {
				t.Errorf(
					"Expected no to delete segment from user",
				)
			}

			if t.Failed() {
				segments, err := json.MarshalIndent(rows, "\t", "  ")
				if err != nil {
					t.Fatalf("Could not marshal the segments that we've got: %v", err)
				}

				t.Logf("Got segments: %s", string(segments))
			}
		})
}

func TestSql_GetUserSegments(t *testing.T) {
	ctx := context.Background()
	storage := NewSqlStorage(pgPool)

	if err := storage.InitDb(ctx); err != nil {
		t.Fatalf("Could not init database: %v", err)
	}

	t.Run("given database with one test segment and user has this segment, when trying to get user segments", func(t *testing.T) {
		_, err := pgPool.Exec(ctx, "DELETE FROM users_in_segment; DELETE FROM segment;")
		if err != nil {
			t.Fatalf("Could not clean database: %v", err)
		}

		if err := storage.CreateSegment(ctx, "TEST_SEGMENT"); err != nil {
			t.Fatalf("Could not create test segment: %v", err)
		}

		if err := storage.AddUserToSegment(ctx, 1000, []string{"TEST_SEGMENT"}); err != nil {
			t.Fatalf("Expected to add test segment to user, but got error: %v", err)
		}

		segments, err := storage.GetUserSegments(ctx, 1000)
		if err != nil {
			t.Errorf("Expected to get user segments, but got error: %v", err)
		}

		if len(segments) > 1 {
			t.Fatal(
				"Expected user  to have only test segment, but user got other segments",
			)
		}

		if len(segments) == 0 {
			t.Fatal(
				"Expected user to have one test segment",
			)
		}

		if t.Failed() {
			t.Logf("Got segments: %s", segments)
		}
	})
}
