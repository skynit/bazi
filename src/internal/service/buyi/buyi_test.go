package buyi

import (
	"strings"
	"testing"
)

func TestHexagramsIntegrity(t *testing.T) {
	if len(Hexagrams) != 64 {
		t.Fatalf("expected 64 hexagrams, got %d", len(Hexagrams))
	}

	seen := make(map[int]bool, len(Hexagrams))
	for i, hexagram := range Hexagrams {
		wantNumber := i + 1
		if hexagram.Number != wantNumber {
			t.Fatalf("hexagram at index %d: expected number %d, got %d", i, wantNumber, hexagram.Number)
		}
		if seen[hexagram.Number] {
			t.Fatalf("duplicate hexagram number %d", hexagram.Number)
		}
		seen[hexagram.Number] = true
		if hexagram.Name == "" {
			t.Fatalf("hexagram %d has an empty name", hexagram.Number)
		}
	}
}

func TestHexagramProfilesAreComplete(t *testing.T) {
	if len(hexagramThemes) != len(Hexagrams) {
		t.Fatalf("expected %d hexagram themes, got %d", len(Hexagrams), len(hexagramThemes))
	}
	for _, hexagram := range Hexagrams {
		if themeFor(hexagram) == "" {
			t.Fatalf("hexagram %d has no theme", hexagram.Number)
		}
		if _, _, ok := trigramsFor(hexagram); !ok {
			t.Fatalf("hexagram %d cannot be decomposed into trigrams", hexagram.Number)
		}
	}
}

func TestBuildReadingUsesNeutralReflectionLanguage(t *testing.T) {
	for _, hexagram := range Hexagrams {
		reading := BuildReading(hexagram)
		if reading.Summary == "" || reading.Reflection == "" || reading.ImagePrompt == "" || reading.Advice == "" {
			t.Fatalf("%s produced an incomplete reading: %+v", hexagram.Name, reading)
		}
	}
}

func TestBuildReadingUsesHexagramContent(t *testing.T) {
	reading := BuildReading(Hexagrams[10])
	if reading.Hexagram.Name != "地天泰" {
		t.Fatalf("expected 地天泰, got %s", reading.Hexagram.Name)
	}
	if reading.Summary == "" || reading.Reflection == "" || reading.ImagePrompt == "" || reading.Advice == "" {
		t.Fatalf("expected complete reading, got %+v", reading)
	}
}

func TestXunReadingExplainsTheHexagram(t *testing.T) {
	reading := BuildReading(Hexagrams[56])
	if reading.Hexagram.Name != "巽为风" {
		t.Fatalf("expected 巽为风, got %s", reading.Hexagram.Name)
	}
	wants := []struct {
		name string
		got  string
		want string
	}{
		{name: "summary", got: reading.Summary, want: "上卦与下卦皆为巽（风）"},
		{name: "summary theme", got: reading.Summary, want: "持续沟通"},
		{name: "reflection", got: reading.Reflection, want: "目标、底线和可调整部分"},
		{name: "observation", got: reading.ImagePrompt, want: "反复摇摆"},
		{name: "advice", got: reading.Advice, want: "确认对方反馈"},
	}
	for _, item := range wants {
		if !strings.Contains(item.got, item.want) {
			t.Errorf("%s = %q, want it to contain %q", item.name, item.got, item.want)
		}
	}
}
