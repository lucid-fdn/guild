package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/lucid-fdn/guild/server/internal/artifacts"
	"github.com/lucid-fdn/guild/server/internal/bootstrap"
	"github.com/lucid-fdn/guild/server/internal/config"
	"github.com/lucid-fdn/guild/server/internal/dri"
	"github.com/lucid-fdn/guild/server/internal/evaluator"
	"github.com/lucid-fdn/guild/server/internal/httpapi"
	"github.com/lucid-fdn/guild/server/internal/institution"
	"github.com/lucid-fdn/guild/server/internal/promotions"
	"github.com/lucid-fdn/guild/server/internal/storage"
	"github.com/lucid-fdn/guild/server/internal/tasks"
)

type App struct {
	cfg         config.Config
	tasks       *tasks.Service
	dri         *dri.Service
	artifacts   *artifacts.Service
	promotions  *promotions.Service
	evaluator   *evaluator.Service
	institution *institution.Service
	router      http.Handler
	cancel      context.CancelFunc
}

func New(cfg config.Config) (*App, error) {
	store, err := newStore(cfg)
	if err != nil {
		return nil, err
	}
	if err := bootstrap.SeedIfEmpty(store); err != nil {
		return nil, err
	}

	taskService := tasks.NewService(store)
	driService := dri.NewService(store)
	objectDir := cfg.ObjectDir
	if objectDir == "" {
		objectDir = filepath.Join(cfg.DataDir, "objects")
	}
	objectStore, err := artifacts.NewLocalObjectStore(objectDir)
	if err != nil {
		return nil, err
	}
	artifactService := artifacts.NewService(store, objectStore)
	promotionService := promotions.NewService(store)
	institutionService := institution.NewService(store)
	evaluatorService := evaluator.NewService(store, taskService, driService, artifactService, promotionService)
	ctx, cancel := context.WithCancel(context.Background())
	if cfg.WorkerEnabled {
		go evaluator.NewWorker(evaluatorService, cfg.WorkerInterval, slog.Default()).Start(ctx)
	}

	return &App{
		cfg:         cfg,
		tasks:       taskService,
		dri:         driService,
		artifacts:   artifactService,
		promotions:  promotionService,
		evaluator:   evaluatorService,
		institution: institutionService,
		cancel:      cancel,
		router: httpapi.NewRouter(httpapi.RouterDeps{
			Config:      cfg,
			Tasks:       taskService,
			DRI:         driService,
			Artifacts:   artifactService,
			Promotions:  promotionService,
			Evaluator:   evaluatorService,
			Institution: institutionService,
		}),
	}, nil
}

func (a *App) Router() http.Handler {
	return a.router
}

func (a *App) Close() {
	if a.cancel != nil {
		a.cancel()
	}
}

func newStore(cfg config.Config) (storage.Store, error) {
	switch cfg.StorageDriver {
	case "", "file":
		return storage.NewFileStore(cfg.DataDir)
	case "postgres":
		store, err := storage.NewPostgresStore(cfg.DatabaseURL)
		if err != nil {
			return nil, err
		}
		if err := store.RunMigrations(cfg.MigrationsDir); err != nil {
			_ = store.Close()
			return nil, err
		}
		return store, nil
	default:
		return nil, fmt.Errorf("unsupported GUILD_STORAGE_DRIVER %q", cfg.StorageDriver)
	}
}
