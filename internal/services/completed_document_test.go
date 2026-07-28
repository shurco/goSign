package services

import (
	"testing"
	"time"

	"github.com/shurco/gosign/internal/models"
)

func TestResolveReadonlyDefaults(t *testing.T) {
	completedAt := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)

	fields := []models.Field{
		{ID: "f-date", Type: models.FieldTypeDate, Readonly: true, DefaultValue: "{{date}}"},
		{ID: "f-text", Type: models.FieldTypeText, Readonly: true, DefaultValue: "Company Inc."},
		{ID: "f-editable", Type: models.FieldTypeText, Readonly: false, DefaultValue: "ignored"},
		{ID: "f-filled", Type: models.FieldTypeText, Readonly: true, DefaultValue: "default"},
		{ID: "f-no-default", Type: models.FieldTypeText, Readonly: true},
	}
	values := map[string]any{
		"f-filled": "user value",
	}

	resolveReadonlyDefaults(fields, values, &completedAt)

	if got := values["f-date"]; got != "2026-07-28" {
		t.Errorf("expected {{date}} resolved to 2026-07-28, got %v", got)
	}
	if got := values["f-text"]; got != "Company Inc." {
		t.Errorf("expected readonly text default, got %v", got)
	}
	if _, ok := values["f-editable"]; ok {
		t.Error("non-readonly field must not be auto-filled")
	}
	if got := values["f-filled"]; got != "user value" {
		t.Errorf("existing value must not be overwritten, got %v", got)
	}
	if _, ok := values["f-no-default"]; ok {
		t.Error("readonly field without default must stay empty")
	}
}
