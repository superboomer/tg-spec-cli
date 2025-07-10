package main

import "testing"

func TestMain(t *testing.T) {
	// Just test that main does not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("main panicked: %v", r)
		}
	}()
	main()
}
