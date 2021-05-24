package helpers

import "testing"

func TestCyrillicToLatin(t *testing.T) {
	data := map[string]string{
		"Первый":        "pervii",
		"ТНТ":           "tnt",
		"Матч! Премьер": "match!_premer",
	}

	for i, v := range data {
		if result := CyrillicToLatin(i); result != v {
			t.Errorf("expected %q, got %q\n", v, result)
		}
	}
}
