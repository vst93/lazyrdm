package service

import (
	"testing"
	"tinyrdm/backend/types"
)

func newTestConnectionConfig() types.ConnectionConfig {
	return types.ConnectionConfig{
		Name:    "test",
		Addr:    "127.0.0.1",
		Port:    6379,
		Network: "tcp",
	}
}

func TestSplitRedisCommand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "simple command",
			input: "GET key1",
			want:  []string{"GET", "key1"},
		},
		{
			name:  "single command",
			input: "PING",
			want:  []string{"PING"},
		},
		{
			name:  "command with multiple args",
			input: "SET key1 value1 EX 60",
			want:  []string{"SET", "key1", "value1", "EX", "60"},
		},
		{
			name:  "quoted value with spaces",
			input: `SET key1 "hello world"`,
			want:  []string{"SET", "key1", "hello world"},
		},
		{
			name:  "single quoted value",
			input: `SET key1 'hello world'`,
			want:  []string{"SET", "key1", "hello world"},
		},
		{
			name:  "empty command",
			input: "",
			want:  nil,
		},
		{
			name:  "command with extra spaces",
			input: "  GET    key1  ",
			want:  []string{"GET", "key1"},
		},
		{
			name:  "HGETALL command",
			input: "HGETALL myhash",
			want:  []string{"HGETALL", "myhash"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitRedisCommand(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("splitRedisCommand(%q) = %v, want %v", tt.input, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitRedisCommand(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestFormatRedisResult(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{"nil", nil, "(nil)"},
		{"string", "hello", `"hello"`},
		{"int64", int64(42), "(integer) 42"},
		{"bool true", true, "(true)"},
		{"bool false", false, "(false)"},
		{"empty array", []any{}, "(empty array)"},
		{"string array", []any{"a", "b"}, `1) "a"
2) "b"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatRedisResult(tt.input)
			if got != tt.want {
				t.Errorf("formatRedisResult(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDisplayWidth(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"hello", 5},
		{"", 0},
		{"你好", 4},  // Chinese chars are width 2 each
		{"a中b", 4}, // mix of ASCII and Chinese
	}

	for _, tt := range tests {
		got := DisplayWidth(tt.input)
		if got != tt.want {
			t.Errorf("DisplayWidth(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeSetKeyType(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"string", "string"},
		{"list", "list"},
		{"hash", "hash"},
		{"set", "set"},
		{"zset", "zset"},
		{"stream", "stream"},
		{"rejson-rl", "json"},
		// Note: normalizeSetKeyType only handles lowercase "rejson-rl"
	}

	for _, tt := range tests {
		got := normalizeSetKeyType(tt.input)
		if got != tt.want {
			t.Errorf("normalizeSetKeyType(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsCollectionKeyType(t *testing.T) {
	collectionTypes := []string{"list", "hash", "set", "zset", "stream"}
	for _, kt := range collectionTypes {
		if !isCollectionKeyType(kt) {
			t.Errorf("isCollectionKeyType(%q) should be true", kt)
		}
	}

	nonCollectionTypes := []string{"string", "json", "rejson-rl", "", "unknown"}
	for _, kt := range nonCollectionTypes {
		if isCollectionKeyType(kt) {
			t.Errorf("isCollectionKeyType(%q) should be false", kt)
		}
	}
}

func TestBuildRedisOptions(t *testing.T) {
	// Test basic TCP options
	config := newTestConnectionConfig()
	opts := buildRedisOptions(config, 0)
	if opts.Addr != "127.0.0.1:6379" {
		t.Errorf("expected addr 127.0.0.1:6379, got %s", opts.Addr)
	}
	if opts.Network != "tcp" {
		t.Errorf("expected network tcp, got %s", opts.Network)
	}
	if opts.DB != 0 {
		t.Errorf("expected db 0, got %d", opts.DB)
	}

	// Test with custom port
	config.Addr = "redis.example.com"
	config.Port = 6380
	opts = buildRedisOptions(config, 3)
	if opts.Addr != "redis.example.com:6380" {
		t.Errorf("expected addr redis.example.com:6380, got %s", opts.Addr)
	}
	if opts.DB != 3 {
		t.Errorf("expected db 3, got %d", opts.DB)
	}
}

func TestSplitCommand(t *testing.T) {
	// Test the splitCommand function from vi_editor.go
	tests := []struct {
		input string
		want  []string
	}{
		{`vim`, []string{"vim"}},
		{`vim -c "set nu"`, []string{"vim", "-c", "set nu"}},
		{`nano -B`, []string{"nano", "-B"}},
		{``, []string{}},
	}

	for _, tt := range tests {
		got := splitCommand(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("splitCommand(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("splitCommand(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestParseHashFieldItems(t *testing.T) {
	// Test array format
	result, err := parseHashFieldItems(`[{"field":"f1","value":"v1"}]`)
	if err != nil {
		t.Fatalf("parseHashFieldItems array failed: %v", err)
	}
	if len(result) != 2 || result[0] != "f1" || result[1] != "v1" {
		t.Errorf("parseHashFieldItems array = %v, want [f1 v1]", result)
	}

	// Test object format
	result, err = parseHashFieldItems(`{"f1":"v1"}`)
	if err != nil {
		t.Fatalf("parseHashFieldItems object failed: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("parseHashFieldItems object = %v, want 2 items", result)
	}

	// Test invalid format
	_, err = parseHashFieldItems(`not json`)
	if err == nil {
		t.Error("parseHashFieldItems should fail on invalid JSON")
	}
}

func TestParseSetMembers(t *testing.T) {
	result, err := parseSetMembers(`["m1","m2","m3"]`)
	if err != nil {
		t.Fatalf("parseSetMembers failed: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("parseSetMembers = %v, want 3 members", result)
	}

	_, err = parseSetMembers(`invalid`)
	if err == nil {
		t.Error("parseSetMembers should fail on invalid JSON")
	}
}

func TestParseZSetDeleteMember(t *testing.T) {
	member, err := parseZSetDeleteMember(`{"value":"m1"}`)
	if err != nil {
		t.Fatalf("parseZSetDeleteMember failed: %v", err)
	}
	if member != "m1" {
		t.Errorf("parseZSetDeleteMember = %q, want m1", member)
	}

	member, err = parseZSetDeleteMember(`{"member":"m2"}`)
	if err != nil {
		t.Fatalf("parseZSetDeleteMember with member failed: %v", err)
	}
	if member != "m2" {
		t.Errorf("parseZSetDeleteMember with member = %q, want m2", member)
	}
}

func TestFormatZSetScore(t *testing.T) {
	if got := formatZSetScore(1.5, ""); got != "1.5" {
		t.Errorf("formatZSetScore(1.5, '') = %q, want 1.5", got)
	}
	if got := formatZSetScore(0, "10"); got != "10" {
		t.Errorf("formatZSetScore(0, '10') = %q, want 10", got)
	}
}

func TestScoreAnyToString(t *testing.T) {
	if got := scoreAnyToString("3.14"); got != "3.14" {
		t.Errorf("scoreAnyToString('3.14') = %q, want 3.14", got)
	}
	if got := scoreAnyToString(float64(42)); got != "42" {
		t.Errorf("scoreAnyToString(42.0) = %q, want 42", got)
	}
	if got := scoreAnyToString(int(5)); got != "5" {
		t.Errorf("scoreAnyToString(5) = %q, want 5", got)
	}
	if got := scoreAnyToString(nil); got != "" {
		t.Errorf("scoreAnyToString(nil) = %q, want ''", got)
	}
}

func TestTruncateByRuneCount(t *testing.T) {
	if got := truncateByRuneCount("hello", 10); got != "hello" {
		t.Errorf("truncateByRuneCount('hello', 10) = %q, want hello", got)
	}
	if got := truncateByRuneCount("hello world", 5); got != "he..." {
		t.Errorf("truncateByRuneCount('hello world', 5) = %q, want he...", got)
	}
	if got := truncateByRuneCount("hi", 0); got != "" {
		t.Errorf("truncateByRuneCount('hi', 0) = %q, want ''", got)
	}
}

func TestRequireNonEmpty(t *testing.T) {
	val, err := requireNonEmpty(map[string]string{"key": "value"}, "key")
	if err != nil || val != "value" {
		t.Errorf("requireNonEmpty with value failed: val=%q, err=%v", val, err)
	}

	_, err = requireNonEmpty(map[string]string{"key": ""}, "key")
	if err == nil {
		t.Error("requireNonEmpty should fail on empty value")
	}
}

func TestParseRequiredInt(t *testing.T) {
	val, err := parseRequiredInt(map[string]string{"num": "42"}, "num")
	if err != nil || val != 42 {
		t.Errorf("parseRequiredInt with 42 failed: val=%d, err=%v", val, err)
	}

	_, err = parseRequiredInt(map[string]string{"num": "abc"}, "num")
	if err == nil {
		t.Error("parseRequiredInt should fail on non-integer")
	}
}
