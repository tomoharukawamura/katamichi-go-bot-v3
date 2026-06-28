package scraper

import "testing"

func TestGroupForCity(t *testing.T) {
	tests := []struct {
		city      string
		wantGroup string
		wantOk    bool
	}{
		{"青森", "北東北", true},
		{"岩手", "北東北", true},
		{"宮城", "南東北", true},
		{"山形", "南東北", true},
		{"福島", "南東北", true},
		{"静岡", "静岡", true},
		{"愛知", "愛知", true},
		{"長野", "甲信越", true},
		{"東京", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.city, func(t *testing.T) {
			got, ok := GroupForCity(tt.city)
			if ok != tt.wantOk {
				t.Errorf("GroupForCity(%q) ok=%v, want %v", tt.city, ok, tt.wantOk)
			}
			if got != tt.wantGroup {
				t.Errorf("GroupForCity(%q) = %q, want %q", tt.city, got, tt.wantGroup)
			}
		})
	}
}
