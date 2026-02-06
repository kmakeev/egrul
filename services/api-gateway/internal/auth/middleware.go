package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	// UserIDKey ключ для UserID в context
	UserIDKey contextKey = "user_id"
	// EmailKey ключ для Email в context
	EmailKey contextKey = "email"
)

// Middleware создает HTTP middleware для проверки JWT токена
func (m *JWTManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// Токен не обязателен, пропускаем запросы без токена
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Проверяем формат: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		// Валидируем токен
		claims, err := m.Verify(parts[1])
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Добавляем данные пользователя в контекст
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, EmailKey, claims.Email)

		// Передаем запрос дальше с обновленным контекстом
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext извлекает UserID из контекста
func GetUserIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(UserIDKey).(string)
	return userID
}

// GetEmailFromContext извлекает Email из контекста
func GetEmailFromContext(ctx context.Context) string {
	email, _ := ctx.Value(EmailKey).(string)
	return email
}
