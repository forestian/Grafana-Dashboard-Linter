package parser

import "testing"

func TestParseDashboardBytesDirect(t *testing.T) {
	dashboard, err := ParseDashboardBytes([]byte(`{
		"uid": "direct",
		"title": "Direct Dashboard",
		"panels": []
	}`))
	if err != nil {
		t.Fatalf("ParseDashboardBytes returned error: %v", err)
	}
	if dashboard["uid"] != "direct" {
		t.Fatalf("uid = %v, want direct", dashboard["uid"])
	}
}

func TestParseDashboardBytesWrapped(t *testing.T) {
	dashboard, err := ParseDashboardBytes([]byte(`{
		"dashboard": {
			"uid": "wrapped",
			"title": "Wrapped Dashboard",
			"panels": []
		},
		"overwrite": true
	}`))
	if err != nil {
		t.Fatalf("ParseDashboardBytes returned error: %v", err)
	}
	if dashboard["uid"] != "wrapped" {
		t.Fatalf("uid = %v, want wrapped", dashboard["uid"])
	}
	if _, ok := dashboard["overwrite"]; ok {
		t.Fatalf("wrapped export metadata leaked into dashboard object")
	}
}

func TestParseDashboardBytesInvalidJSON(t *testing.T) {
	if _, err := ParseDashboardBytes([]byte(`{"title":`)); err == nil {
		t.Fatalf("ParseDashboardBytes returned nil error for invalid JSON")
	}
}
