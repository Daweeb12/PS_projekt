package storage

import "sync"

// generic Lockable type to user for data sources
type Lockable[T any] struct {
	Mu   sync.Mutex
	Data T
}

func NewLockable[T any](value T) *Lockable[T] {
	return &Lockable[T]{Data: value}
}

func (l *Lockable[T]) Get() T {
	l.Mu.Lock()
	defer l.Mu.Unlock()
	return l.Data
}

func (l *Lockable[T]) Set(value T) {
	l.Mu.Lock()
	defer l.Mu.Unlock()
	l.Data = value
}

type LockableMap[K comparable, V any] struct {
	Lockable *Lockable[map[K]V]
}

func NewLockableMap[k comparable, v any]() *LockableMap[k, v] {
	return &LockableMap[k, v]{Lockable: &Lockable[map[k]v]{Data: make(map[k]v), Mu: sync.Mutex{}}}
}

func (lm *LockableMap[K, V]) Put(key K, value V) {
	lm.Lockable.Mu.Lock()
	defer lm.Lockable.Mu.Unlock()
	lm.Lockable.Data[key] = value
}

func (lm *LockableMap[K, V]) Delete(key K) {
	lm.Lockable.Mu.Lock()
	defer lm.Lockable.Mu.Unlock()
	delete(lm.Lockable.Data, key)
}

func (lm *LockableMap[K, V]) Find(key K, f func(K) (V, error)) (V, error) {
	lm.Lockable.Mu.Lock()
	defer lm.Lockable.Mu.Unlock()
	return f(key)
}

func (lm *LockableMap[K, V]) GetValByKey(key K) (V, bool) {
	lm.Lockable.Mu.Lock()
	defer lm.Lockable.Mu.Unlock()
	val, ok := lm.Lockable.Data[key]
	return val, ok
}

func (lm *LockableMap[K, V]) GetAllValues() []V {
	lm.Lockable.Mu.Lock()
	defer lm.Lockable.Mu.Unlock()
	listValues := []V{}
	for _, message := range lm.Lockable.Data {
		listValues = append(listValues, message)
	}
	return listValues
}
