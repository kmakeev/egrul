package graph

// Простой DataLoader для устранения повторных запросов (per-request cache).

import (
	"context"
	"fmt"
	"net/http"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
)

type ctxKey string

const foundersLoaderKey ctxKey = "foundersLoader"

// FoundersLoader кэширует учредителей компании в рамках одного запроса.
type FoundersLoader struct {
	service *Resolver
	cache   map[string][]*model.Founder
}

// newFoundersLoader создаёт новый loader.
func newFoundersLoader(resolver *Resolver) *FoundersLoader {
	return &FoundersLoader{
		service: resolver,
		cache:   make(map[string][]*model.Founder),
	}
}

// Load загружает учредителей с кэшированием по ключу (ogrn+limit+offset).
func (l *FoundersLoader) Load(ctx context.Context, ogrn string, limit, offset int) ([]*model.Founder, error) {
	key := fmt.Sprintf("%s:%d:%d", ogrn, limit, offset)
	if founders, ok := l.cache[key]; ok {
		return founders, nil
	}

	founders, err := l.service.CompanyService.GetFounders(ctx, ogrn, limit, offset)
	if err != nil {
		return nil, err
	}
	l.cache[key] = founders
	return founders, nil
}

// DataLoaderMiddleware добавляет DataLoader'ы в контекст GraphQL-запросов.
func DataLoaderMiddleware(resolver *Resolver) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Инициализируем loader-ы на каждый запрос
			ctx = context.WithValue(ctx, foundersLoaderKey, newFoundersLoader(resolver))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// foundersLoaderFromContext достаёт FoundersLoader из контекста.
func foundersLoaderFromContext(ctx context.Context) *FoundersLoader {
	if loader, ok := ctx.Value(foundersLoaderKey).(*FoundersLoader); ok {
		return loader
	}
	return nil
}


