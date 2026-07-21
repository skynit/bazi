package bazi

import "testing"

func TestValidRuleMetaRequiresCompleteAuthoritativeManifest(t *testing.T) {
	meta := DefaultRuleMeta()
	if !ValidRuleMeta(meta) {
		t.Fatal("default rule manifest must validate")
	}

	historicalLabel := cloneRuleMeta(meta)
	historicalLabel.RuleVersion = "stored-rule-v1"
	historicalLabel.School = "stored-school-v1"
	if !ValidRuleMeta(historicalLabel) {
		t.Fatal("version labels may be preserved when the complete manifest is authoritative")
	}

	tamperedTable := cloneRuleMeta(meta)
	tamperedTable.Tables[0].Description = "tampered"
	if ValidRuleMeta(tamperedTable) {
		t.Fatal("tampered rule table metadata must be rejected")
	}

	tamperedYueLing := cloneRuleMeta(meta)
	tamperedYueLing.BodyStrength.YueLing.TableSHA256 = "sha256:bad"
	if ValidRuleMeta(tamperedYueLing) {
		t.Fatal("tampered yue-ling table fingerprint must be rejected")
	}

	tamperedRoot := cloneRuleMeta(meta)
	tamperedRoot.BodyStrength.Root.TouGanMultiplier = 9
	if ValidRuleMeta(tamperedRoot) {
		t.Fatal("tampered root profile must be rejected")
	}

	tamperedBonus := cloneRuleMeta(meta)
	tamperedBonus.BodyStrength.Bonus.TableSHA256 = "sha256:bad"
	if ValidRuleMeta(tamperedBonus) {
		t.Fatal("tampered bonus table fingerprint must be rejected")
	}

	tamperedInfluence := cloneRuleMeta(meta)
	tamperedInfluence.BodyStrength.Influence.OfficerKillerWeight = 9
	if ValidRuleMeta(tamperedInfluence) {
		t.Fatal("tampered influence profile must be rejected")
	}

	tamperedNormalization := cloneRuleMeta(meta)
	tamperedNormalization.BodyStrength.Normalizers.ShengFormula = "centered_logistic_v1"
	if ValidRuleMeta(tamperedNormalization) {
		t.Fatal("tampered normalization formula must be rejected")
	}

	tamperedAdjustmentForce := cloneRuleMeta(meta)
	tamperedAdjustmentForce.BodyStrength.AdjustmentForce.StemForce = 9
	if ValidRuleMeta(tamperedAdjustmentForce) {
		t.Fatal("tampered adjustment-force profile must be rejected")
	}

	tamperedSource := cloneRuleMeta(meta)
	for tableIndex := range tamperedSource.Tables {
		if len(tamperedSource.Tables[tableIndex].Sources) == 0 {
			continue
		}
		tamperedSource.Tables[tableIndex].Sources[0].Files["tampered"] = "sha256:bad"
		if ValidRuleMeta(tamperedSource) {
			t.Fatal("tampered source fingerprint must be rejected")
		}
		return
	}
	t.Fatal("default manifest has no source fingerprint to test")
}
