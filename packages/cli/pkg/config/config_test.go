package config_test

import (
	"testing"

	"github.com/konflux-ci/kite/packages/cli/pkg/config"
)

func TestDefaultAPIURL(t *testing.T) {
	if config.DefaultAPIURL == "" {
		t.Fatal("expected non-empty default API URL")
	}
}
