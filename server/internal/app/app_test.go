package app

import (
	"os"
	"testing"

	"github.com/guild-labs/guild/server/internal/config"
)

func TestNewReturnsStartupError(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "guild-data")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	_, err = New(config.Config{
		Addr:     ":0",
		UIOrigin: "http://localhost:3000",
		DataDir:  file.Name(),
	})

	if err == nil {
		t.Fatal("expected startup error")
	}
}

func TestNewSeedsBootstrapData(t *testing.T) {
	app, err := New(config.Config{
		Addr:     ":0",
		UIOrigin: "http://localhost:3000",
		DataDir:  t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if app.Router() == nil {
		t.Fatal("expected router to be initialized")
	}
}
