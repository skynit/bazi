package model

import (
	"encoding/json"
	"testing"

	"gorm.io/datatypes"
)

func TestBirthChartZiWeiFields(t *testing.T) {
	chart := BirthChart{
		Name:          "测试",
		Gender:        "男",
		BirthYear:     1990,
		BirthMonth:    6,
		BirthDay:      15,
		BirthHour:     8,
		ZiWeiComputed: true,
		ZiWeiResult:   datatypes.JSON(json.RawMessage(`{"ming_gong":"寅","shen_gong":"申"}`)),
	}

	data, err := json.Marshal(chart)
	if err != nil {
		t.Fatalf("marshal BirthChart: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	if v, ok := result["ziwei_computed"]; !ok {
		t.Error("missing ziwei_computed in JSON")
	} else if v != true {
		t.Errorf("ziwei_computed should be true, got %v", v)
	}

	if v, ok := result["ziwei_result"]; !ok {
		t.Error("missing ziwei_result in JSON")
	} else {
		zr, ok := v.(map[string]interface{})
		if !ok {
			t.Fatalf("ziwei_result should be object, got %T", v)
		}
		if zr["ming_gong"] != "寅" {
			t.Errorf("ming_gong should be 寅, got %v", zr["ming_gong"])
		}
	}

	// Round-trip
	var chart2 BirthChart
	if err := json.Unmarshal(data, &chart2); err != nil {
		t.Fatalf("unmarshal back: %v", err)
	}
	if chart2.ZiWeiComputed != true {
		t.Error("round-trip: ziwei_computed mismatch")
	}
	if string(chart2.ZiWeiResult) != `{"ming_gong":"寅","shen_gong":"申"}` {
		t.Errorf("round-trip: ziwei_result mismatch: %s", string(chart2.ZiWeiResult))
	}
}

func TestBirthChartZiWeiDefaults(t *testing.T) {
	chart := BirthChart{
		Name:      "默认",
		Gender:    "女",
		BirthYear: 2000,
	}

	if chart.ZiWeiComputed != false {
		t.Error("ZiWeiComputed default should be false")
	}
	if chart.ZiWeiResult != nil {
		t.Error("ZiWeiResult default should be nil")
	}

	data, _ := json.Marshal(chart)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	if v, ok := result["ziwei_computed"]; !ok || v != false {
		t.Error("default ziwei_computed should be false in JSON")
	}
}
