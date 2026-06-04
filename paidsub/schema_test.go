package paidsub

import (
	"strings"
	"testing"

	"github.com/deposist/s-ui-x/database"

	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	t.Setenv("SUI_DB_FOLDER", t.TempDir())
	if err := database.InitDB("file::memory:?cache=shared"); err != nil {
		if strings.Contains(err.Error(), "go-sqlite3 requires cgo") {
			t.Skip(err)
		}
		t.Fatal(err)
	}
	db := database.GetDB()
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

func TestEnsureSchemaIdempotent(t *testing.T) {
	db := openTestDB(t)
	if err := EnsureSchema(db); err != nil {
		t.Fatalf("EnsureSchema first run: %v", err)
	}
	// Second run must be a no-op (all statements use IF NOT EXISTS).
	if err := EnsureSchema(db); err != nil {
		t.Fatalf("EnsureSchema second run: %v", err)
	}
	for _, table := range []string{"paidsub_bindings", "tariffs", "payment_orders"} {
		if !db.Migrator().HasTable(table) {
			t.Fatalf("table %q missing after EnsureSchema", table)
		}
	}
}

func TestBindingUniqueness(t *testing.T) {
	db := openTestDB(t)
	if err := EnsureSchema(db); err != nil {
		t.Fatalf("EnsureSchema: %v", err)
	}
	if err := db.Create(&Binding{ClientId: 1, TgUserId: 1000}).Error; err != nil {
		t.Fatalf("first binding: %v", err)
	}
	// Same tg id, different client → must violate the UNIQUE(tg_user_id) index.
	if err := db.Create(&Binding{ClientId: 2, TgUserId: 1000}).Error; err == nil {
		t.Fatal("expected duplicate tg_user_id to be rejected")
	}
	// Same client, different tg id → must violate the UNIQUE(client_id) index.
	if err := db.Create(&Binding{ClientId: 1, TgUserId: 2000}).Error; err == nil {
		t.Fatal("expected duplicate client_id to be rejected")
	}
}

func TestSetBindingReleasesPrevious(t *testing.T) {
	db := openTestDB(t)
	if err := EnsureSchema(db); err != nil {
		t.Fatalf("EnsureSchema: %v", err)
	}
	svc := NewService()
	if err := svc.SetBinding(1, 1000); err != nil {
		t.Fatalf("SetBinding 1->1000: %v", err)
	}
	// Rebind the same tg id to a different client: old row must be released.
	if err := svc.SetBinding(2, 1000); err != nil {
		t.Fatalf("SetBinding 2->1000: %v", err)
	}
	var count int64
	db.Model(&Binding{}).Where("tg_user_id = ?", 1000).Count(&count)
	if count != 1 {
		t.Fatalf("expected exactly 1 binding for tg 1000, got %d", count)
	}
	if _, err := svc.BindingForClient(1); err == nil {
		t.Fatal("expected client 1 binding to be released")
	}
	b, err := svc.BindingForClient(2)
	if err != nil {
		t.Fatalf("BindingForClient(2): %v", err)
	}
	if b.TgUserId != 1000 {
		t.Fatalf("expected tg 1000 bound to client 2, got %d", b.TgUserId)
	}
}
