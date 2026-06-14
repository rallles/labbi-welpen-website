package config

import (
	"strings"
	"testing"
)

func TestConfigValidateRejectsMissingRequiredValues(t *testing.T) {
	cfg := Config{
		Neo4jUri:      "bolt://neo4j:7687",
		Neo4jUser:     "neo4j",
		Neo4jPassword: "change_me_neo4j_password",
		AdminUser:     "admin",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want missing ADMIN_PASSWORD error")
	}
	if !strings.Contains(err.Error(), "ADMIN_PASSWORD") {
		t.Fatalf("Validate() error = %q, want ADMIN_PASSWORD", err)
	}
	if strings.Contains(err.Error(), cfg.Neo4jPassword) {
		t.Fatalf("Validate() leaked secret in error: %q", err)
	}
}

func TestConfigValidateAcceptsRequiredValuesWithoutSMTP(t *testing.T) {
	cfg := Config{
		Neo4jUri:      "bolt://neo4j:7687",
		Neo4jUser:     "neo4j",
		Neo4jPassword: "change_me_neo4j_password",
		AdminUser:     "admin",
		AdminPassword: "change_me_admin_password",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}
