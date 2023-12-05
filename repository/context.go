package repository

import (
	"context"
)

type (
	// ContextCancelRepository - Contextを保存するリポジトリ
	ContextCancelRepository interface {
		Save(botMessageTimeStamp string, ctx context.CancelFunc) error
		Load(botMessageTimeStamp string) (context.CancelFunc, bool)
		Delete(botMessageTimeStamp string)
	}

	inMemory struct {
		store map[string]context.CancelFunc
	}
)

func (i *inMemory) Save(botMessageTimeStamp string, ctx context.CancelFunc) error {
	i.store[botMessageTimeStamp] = ctx
	return nil
}

func (i *inMemory) Load(botMessageTimeStamp string) (context.CancelFunc, bool) {
	ctx, ok := i.store[botMessageTimeStamp]
	return ctx, ok
}

func (i *inMemory) Delete(botMessageTimeStamp string) {
	delete(i.store, botMessageTimeStamp)
}

func NewInMemoryContextRepository() ContextCancelRepository {
	return &inMemory{
		store: make(map[string]context.CancelFunc),
	}
}

var singleton ContextCancelRepository

func ProvideContextCancelRepository() ContextCancelRepository {
	if singleton == nil {
		singleton = NewInMemoryContextRepository()
	}
	return singleton
}
