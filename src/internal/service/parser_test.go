package service

import (
	"testing"
)

func TestParseSolar(t *testing.T) {
	p := &InputParser{}
	req := ParseRequest{
		Input:        "1990-01-15 08:00",
		CalendarType: "SOLAR",
		Gender:       "MALE",
	}
	result, err := p.Parse(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Year != 1990 {
		t.Errorf("Year: want 1990, got %d", result.Year)
	}
	if result.Month != 1 {
		t.Errorf("Month: want 1, got %d", result.Month)
	}
	if result.Day != 15 {
		t.Errorf("Day: want 15, got %d", result.Day)
	}
	if result.Hour != 8 {
		t.Errorf("Hour: want 8, got %d", result.Hour)
	}
	if result.Minute != 0 {
		t.Errorf("Minute: want 0, got %d", result.Minute)
	}
	if result.Gender != "MALE" {
		t.Errorf("Gender: want MALE, got %s", result.Gender)
	}
}

func TestParseSolarAuto(t *testing.T) {
	p := &InputParser{}
	req := ParseRequest{
		Input:        "1990-01-15 08:00",
		CalendarType: "",
		Gender:       "FEMALE",
	}
	result, err := p.Parse(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Year != 1990 || result.Month != 1 || result.Day != 15 || result.Hour != 8 || result.Minute != 0 {
		t.Errorf("expected {1990,1,15,8,0}, got {%d,%d,%d,%d,%d}",
			result.Year, result.Month, result.Day, result.Hour, result.Minute)
	}
	if result.Gender != "FEMALE" {
		t.Errorf("Gender: want FEMALE, got %s", result.Gender)
	}
}

func TestParseBazi(t *testing.T) {
	p := &InputParser{}
	req := ParseRequest{
		Input:        "己巳 丁丑 庚辰 庚辰",
		CalendarType: "BAZI",
		Gender:       "FEMALE",
	}
	result, err := p.Parse(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Gender != "FEMALE" {
		t.Errorf("Gender: want FEMALE, got %s", result.Gender)
	}
	// Bazi format has no calendar date — date fields are zero.
	if result.Year != 0 || result.Month != 0 || result.Day != 0 || result.Hour != 0 || result.Minute != 0 {
		t.Errorf("expected zero date fields for bazi input, got {%d,%d,%d,%d,%d}",
			result.Year, result.Month, result.Day, result.Hour, result.Minute)
	}
}

func TestParseInvalid(t *testing.T) {
	p := &InputParser{}
	req := ParseRequest{
		Input:        "invalid date string",
		CalendarType: "",
		Gender:       "MALE",
	}
	_, err := p.Parse(req)
	if err == nil {
		t.Fatal("expected error for invalid input, got nil")
	}
}

func TestParseInvalidGender(t *testing.T) {
	p := &InputParser{}
	req := ParseRequest{
		Input:        "1990-01-15 08:00",
		CalendarType: "SOLAR",
		Gender:       "OTHER",
	}
	_, err := p.Parse(req)
	if err == nil {
		t.Fatal("expected error for invalid gender, got nil")
	}
}
