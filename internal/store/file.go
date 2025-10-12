package store

import (
	"github.com/user/truckllm/internal/config"
)

type FileStore struct {
	path string
}

func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

func (f *FileStore) LoadConfig() (*config.Config, error) {
	return config.LoadConfig(f.path)
}

func (f *FileStore) SaveConfig(cfg *config.Config) error {
	return config.SaveConfig(cfg, f.path)
}
