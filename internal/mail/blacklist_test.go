package mail

import "testing"

func TestIsBlacklisted(t *testing.T) {
	tests := []struct {
		shortID     string
		blacklist   string
		wantBlocked bool
	}{
		{"admin", "admin,root", true},
		{"admin123", "admin,root", true},
		{"myadminbox", "admin", true},
		{"test", "admin,root", false},
		{"", "admin", false},
		{"test", "", false},
		{"", "", false},
		{"Admin", "admin", true},  // case insensitive
		{"ROOT", "root", true},    // case insensitive
		{"user", "admin,root", false},
	}

	for _, tt := range tests {
		got := IsBlacklisted(tt.shortID, tt.blacklist)
		if got != tt.wantBlocked {
			t.Errorf("IsBlacklisted(%q, %q) = %v, want %v", tt.shortID, tt.blacklist, got, tt.wantBlocked)
		}
	}
}

func TestIsDomainAllowed(t *testing.T) {
	tests := []struct {
		domain    string
		mailHost  string
		wantAllow bool
	}{
		{"example.com", "example.com", true},
		{"example.com", "example.com,test.com", true},
		{"test.com", "example.com,test.com", true},
		{"other.com", "example.com,test.com", false},
		{"example.com", "", true}, // empty means allow all
		{"", "example.com", false},
		{"EXAMPLE.COM", "example.com", true}, // case insensitive
	}

	for _, tt := range tests {
		got := IsDomainAllowed(tt.domain, tt.mailHost)
		if got != tt.wantAllow {
			t.Errorf("IsDomainAllowed(%q, %q) = %v, want %v", tt.domain, tt.mailHost, got, tt.wantAllow)
		}
	}
}

func TestParseDomains(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"example.com", []string{"example.com"}},
		{"example.com,test.com", []string{"example.com", "test.com"}},
		{"  a.com , b.com  ", []string{"a.com", "b.com"}},
		{"", nil},
		{",,,", nil},
	}

	for _, tt := range tests {
		got := ParseDomains(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("ParseDomains(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("ParseDomains(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}
