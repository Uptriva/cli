package cmd

import "testing"

func TestRootCommands(t *testing.T) {
	root := NewRoot("test")
	names := map[string]bool{}
	for _, c := range root.Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"auth", "monitors", "incidents", "status-pages"} {
		if !names[want] {
			t.Fatalf("missing command %s", want)
		}
	}
}
