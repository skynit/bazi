package handler

import (
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/bazi"
)

func TestValidatedBaziSnapshotRejectsTimeDependentEvidenceTampering(t *testing.T) {
	service := &bazi.BaziService{}
	calculated, err := service.Calculate(1990, 6, 15, 8, 30, model.GenderMale)
	if err != nil {
		t.Fatal(err)
	}
	chart := &model.BirthChart{
		EngineVersion: bazi.EngineVersion,
		RuleVersion:   calculated.RuleVersion,
	}

	assertRejected := func(name string, snapshot *bazi.BaziResult) {
		t.Helper()
		chart.BaziSnapshot = mustJSON(t, snapshot)
		if _, ok := validatedBaziSnapshot(chart, calculated, model.GenderMale); ok {
			t.Fatalf("snapshot with tampered %s must be rejected", name)
		}
	}

	tamperedDaYun := *calculated
	tamperedDaYun.DaYunInfo.StartAt = "2099-01-01T00:00:00"
	assertRejected("da-yun start instant", &tamperedDaYun)

	tamperedTiaohou := *calculated
	tiaohou := *calculated.Tiaohou
	tiaohou.DepthEvidence.ElapsedSeconds++
	tiaohou.DepthEvidence.Position = float64(tiaohou.DepthEvidence.ElapsedSeconds) /
		float64(tiaohou.DepthEvidence.IntervalSeconds)
	tamperedTiaohou.Tiaohou = &tiaohou
	assertRejected("tiaohou solar-term depth", &tamperedTiaohou)

	tamperedManifest := *calculated
	tamperedManifest.RuleMeta = bazi.DefaultRuleMeta()
	tamperedManifest.RuleMeta.Tables[0].Description = "tampered"
	assertRejected("rule manifest", &tamperedManifest)
}
