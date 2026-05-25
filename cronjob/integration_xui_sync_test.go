package cronjob

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/deposist/s-ui-x/database"
	"github.com/deposist/s-ui-x/database/importxui"
	"github.com/deposist/s-ui-x/database/model"
)

func TestIntegrationXUISyncRunProfileUpdatesRunFields(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		initCronJobTestDB(t)
		profile := createXUISyncProfileForTest(t, createXUISyncSourceDB(t))
		job := &XUISyncJob{now: func() time.Time { return time.Unix(1700001000, 0) }}

		if err := job.RunProfile(context.Background(), profile); err != nil {
			t.Fatal(err)
		}

		stored := loadIntegrationSyncProfile(t, profile.Id)
		if stored.LastRunStatus != "success" || stored.LastRunAt != 1700001000 {
			t.Fatalf("unexpected success run fields: %#v", stored)
		}
		var summary importxui.Report
		if err := json.Unmarshal(stored.LastRunSummary, &summary); err != nil {
			t.Fatal(err)
		}
		var audit model.AuditEvent
		if err := database.GetDB().Where("event = ?", "xui_sync_run").First(&audit).Error; err != nil {
			t.Fatal(err)
		}
	})

	t.Run("failure", func(t *testing.T) {
		initCronJobTestDB(t)
		profile := createXUISyncProfileForTest(t, filepath.Join(t.TempDir(), "missing.db"))
		job := &XUISyncJob{now: func() time.Time { return time.Unix(1700001100, 0) }}

		if err := job.RunProfile(context.Background(), profile); err == nil {
			t.Fatal("missing source should fail")
		}

		stored := loadIntegrationSyncProfile(t, profile.Id)
		if stored.LastRunStatus != "failed" || stored.LastRunAt != 1700001100 {
			t.Fatalf("unexpected failed run fields: %#v", stored)
		}
		var summary map[string]any
		if err := json.Unmarshal(stored.LastRunSummary, &summary); err != nil {
			t.Fatalf("unmarshal LastRunSummary: %v", err)
		}
		errStr, ok := summary["error"].(string)
		if !ok {
			t.Fatalf("summary error should be string, got %#v", summary["error"])
		}
		if errStr == "failed" {
			t.Fatalf("summary error should be sanitized lastErr, got generic %q", errStr)
		}
		if !strings.Contains(errStr, "missing.db") {
			t.Fatalf("summary error should reference missing source path, got %q", errStr)
		}
		if summary["errorClass"] != "source" {
			t.Fatalf("summary errorClass should be source, got %#v", summary["errorClass"])
		}
		var audit model.AuditEvent
		if err := database.GetDB().Where("event = ?", "xui_sync_failed").First(&audit).Error; err != nil {
			t.Fatal(err)
		}
	})

	t.Run("min interval", func(t *testing.T) {
		job := &XUISyncJob{now: func() time.Time { return time.Unix(1700001200, 0) }}
		profile := &model.XUISyncProfile{Id: 42, LastRunAt: time.Unix(1700001200, 0).Add(-time.Minute).Unix()}
		err := job.RunProfile(context.Background(), profile)
		if err == nil || !strings.Contains(err.Error(), "too recently") {
			t.Fatalf("expected min-interval error, got %v", err)
		}
	})
}

func loadIntegrationSyncProfile(t *testing.T, id uint) model.XUISyncProfile {
	t.Helper()
	var stored model.XUISyncProfile
	if err := database.GetDB().First(&stored, id).Error; err != nil {
		t.Fatal(err)
	}
	return stored
}
