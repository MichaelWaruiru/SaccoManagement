package main

import "testing"

func TestMyFunction(t *testing.T) {
	result := MyFunction(2, 3)
	expected := 5
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}

func MyFunction(a, b int) int {
	return a + b
}
