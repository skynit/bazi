package ziwei

import "testing"

func TestBuildQueryViewRequiresAuthenticatedNatalOrDerivedChart(t *testing.T) {
	svc := NewZiWeiService()
	natal, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	natalView := svc.BuildQueryView(natal)
	if natalView == nil || natalView.RuleVersion != QueryRuleVersion || len(natalView.Palaces) != 12 {
		t.Fatalf("authenticated natal query view = %+v", natalView)
	}

	derived := svc.CalculateLiunian(natal, 2026)
	if derived == nil {
		t.Fatal("failed to build valid derived chart fixture")
	}
	derivedView := svc.BuildQueryView(derived)
	if derivedView == nil || derivedView.RuleVersion != QueryRuleVersion || len(derivedView.Palaces) != 12 {
		t.Fatalf("authenticated derived query view = %+v", derivedView)
	}

	tampered := roundTripProjectionFixture(t, derived)
	tampered.Palaces[0].Name = "篡改宫"
	restampZiWeiDerivedContentHash(t, tampered)
	if svc.BuildQueryView(tampered) != nil {
		t.Fatal("query view accepted a rehashed derived chart that cannot be reconstructed")
	}
}
