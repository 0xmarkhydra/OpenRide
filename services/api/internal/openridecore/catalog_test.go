package openridecore

import "testing"

func TestDefaultServiceCatalog(t *testing.T) {
	manifests := DefaultServiceManifests()
	if len(manifests) < 2 {
		t.Fatalf("expected at least 2 first-party modules, got %d", len(manifests))
	}
	ids := map[string]bool{}
	for _, manifest := range manifests {
		ids[string(manifest.ID)] = true
	}
	if !ids["passenger.car"] {
		t.Fatal("passenger.car module missing")
	}
	if !ids["carpool.intercity"] {
		t.Fatal("carpool.intercity module missing")
	}
}
