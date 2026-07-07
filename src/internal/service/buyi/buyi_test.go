package buyi

import "testing"

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
		if hexagram.Name == "" || hexagram.HumanWay == "" || hexagram.ImageReading == "" {
			t.Fatalf("hexagram %d has empty required text", hexagram.Number)
		}
	}
}

func TestScoreHexagramRangeAndLevel(t *testing.T) {
	for _, hexagram := range Hexagrams {
		score := ScoreHexagram(hexagram)
		if score < 35 || score > 90 {
			t.Fatalf("%s score out of range: %d", hexagram.Name, score)
		}
		level := LevelForScore(score)
		if level == "" {
			t.Fatalf("%s got empty level for score %d", hexagram.Name, score)
		}
	}
}

func TestLevelForScoreThresholds(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{90, "大吉"},
		{82, "大吉"},
		{81, "吉"},
		{70, "吉"},
		{69, "平"},
		{55, "平"},
		{54, "谨慎"},
		{40, "谨慎"},
		{39, "凶险"},
		{35, "凶险"},
	}

	for _, tt := range tests {
		if got := LevelForScore(tt.score); got != tt.want {
			t.Fatalf("LevelForScore(%d): want %s, got %s", tt.score, tt.want, got)
		}
	}
}

func TestBuildReadingUsesHexagramContent(t *testing.T) {
	reading := BuildReading(Hexagrams[10])
	if reading.Hexagram.Name != "地天泰" {
		t.Fatalf("expected 地天泰, got %s", reading.Hexagram.Name)
	}
	if reading.Score == 0 || reading.Level == "" || reading.Summary == "" || reading.Advice == "" {
		t.Fatalf("expected complete reading, got %+v", reading)
	}
}
