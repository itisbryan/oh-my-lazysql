package ui

import "testing"

func TestConnectionFormPostgresSSLCheckboxBuildsRequireParam(t *testing.T) {
	model := NewConnectionFormModel(nil)
	model.provider = "PostgreSQL"
	model.connectionMode = connectionModeFields
	model.name = "prod"
	model.hostname = "db.example.com"
	model.port = "5432"
	model.username = "payflow"
	model.password = "secret"
	model.database = "payflow_production"
	model.sslEnabled = true

	conn := model.buildConnection()
	if conn.URLParams != "?sslmode=require" {
		t.Fatalf("expected sslmode=require URLParams, got %q", conn.URLParams)
	}
}

func TestConnectionFormPostgresSSLCheckboxUpdatesPastedURL(t *testing.T) {
	model := NewConnectionFormModel(nil)
	model.provider = "PostgreSQL"
	model.connectionMode = connectionModeURL
	model.url = "postgres://payflow:secret@db.example.com:5432/payflow_production?connect_timeout=5"
	model.sslEnabled = true

	conn := model.buildConnection()
	if conn.URL != "postgres://payflow:secret@db.example.com:5432/payflow_production?connect_timeout=5&sslmode=require" {
		t.Fatalf("expected pasted URL to include sslmode=require, got %q", conn.URL)
	}
}

func TestConnectionFormSyncURLFieldsSetsSSLCheckbox(t *testing.T) {
	model := NewConnectionFormModel(nil)
	model.connectionMode = connectionModeURL
	model.url = "postgres://payflow:secret@db.example.com:5432/payflow_production?sslmode=require"

	model.syncURLFields()

	if !model.sslEnabled {
		t.Fatal("expected ssl checkbox to be enabled when pasted URL has sslmode=require")
	}
}
