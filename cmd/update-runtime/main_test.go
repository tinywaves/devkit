package main

import (
	"strings"
	"testing"
)

func TestLatest(t *testing.T) {
	got := latest(
		[]string{"v17.9.1", "v18.20.8", "v18.20.7", "20.20.1", "20.20.2", "20.21.0-rc.1", "12.1.0"},
		18,
	)
	if got := strings.Join([]string{got[0].String(), got[1].String()}, ","); got != "20.20.2,18.20.8" {
		t.Fatalf("latest versions = %q", got)
	}
}
