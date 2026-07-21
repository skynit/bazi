package ziwei

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const (
	ziweiInputFingerprintVersion      = "ziwei-input-v1"
	ziweiDerivationFingerprintVersion = "ziwei-derivation-v2"
)

// ZiWeiCalculationInput is the exact normalized solar-minute input consumed by
// the current Zi Wei engine.
type ZiWeiCalculationInput struct {
	CalendarType string `json:"calendar_type"`
	Year         int    `json:"year"`
	Month        int    `json:"month"`
	Day          int    `json:"day"`
	Hour         int    `json:"hour"`
	Minute       int    `json:"minute"`
	Gender       string `json:"gender"`
	Basis        string `json:"basis"`
}

// ZiWeiDerivationInput records the complete normalized period query consumed
// when a transit chart is derived from a natal chart.
type ZiWeiDerivationInput struct {
	CalendarType      string                 `json:"calendar_type"`
	Year              int                    `json:"year"`
	Month             int                    `json:"month"`
	Day               int                    `json:"day"`
	Basis             string                 `json:"basis"`
	BoundaryPolicy    string                 `json:"boundary_policy"`
	ResolvedLunarDate ZiWeiResolvedLunarDate `json:"resolved_lunar_date"`
	PeriodGanZhi      string                 `json:"period_gan_zhi"`
}

type ZiWeiResolvedLunarDate struct {
	Year        int  `json:"year"`
	Month       int  `json:"month"`
	Day         int  `json:"day"`
	IsLeapMonth bool `json:"is_leap_month"`
}

func calculationInputFromBirth(birth *BirthData) ZiWeiCalculationInput {
	return ZiWeiCalculationInput{
		CalendarType: "SOLAR",
		Year:         birth.SolarYear,
		Month:        birth.SolarMonth,
		Day:          birth.SolarDay,
		Hour:         birth.Hour,
		Minute:       birth.Minute,
		Gender:       birth.Gender,
		Basis:        "normalized_solar_minute",
	}
}

func ziweiInputFingerprint(input ZiWeiCalculationInput) string {
	payload := fmt.Sprintf(
		"%s|%04d-%02d-%02dT%02d:%02d|%s",
		ziweiInputFingerprintVersion,
		input.Year,
		input.Month,
		input.Day,
		input.Hour,
		input.Minute,
		input.Gender,
	)
	digest := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(digest[:])
}

func chartContentHash(chart *ZiWeiChart) (string, error) {
	if chart == nil {
		return "", fmt.Errorf("chart is nil")
	}
	clone := *chart
	clone.ContentHash = ""
	clone.DerivedContentHash = ""
	payload, err := json.Marshal(&clone)
	if err != nil {
		return "", fmt.Errorf("marshal ziwei chart content: %w", err)
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:]), nil
}

func ziweiDerivationFingerprint(derivationType string, input ZiWeiDerivationInput) string {
	encoded, _ := json.Marshal(struct {
		Version        string               `json:"version"`
		DerivationType string               `json:"derivation_type"`
		Input          ZiWeiDerivationInput `json:"input"`
	}{
		Version:        ziweiDerivationFingerprintVersion,
		DerivationType: derivationType,
		Input:          input,
	})
	payload := string(encoded)
	digest := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(digest[:])
}

func validZiWeiDerivationInput(derivationType string, input ZiWeiDerivationInput) bool {
	want, err := buildZiWeiDerivationInput(derivationType, input.Year, input.Month, input.Day)
	return err == nil && input == want
}

func stampDerivedChartContract(result, base *ZiWeiChart, derivationType string, input ZiWeiDerivationInput) error {
	if result == nil || base == nil || !validZiWeiDerivationInput(derivationType, input) {
		return fmt.Errorf("valid result, base chart, and derivation input are required")
	}
	baseHash, err := chartContentHash(base)
	if err != nil {
		return fmt.Errorf("hash base ziwei chart: %w", err)
	}
	inputCopy := input
	result.ContentHash = ""
	result.DerivationType = derivationType
	result.DerivationInput = &inputCopy
	result.DerivationFingerprint = ziweiDerivationFingerprint(derivationType, input)
	result.BaseContentHash = baseHash
	result.DerivedContentHash = ""
	derivedHash, err := chartContentHash(result)
	if err != nil {
		return fmt.Errorf("hash derived ziwei chart: %w", err)
	}
	result.DerivedContentHash = derivedHash
	return nil
}

