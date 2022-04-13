package hvc

import (
	"os"
	"testing"
)

func handleIntegrationTest(t *testing.T) {
	if v := os.Getenv("TEST_INTEGRATION"); v != "1" {
		t.Skip()
	}
}
