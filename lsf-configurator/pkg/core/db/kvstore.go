package db

import (
	"lsf-configurator/pkg/core"
	"lsf-configurator/pkg/kvstore"
)

type kvFunctionAppStore struct {
	store kvstore.Client[core.FunctionApp]
}

func NewKvFunctionAppStore() core.FunctionAppStore {
	client := kvstore.NewStore[core.FunctionApp]()
	return &kvFunctionAppStore{
		store: client,
	}
}

func (s *kvFunctionAppStore) Set(id string, app core.FunctionApp) {
	s.store.Set(id, app)
}

func (s *kvFunctionAppStore) Get(id string) (core.FunctionApp, error) {
	return s.store.Get(id)
}

func (s *kvFunctionAppStore) Delete(id string) {
	s.store.Delete(id)
}
