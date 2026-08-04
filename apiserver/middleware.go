package apiserver

import (
	"async/store"
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func NewLoggerMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("http request", "method", r.Method, "path", r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}

type userCtxKey struct{}

func ContextWithUser(ctx context.Context, user *store.User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, user)
}

func NewAuthMiddleware(jwtManager *JwtManager, userStore *store.UserStore) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For routes that don't need any authentication
			if strings.HasPrefix(r.URL.Path, "/auth") {
				next.ServeHTTP(w, r)
				return
			}

			// authorization header is Authorization: Bearer <access_token>
			authHeader := r.Header.Get("Authorization")
			var token string
			if parts := strings.Split(authHeader, "Bearer"); len(parts) == 2 {
				token = parts[1]
			}

			parsedToken, err := jwtManager.Parse(token)
			if err != nil {
				slog.Error("failed to parse token", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if !jwtManager.IsAccessToken(parsedToken) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("not an access token"))
				return
			}
			userIdStr, err := parsedToken.Claims.GetSubject()
			if err != nil {
				slog.Error("failed to extract subject claim from token", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			userId, err := uuid.Parse(userIdStr)
			if err != nil {
				slog.Error("token subject i snot valid uuid", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			user, err := userStore.ById(r.Context(), userId)
			if err != nil {
				slog.Error("failed to get user by id", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(ContextWithUser(r.Context(), user)))
		})
	}
}