// ValidDerivedChartContract reconstructs the embedded natal chart, validates it
// under the current profile, then reruns the normalized transit query. A caller
// cannot make altered transit content valid by merely recomputing its hash.
func ValidDerivedChartContract(chart *ZiWeiChart) bool {
	if chart == nil || chart.ContentHash != "" || chart.DerivationInput == nil ||
		len(chart.DerivationFingerprint) != sha256.Size*2 ||
		len(chart.BaseContentHash) != sha256.Size*2 ||
		len(chart.DerivedContentHash) != sha256.Size*2 ||
		!validZiWeiDerivationInput(chart.DerivationType, *chart.DerivationInput) ||
		chart.DerivationFingerprint != ziweiDerivationFingerprint(chart.DerivationType, *chart.DerivationInput) {
		return false
	}
	want, err := chartContentHash(chart)
	if err != nil || chart.DerivedContentHash != want {
		return false
	}

	base, profile, ok := natalChartFromDerived(chart)
	if !ok || !chartMatchesProfile(base, profile) {
		return false
	}
	expected := rebuildDerivedChart(base, chart.DerivationType, *chart.DerivationInput)
	return expected != nil && expected.BaseContentHash == chart.BaseContentHash &&
		expected.DerivationFingerprint == chart.DerivationFingerprint &&
		expected.DerivedContentHash == chart.DerivedContentHash
}

// DerivedChartMatchesBase additionally verifies that base_content_hash points
// to the exact current public payload of the supplied parent chart.
func DerivedChartMatchesBase(derived, base *ZiWeiChart) bool {
	if !ValidDerivedChartContract(derived) || base == nil {
		return false
	}
	profile, err := ResolveProfile(base.ProfileID)
	return err == nil && chartMatchesProfile(base, profile) && derived.BaseContentHash == base.ContentHash
}

func natalChartFromDerived(derived *ZiWeiChart) (*ZiWeiChart, CalculationProfile, bool) {
	if derived == nil {
		return nil, CalculationProfile{}, false
	}
	profile, err := ResolveProfile(derived.ProfileID)
	if err != nil {
		return nil, CalculationProfile{}, false
	}
	base := *derived
	base.ContentHash = derived.BaseContentHash
	base.DerivationType = ""
	base.DerivationInput = nil
	base.DerivationFingerprint = ""
	base.BaseContentHash = ""
	base.DerivedContentHash = ""
	base.LiuNianStars = [12][]string{}
	base.LiuYueStars = [12][]string{}
	base.LiuRiStars = [12][]string{}
	base.LiuNianFourHua = [12][]string{}
	base.LiuYueFourHua = [12][]string{}
	base.LiuRiFourHua = [12][]string{}
	base.LiuNianPalaces = [12]string{}
	base.LiuYuePalaces = [12]string{}
	base.LiuRiPalaces = [12]string{}
	return &base, profile, true
}

func rebuildDerivedChart(base *ZiWeiChart, derivationType string, input ZiWeiDerivationInput) *ZiWeiChart {
	svc, err := NewZiWeiServiceWithProfile(base.ProfileID)
	if err != nil {
		return nil
	}
	switch derivationType {
	case "liunian":
		return svc.CalculateLiunian(base, input.Year)
	case "liuyue":
		return svc.CalculateLiuyueForDate(base, input.Year, input.Month, input.Day)
	case "liuri":
		return svc.CalculateLiuriForDate(base, input.Year, input.Month, input.Day)
	default:
		return nil
	}
}

func stampChartCacheContract(chart *ZiWeiChart, birth *BirthData) error {
	if chart == nil || birth == nil {
		return fmt.Errorf("chart and birth data are required")
	}
	chart.CalculationInput = calculationInputFromBirth(birth)
	chart.InputFingerprint = ziweiInputFingerprint(chart.CalculationInput)
	hash, err := chartContentHash(chart)
	if err != nil {
		return err
	}
	chart.ContentHash = hash
	return nil
}

func validChartContentHash(chart *ZiWeiChart) bool {
	if chart == nil || len(chart.ContentHash) != sha256.Size*2 {
		return false
	}
	want, err := chartContentHash(chart)
	return err == nil && chart.ContentHash == want
}

// birthDataFromPublishedChart restores calculation context for published
// consumers exclusively from the authenticated natal-chart JSON contract.
// Runtime-only fields are deliberately ignored so fresh and cached charts use
// the same evidence path.
func birthDataFromPublishedChart(chart *ZiWeiChart) (*BirthData, bool) {
	if chart == nil || !validChartContentHash(chart) ||
		chart.InputFingerprint != ziweiInputFingerprint(chart.CalculationInput) {
		return nil, false
	}
	birth, err := buildBirthData(
		chart.CalculationInput.Year,
		chart.CalculationInput.Month,
		chart.CalculationInput.Day,
		chart.CalculationInput.Hour,
		chart.CalculationInput.Minute,
		chart.CalculationInput.Gender,
	)
	if err != nil || calculationInputFromBirth(birth) != chart.CalculationInput {
		return nil, false
	}
	return birth, true
}
