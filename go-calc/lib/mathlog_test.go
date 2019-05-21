package lib

import (
	"math"
	"testing"
)

func TestMakeWithBase2(t *testing.T) {
	if _, err := Make("base 2", 2); err != nil {
		t.Error("expected OK")
	}
}

func TestMakeWithBase10(t *testing.T) {
	if _, err := Make("base 10", 10); err != nil {
		t.Error("expected OK")
	}
}
func TestMakeWithUnsupportedBase(t *testing.T) {
	if _, err := Make("base 10", 100); err == nil {
		t.Error("expected Unsupported")
	}
}

func TestBase2LogZero(t *testing.T) {
	f, _ := Make("2", 2)

	if v := f(0); v != math.Inf(-1) {
		t.Errorf("expected 0, got %f", v)
	}
}

func TestBase2LogTen(t *testing.T) {
	f, _ := Make("10", 10)

	if v := f(0); v != math.Inf(-1) {
		t.Errorf("expected 0, got %f", v)
	}
}
