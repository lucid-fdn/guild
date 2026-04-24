package dri

import (
	"strings"
	"testing"

	"github.com/lucid-fdn/guild/pkg/spec"
	"github.com/lucid-fdn/guild/server/internal/storage"
)

func TestCreateRejectsMissingTaskpack(t *testing.T) {
	store, err := storage.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(store)

	err = service.Create(spec.DriBinding{
		SchemaVersion: "v1alpha1",
		DriBindingID:  "19887415-bb68-438b-9599-0b07b5f13e97",
		TaskpackID:    "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
		Owner: spec.ActorRef{
			ActorID:     "0a3657eb-2f37-4614-a8a7-9c6bd51714a8",
			ActorType:   "agent",
			DisplayName: "payments-dri",
		},
		Status:    "assigned",
		CreatedAt: "2026-04-24T10:01:00Z",
	})

	if err == nil {
		t.Fatal("expected missing taskpack error")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected referential integrity error, got %q", err.Error())
	}
}
