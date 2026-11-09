package logger

import (
	"os"
	"testing"
)

func TestInit_DefaultLevel(t *testing.T) {
	os.Unsetenv("LOG_LEVEL")
	if err := Init(); err != nil {
		t.Fatal(err)
	}
	if global == nil {
		t.Fatal("global logger should not be nil after Init")
	}
}

func TestInit_DebugLevel(t *testing.T) {
	os.Setenv("LOG_LEVEL", "debug")
	defer os.Unsetenv("LOG_LEVEL")
	if err := Init(); err != nil {
		t.Fatal(err)
	}
}

func TestNamed_ReturnsLogger(t *testing.T) {
	_ = Init()
	log := Named("test-module")
	if log == nil {
		t.Fatal("Named logger should not be nil")
	}
}
