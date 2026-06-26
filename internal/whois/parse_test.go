package whois

import "testing"

func TestParseExpiration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "registrar registration expiration",
			input: "Registrar Registration Expiration Date: 2026-12-31T23:59:59Z",
			want:  "2026-12-31",
		},
		{
			name:  "expiry date dd-mon-yyyy",
			input: "Expiry Date: 31-Dec-2026",
			want:  "2026-12-31",
		},
		{
			name:  "paid-till dot format",
			input: "paid-till: 2026.07.03",
			want:  "2026-07-03",
		},
		{
			name:    "no match",
			input:   "Domain Name: EXAMPLE.COM",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseExpiration(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}