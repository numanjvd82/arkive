package migrations

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type migration struct {
	version int64
	path    string
	name    string
}

// Run applies all pending *.up.sql migrations in order and records them.
func Run(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	migrations, err := loadMigrations(dir)
	if err != nil {
		return err
	}
	if len(migrations) == 0 {
		return nil
	}

	if err := ensureSchemaMigrations(ctx, pool); err != nil {
		return err
	}

	for _, m := range migrations {
		applied, err := isApplied(ctx, pool, m.version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		sqlBytes, err := os.ReadFile(m.path)
		if err != nil {
			return err
		}

		fmt.Printf("running %s\n", m.name)
		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			return err
		}
		if err := markApplied(ctx, pool, m.version); err != nil {
			return err
		}
		fmt.Printf("done %s\n", m.name)
	}

	return nil
}

func loadMigrations(dir string) ([]migration, error) {
	var migrations []migration
	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			return nil
		}
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 || parts[0] == "" {
			return fmt.Errorf("invalid migration name: %s", name)
		}
		version, parseErr := strconv.ParseInt(parts[0], 10, 64)
		if parseErr != nil {
			return fmt.Errorf("invalid migration version in %s: %w", name, parseErr)
		}
		migrations = append(migrations, migration{version: version, path: path, name: name})
		return nil
	}

	if err := filepath.WalkDir(dir, walkFn); err != nil {
		return nil, err
	}

	sort.Slice(migrations, func(i, j int) bool {
		if migrations[i].version == migrations[j].version {
			return migrations[i].name < migrations[j].name
		}
		return migrations[i].version < migrations[j].version
	})

	return migrations, nil
}

func ensureSchemaMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations (version bigint PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now());")
	return err
}

func isApplied(ctx context.Context, pool *pgxpool.Pool, version int64) (bool, error) {
	var exists bool
	row := pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)", version)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func markApplied(ctx context.Context, pool *pgxpool.Pool, version int64) error {
	_, err := pool.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version)
	return err
}
