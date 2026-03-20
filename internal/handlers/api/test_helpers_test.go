package api

import (
	"fmt"
	"reflect"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5"
)

// memRepo is an in-memory ResourceRepository used by handler tests.
// It is intentionally minimal but supports:
// - predictable List ordering (by generated key sort order)
// - ID generation when item.ID is empty
// - sql.ErrNoRows / pgx.ErrNoRows on missing Get/Delete/Update
type memRepo[T any] struct {
	mu     sync.Mutex
	items  map[string]*T
	nextID int
}

func newMemRepo[T any]() *memRepo[T] {
	return &memRepo[T]{items: make(map[string]*T)}
}

func (r *memRepo[T]) List(page, pageSize int, filters map[string]string) ([]T, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_ = filters // tests don't depend on filtering logic

	total := len(r.items)
	if total == 0 {
		return nil, 0, nil
	}

	keys := make([]string, 0, len(r.items))
	for k := range r.items {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	start := (page - 1) * pageSize
	if start >= total {
		return []T{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	out := make([]T, 0, end-start)
	for _, k := range keys[start:end] {
		out = append(out, *r.items[k])
	}
	return out, total, nil
}

func (r *memRepo[T]) Get(id string) (*T, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	v, ok := r.items[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	clone := *v
	return &clone, nil
}

func (r *memRepo[T]) Create(item *T) error {
	if item == nil {
		return fmt.Errorf("item is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	id := getStringField(item, "ID")
	if id == "" {
		r.nextID++
		id = fmt.Sprintf("%d", r.nextID)
		setStringField(item, "ID", id)
	}

	clone := *item
	r.items[id] = &clone
	return nil
}

func (r *memRepo[T]) Update(id string, item *T) error {
	if item == nil {
		return fmt.Errorf("item is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[id]; !ok {
		return pgx.ErrNoRows
	}

	// Keep repository key consistent with handler ID.
	setStringField(item, "ID", id)

	clone := *item
	r.items[id] = &clone
	return nil
}

func (r *memRepo[T]) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[id]; !ok {
		return pgx.ErrNoRows
	}
	delete(r.items, id)
	return nil
}

func getStringField[T any](ptr *T, fieldName string) string {
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return ""
	}
	e := v.Elem()
	if e.Kind() != reflect.Struct {
		return ""
	}
	f := e.FieldByName(fieldName)
	if !f.IsValid() || f.Kind() != reflect.String {
		return ""
	}
	return f.String()
}

func setStringField[T any](ptr *T, fieldName, val string) {
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return
	}
	e := v.Elem()
	if e.Kind() != reflect.Struct {
		return
	}
	f := e.FieldByName(fieldName)
	if !f.IsValid() || f.Kind() != reflect.String || !f.CanSet() {
		return
	}
	f.SetString(val)
}
