package diff

import "testing"

func TestDiffAddedRemoved(t *testing.T) {
	old := map[string]any{
		"headers": map[string]any{"missing_headers": []any{"X-Frame-Options"}},
		"ports": map[string]any{"hosts": []any{
			map[string]any{"ip": "1.2.3.4", "open_ports": []any{
				map[string]any{"port": float64(80)},
			}},
		}},
	}
	newer := map[string]any{
		"headers": map[string]any{"missing_headers": []any{"X-Frame-Options"}},
		"ports": map[string]any{"hosts": []any{
			map[string]any{"ip": "1.2.3.4", "open_ports": []any{
				map[string]any{"port": float64(80)},
				map[string]any{"port": float64(3389)},
			}},
		}},
		"cors": map[string]any{"misconfigured": true},
	}
	d := Diff(old, newer)
	if len(d.Added) != 2 || d.Added[0] != "cors_misconfig" || d.Added[1] != "open_port:1.2.3.4:3389" {
		t.Fatalf("added = %v", d.Added)
	}
	if len(d.Removed) != 0 {
		t.Fatalf("removed = %v", d.Removed)
	}
}

func TestDiffIdenticalEmpty(t *testing.T) {
	d := Diff(map[string]any{}, map[string]any{})
	if len(d.Added) != 0 || len(d.Removed) != 0 {
		t.Fatalf("got %+v, want empty", d)
	}
	if d.Added == nil || d.Removed == nil {
		t.Fatal("added/removed must be non-nil for JSON rendering")
	}
}
