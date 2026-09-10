package main

import (
	"strings"
	"testing"
)

func TestRunRejectsUnknownFlag(t *testing.T) {
	err := run([]string{"-nope"})
	if err == nil {
		t.Fatal("expected flag error")
	}
	if !strings.Contains(err.Error(), "flag provided but not defined") && !strings.Contains(err.Error(), "nope") {
		t.Fatalf("err=%v", err)
	}
}
