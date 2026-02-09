package common

import "testing"

func TestExtractGoogleResourceID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain id",
			input: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
			want:  "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
		},
		{
			name:  "drive file view url",
			input: "https://drive.google.com/file/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/view",
			want:  "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
		},
		{
			name:  "drive open url",
			input: "https://drive.google.com/open?id=1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
			want:  "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
		},
		{
			name:  "docs url with edit",
			input: "https://docs.google.com/document/d/abc123/edit",
			want:  "abc123",
		},
		{
			name:  "docs url without edit",
			input: "https://docs.google.com/document/d/abc123",
			want:  "abc123",
		},
		{
			name:  "docs url with query params",
			input: "https://docs.google.com/document/d/abc123/edit?usp=sharing",
			want:  "abc123",
		},
		{
			name:  "sheets url",
			input: "https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit",
			want:  "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
		},
		{
			name:  "id with special characters",
			input: "doc-with-dashes_and_underscores",
			want:  "doc-with-dashes_and_underscores",
		},
		{
			name:  "short plain id",
			input: "abc123",
			want:  "abc123",
		},
		{
			name:  "open url with ampersand",
			input: "https://drive.google.com/open?foo=bar&id=myFileID",
			want:  "myFileID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractGoogleResourceID(tt.input)
			if got != tt.want {
				t.Errorf("ExtractGoogleResourceID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractSpreadsheetID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain id",
			input: "abc123",
			want:  "abc123",
		},
		{
			name:  "sheets url with edit",
			input: "https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit",
			want:  "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
		},
		{
			name:  "sheets url without edit",
			input: "https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
			want:  "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
		},
		{
			name:  "full id",
			input: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
			want:  "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractSpreadsheetID(tt.input)
			if got != tt.want {
				t.Errorf("ExtractSpreadsheetID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
