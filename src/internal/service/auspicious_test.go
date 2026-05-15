package service

import "testing"

func TestLuckyColorByStem(t *testing.T) {
	data := NewAuspiciousData()

	stems := []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	for _, stem := range stems {
		color := data.GetLuckyColor(stem)
		if color == "" {
			t.Errorf("GetLuckyColor(%q) returned empty string", stem)
		}
	}
}

func TestClashZodiac(t *testing.T) {
	data := NewAuspiciousData()

	zodiac := data.GetClashZodiac("子")
	if zodiac != "马" {
		t.Errorf("GetClashZodiac(子) = %q, want %q", zodiac, "马")
	}

	zodiac = data.GetClashZodiac("午")
	if zodiac != "鼠" {
		t.Errorf("GetClashZodiac(午) = %q, want %q", zodiac, "鼠")
	}
}
