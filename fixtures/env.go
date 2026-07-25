package fixtures

import (
	"async/config"
	"async/store"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/require"
)

type TestEnv struct {
	Config *config.Config
	DB     *sql.DB
}

func NewTestEnv(t *testing.T) *TestEnv {
	os.Setenv("ENV", string(config.Env_Test))
	conf, err := config.New()
	require.NoError(t, err)

	db, err := store.NewPostgresDB(conf)
	require.NoError(t, err)

	return &TestEnv{
		Config: conf,
		DB:     db,
	}
}

func (te *TestEnv) SetupDB(t *testing.T) func(t *testing.T) {
	m, err := migrate.New(
		"file:///migrations",
		te.Config.DatabaseUrl(),
	)
	require.NoError(t, err)
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}

	return te.TeardownDB
}

func (te *TestEnv) TeardownDB(t *testing.T) {
	_, err := te.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s", strings.Join([]string{"users", "refresh_token", "reports"}, ", ")))
	require.NoError(t, err)
}
