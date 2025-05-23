package deps

import (
	"io"
	"log/slog"
	"os"

	"github.com/alnah/fla/config"
	"github.com/alnah/fla/logger"
	"github.com/alnah/fla/storage/cache"
)

type CLIDeps struct {
	Reader        io.Reader
	Writer        io.Writer
	ErrWriter     io.Writer
	Logger        logger.Logger
	ConfigManager config.Manager
	Cache         cache.Cache
}

func New(reader io.Reader, writer io.Writer, errWriter io.Writer) (*CLIDeps, error) {
	logger := logger.NewSlogger(os.Stdout, false, slog.LevelDebug)
	manager, err := config.New(config.WithLogger(logger)).Load()
	if err != nil {
		return nil, err
	}
	cache, err := cache.NewRedisCache().
		WithAddress(manager.Cache.Address).
		WithPassword(manager.Cache.Password).Build()
	if err != nil {
		return nil, err
	}
	return &CLIDeps{
		Reader:        reader,
		Writer:        writer,
		ErrWriter:     errWriter,
		Logger:        logger,
		ConfigManager: *manager,
		Cache:         cache,
	}, nil
}
