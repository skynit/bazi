package main

import (
	"testing"

	"bazi/internal/model"
	"gorm.io/datatypes"
)

func TestMemChartStoreUpdate(t *testing.T) {
	s := newMemChartStore()
	c := &model.BirthChart{
		Name:       "test",
		Gender:     "男",
		BirthYear:  1990,
		BirthMonth: 6,
		BirthDay:   15,
	}
	if err := s.Create(c); err != nil {
		t.Fatalf("create: %v", err)
	}

	c.ZiWeiComputed = true
	c.ZiWeiResult = datatypes.JSON(`{"ming_gong":"寅"}`)
	if err := s.Update(c); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, err := s.FindByID(c.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got == nil {
		t.Fatal("chart not found after update")
	}
	if !got.ZiWeiComputed {
		t.Error("ZiWeiComputed should be true")
	}
	if string(got.ZiWeiResult) != `{"ming_gong":"寅"}` {
		t.Errorf("ZiWeiResult mismatch: %s", string(got.ZiWeiResult))
	}
}

func TestMemChartStoreUpdateNotFound(t *testing.T) {
	s := newMemChartStore()
	err := s.Update(&model.BirthChart{})
	if err == nil {
		t.Error("expected error for non-existent chart")
	}
}

func TestMemFortuneStoreSaveAndList(t *testing.T) {
	s := newMemFortuneStore()

	records := []model.HistoryResponse{
		{ChartID: 1, QueryDate: "2024-01-01", DayGanZhi: "甲子", Summary: "good"},
		{ChartID: 1, QueryDate: "2024-01-02", DayGanZhi: "乙丑", Summary: "ok"},
		{ChartID: 2, QueryDate: "2024-01-01", DayGanZhi: "丙寅", Summary: "bad"},
	}
	for _, r := range records {
		s.SaveRecord(r)
	}

	items, total, err := s.ListByChartID(1, 1, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 2 {
		t.Errorf("total should be 2, got %d", total)
	}
	if len(items) != 2 {
		t.Fatalf("should return 2 items, got %d", len(items))
	}
	if items[0].ChartID != 1 {
		t.Errorf("chart_id mismatch: %d", items[0].ChartID)
	}
}

func TestMemFortuneStoreEmptyList(t *testing.T) {
	s := newMemFortuneStore()
	items, total, err := s.ListByChartID(999, 1, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 0 {
		t.Errorf("total should be 0, got %d", total)
	}
	if len(items) != 0 {
		t.Errorf("items should be empty, got %d", len(items))
	}
}
