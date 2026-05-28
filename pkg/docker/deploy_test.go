package docker

import (
	"strings"
	"testing"

	"github.com/compose-spec/compose-go/v2/types"
)

func TestParsePublishedPort(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got, err := parsePublishedPort("")
		if err != nil || got != 0 {
			t.Fatalf("parsePublishedPort(\"\") = (%d, %v)", got, err)
		}
	})

	t.Run("valid", func(t *testing.T) {
		got, err := parsePublishedPort("8080")
		if err != nil || got != 8080 {
			t.Fatalf("parsePublishedPort(\"8080\") = (%d, %v)", got, err)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := parsePublishedPort("not-a-port")
		if err == nil {
			t.Fatal("expected error for invalid port")
		}
	})
}

func TestPublishedPortsRejectsInvalidPort(t *testing.T) {
	_, err := publishedPorts([]types.ServicePortConfig{{
		Target:    80,
		Published: "bad",
	}})
	if err == nil {
		t.Fatal("expected error for invalid published port")
	}
	if !strings.Contains(err.Error(), "invalid published port") {
		t.Fatalf("unexpected error: %v", err)
	}
}
