package show

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/spf13/afero"
)

type Show struct {
	mu   sync.RWMutex
	init bool
	path string
	db   *badger.DB
	fs   afero.Fs
	ctx  context.Context
	cncl context.CancelFunc
}

func (s *Show) NewShow(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Build File system location
	err := os.Mkdir(path, 0755)
	if err != nil {
		return fmt.Errorf("failed to create new show on disk: %w", err)
	}

	// Build Database location
	dbopts := badger.DefaultOptions(path + "/db")
	db, err := badger.Open(dbopts)
	if err != nil {
		return fmt.Errorf("failed to create new show database: %w", err)
	}

	if s.init {
		s.cncl()
		s.db.Close()
	}

	// Set Values for Passing
	s.fs = afero.NewBasePathFs(afero.NewOsFs(), path)
	s.db = db
	s.init = true
	s.path = path
	s.ctx, s.cncl = context.WithCancel(context.Background())

	return nil
}

func (s *Show) Load(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Load FS
	_, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("show path does not exist: %w", err)
	}

	// Load DB
	dbopts := badger.DefaultOptions(path + "/db")
	db, err := badger.Open(dbopts)
	if err != nil {
		return fmt.Errorf("failed to load show database: %w", err)
	}

	if s.init {
		s.cncl()
		s.db.Close()
	}

	// Set Values for Passing
	s.fs = afero.NewBasePathFs(afero.NewOsFs(), path)
	s.db = db
	s.path = path
	s.init = true
	s.ctx, s.cncl = context.WithCancel(context.Background())

	return nil
}

func (s *Show) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.init {
		s.cncl()
		s.db.Close()
		s.init = false
	}

	return nil
}

func (s *Show) IsInitialized() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.init
}

func (s *Show) Ctx() (context.Context, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.init {
		return nil, errors.New("show not initialized")
	}
	return s.ctx, nil
}

func (s *Show) GetDatabase() (*badger.DB, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.init {
		return nil, errors.New("show not initialized")
	}
	return s.db, nil
}

func (s *Show) GetFileSystem() (afero.Fs, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.init {
		return nil, errors.New("show not initialized")
	}
	return s.fs, nil
}
