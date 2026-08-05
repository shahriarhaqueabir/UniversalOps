package devops

import "testing"

func TestFirstField(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if got := firstField(""); got != "" {
			t.Fatalf("firstField(\"\") = %q, want empty string", got)
		}
	})

	t.Run("whitespace only", func(t *testing.T) {
		if got := firstField("   \t\n"); got != "" {
			t.Fatalf("firstField(whitespace) = %q, want empty string", got)
		}
	})

	t.Run("with content", func(t *testing.T) {
		if got := firstField("12MiB /vol"); got != "12MiB" {
			t.Fatalf("firstField(content) = %q, want %q", got, "12MiB")
		}
	})
}
