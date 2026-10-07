package analyzer

import "testing"

func portsResults() map[string]any {
	return map[string]any{
		"ports": map[string]any{
			"host": "localhost", "profile": "top100",
			"hosts": []any{
				map[string]any{"ip": "127.0.0.1", "open_ports": []any{
					// float64 shape = rows decoded from the DB JSON column
					map[string]any{"port": float64(445), "service": "smb"},
					map[string]any{"port": float64(8080), "service": "http-alt"},
				}},
			},
		},
	}
}

func TestExtractOpenPorts(t *testing.T) {
	findings := ExtractFindings(portsResults())
	if len(findings) != 2 {
		t.Fatalf("got %d findings, want 2: %+v", len(findings), findings)
	}
	if findings[0]["type"] != "open_port" {
		t.Fatalf("finding type = %v, want open_port", findings[0]["type"])
	}
}

func TestScoreOpenPorts(t *testing.T) {
	findings := ExtractFindings(portsResults())
	scores := map[int]int{}
	for _, f := range findings {
		p, _ := portNumber(f["port"])
		scores[p] = ScoreFinding(f)
	}
	if scores[445] != 75 {
		t.Fatalf("smb/445 scored %d, want 75 (high)", scores[445])
	}
	if scores[8080] != 40 {
		t.Fatalf("http-alt/8080 scored %d, want 40 (medium)", scores[8080])
	}
	if ScoreToLevel(75) != "high" || ScoreToLevel(40) != "medium" {
		t.Fatal("level mapping wrong")
	}
}

func TestNoPortsSectionNoFindings(t *testing.T) {
	if f := ExtractFindings(map[string]any{"tls": map[string]any{}}); len(f) != 0 {
		t.Fatalf("got %+v, want none", f)
	}
}
