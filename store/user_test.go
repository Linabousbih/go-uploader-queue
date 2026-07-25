package store_test

import (
	"async/fixtures"
	"async/store"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserStore(t *testing.T) {
	env := fixtures.NewTestEnv(t)
	cleanup := env.SetupDB(t)
	t.Cleanup(func() {
		cleanup(t)
	})

	ctx := context.Background()
	userStore := store.NewUserStore(env.DB)
	user, err := userStore.CreateUser(context.Background(), "test@test.com", "testpwd")
	require.NoError(t, err)
	require.Equal(t, "test@test.com", user.Email)
	require.NoError(t, user.ComparePassword("testpwd"))
	user2, err := userStore.ById(ctx, user.Id)

	require.NoError(t, err)
	require.Equal(t, user.Email, user2.Email)
}
