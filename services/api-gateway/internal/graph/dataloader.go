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
const historyLoaderKey ctxKey = "historyLoader"

// FoundersLoader кэширует учредителей компании в рамках одного запроса.
type FoundersLoader struct {
	service *Resolver
	cache   map[string][]*model.Founder
}

// HistoryLoader кэширует историю изменений компании в рамках одного запроса.
type HistoryLoader struct {
	service *Resolver
	cache   map[string][]*model.HistoryRecord
}

// newFoundersLoader создаёт новый loader.
func newFoundersLoader(resolver *Resolver) *FoundersLoader {
	return &FoundersLoader{
		service: resolver,
		cache:   make(map[string][]*model.Founder),
	}
}

// newHistoryLoader создаёт новый loader.
func newHistoryLoader(resolver *Resolver) *HistoryLoader {
	return &HistoryLoader{
		service: resolver,
		cache:   make(map[string][]*model.HistoryRecord),
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

// Load загружает историю с кэшированием по ключу (ogrn+limit+offset).
func (l *HistoryLoader) Load(ctx context.Context, ogrn string, limit, offset int) ([]*model.HistoryRecord, error) {
	key := fmt.Sprintf("%s:%d:%d", ogrn, limit, offset)
	if history, ok := l.cache[key]; ok {
		return history, nil
	}

	history, err := l.service.CompanyService.GetHistory(ctx, ogrn, limit, offset)
	if err != nil {
		return nil, err
	}
	l.cache[key] = history
	return history, nil
}

// DataLoaderMiddleware добавляет DataLoader'ы в контекст GraphQL-запросов.
func DataLoaderMiddleware(resolver *Resolver) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Инициализируем loader-ы на каждый запрос
			ctx = context.WithValue(ctx, foundersLoaderKey, newFoundersLoader(resolver))
			ctx = context.WithValue(ctx, historyLoaderKey, newHistoryLoader(resolver))

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

// historyLoaderFromContext достаёт HistoryLoader из контекста.
func historyLoaderFromContext(ctx context.Context) *HistoryLoader {
	if loader, ok := ctx.Value(historyLoaderKey).(*HistoryLoader); ok {
		return loader
	}
	return nil
}


