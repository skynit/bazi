package bazi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const (
	prospectivePilotProtocolPath   = "../../../../research/pilot/prospective-pilot-protocol-v1.json"
	prospectivePilotConsentPath    = "../../../../research/pilot/participant-consent-v1.zh-CN.md"
	prospectivePilotProtocolSHA256 = "316026343a071e7d90050001119b5b58b83c4f15bcdea2203bf741852fcdce78"
	prospectivePilotConsentSHA256  = "f176e00a4f1dc14c50d7a54e2a2929162bb0fc5fae26cbf7e63f93d3c99f177b"
)

func TestProspectivePilotProtocolIsDraftAndCollectionDisabled(t *testing.T) {
	protocol, raw := loadProspectivePilotProtocol(t)
	if got := prospectivePilotSHA256(raw); got != prospectivePilotProtocolSHA256 {
		t.Fatalf("prospective Pilot protocol SHA-256 = %s, want %s", got, prospectivePilotProtocolSHA256)
	}
	if pilotString(t, protocol, "schema") != "prospective_fortune_pilot_protocol_v1" ||
		pilotString(t, protocol, "protocol_id") != "fortune-prospective-pilot-v1" ||
		pilotString(t, protocol, "version") != "2026-07-17.1" ||
		pilotString(t, protocol, "status") != "draft_not_approved" {
		t.Fatalf("invalid Pilot identity: schema=%v id=%v version=%v status=%v", protocol["schema"], protocol["protocol_id"], protocol["version"], protocol["status"])
	}

	runtime := pilotObject(t, protocol["runtime_boundary"])
	if !pilotBool(t, runtime, "research_only") || pilotBool(t, runtime, "recruitment_enabled") ||
		pilotBool(t, runtime, "event_collection_enabled") || pilotBool(t, runtime, "pilot_api_enabled") ||
		pilotBool(t, runtime, "pilot_database_enabled") || pilotBool(t, runtime, "participant_predictions_visible_during_followup") {
		t.Errorf("Pilot runtime boundary permits unapproved activity: %+v", runtime)
	}

	profiles := pilotObject(t, protocol["profile_gate"])
	if len(pilotArray(t, profiles["eligible_profiles"])) != 0 {
		t.Error("draft Pilot must have no eligible calculation profiles")
	}
	profileCandidates := pilotArray(t, profiles["candidates"])
	if len(profileCandidates) != 3 {
		t.Fatalf("Pilot profile candidates = %d, want 3", len(profileCandidates))
	}
	for _, value := range profileCandidates {
		candidate := pilotObject(t, value)
		if !strings.HasPrefix(pilotString(t, candidate, "status"), "not_eligible_") {
			t.Errorf("Pilot profile %s is prematurely eligible: %+v", pilotString(t, candidate, "profile_id"), candidate)
		}
	}

	design := pilotObject(t, protocol["study_design"])
	if pilotInt(t, design, "proposed_pilot_sample_size") != 200 ||
		pilotString(t, design, "sample_size_status") != "proposal_not_power_approved" || design["followup_duration"] != nil ||
		pilotString(t, design, "followup_status") != "not_approved" || len(pilotArray(t, design["followup_schedule"])) != 0 {
		t.Errorf("Pilot sample/follow-up is being treated as approved: %+v", design)
	}
	randomization := pilotObject(t, design["randomization"])
	if pilotBool(t, randomization, "enabled") || pilotString(t, randomization, "status") != "method_not_approved" ||
		!pilotBool(t, randomization, "seed_commitment_required") || !pilotBool(t, randomization, "allocation_concealment_required") {
		t.Errorf("Pilot randomization contract is incomplete: %+v", randomization)
	}
}

func TestProspectivePilotProtocolFreezesEthicsPrivacyAndScoringGates(t *testing.T) {
	protocol, _ := loadProspectivePilotProtocol(t)
	eventScope := pilotObject(t, protocol["event_scope"])
	if pilotString(t, eventScope, "status") != "taxonomy_not_approved" || pilotBool(t, eventScope, "free_text_enabled") ||
		pilotBool(t, eventScope, "exact_amount_enabled") || pilotBool(t, eventScope, "exact_address_enabled") ||
		pilotBool(t, eventScope, "retrospective_event_entry_enabled") {
		t.Errorf("Pilot event taxonomy permits unapproved or excessive collection: %+v", eventScope)
	}
	prohibited := pilotStringSet(t, eventScope["prohibited_categories"])
	for _, category := range []string{
		"death_or_lifespan", "medical_or_mental_health_diagnosis", "self_harm_or_violence",
		"crime_or_legal_outcome", "investment_return_or_financial_instruction", "pregnancy_or_fertility",
		"sexual_activity", "marriage_or_divorce_decision", "minor_child_outcome",
	} {
		if !prohibited[category] {
			t.Errorf("Pilot prohibited event taxonomy missing %q", category)
		}
	}

	prediction := pilotObject(t, protocol["prediction_contract"])
	if pilotString(t, prediction, "status") != "not_approved" || pilotBool(t, prediction, "probability_output_enabled") ||
		pilotString(t, prediction, "probability_status") != "no_calibrated_event_probability_model" || !pilotBool(t, prediction, "abstention_required") {
		t.Errorf("Pilot prediction contract overstates probability readiness: %+v", prediction)
	}
	preregistration := pilotObject(t, protocol["pre_registration"])
	if pilotString(t, preregistration, "status") != "not_registered" || pilotString(t, preregistration, "registry") != "not_selected" ||
		pilotString(t, preregistration, "registration_id") != "" || len(pilotArray(t, preregistration["must_freeze"])) < 10 {
		t.Errorf("Pilot preregistration is incomplete or overstated: %+v", preregistration)
	}
	scoring := pilotObject(t, protocol["scoring_plan"])
	if pilotString(t, scoring, "status") != "not_approved" || pilotString(t, scoring, "primary_endpoint") != "not_selected" ||
		pilotString(t, scoring, "partial_match_rule") != "not_selected" || pilotString(t, scoring, "missing_data_rule") != "not_selected" ||
		pilotString(t, scoring, "multiple_comparison_rule") != "not_selected" || pilotBool(t, scoring, "single_accuracy_percentage_allowed") ||
		pilotBool(t, scoring, "pilot_effect_claim_allowed") {
		t.Errorf("Pilot scoring plan permits post-hoc or premature accuracy claims: %+v", scoring)
	}

	participants := pilotObject(t, protocol["participants"])
	if pilotInt(t, participants, "minimum_age") != 18 || len(pilotArray(t, participants["recruitment_channels"])) != 0 ||
		participants["compensation"] != nil || pilotString(t, participants, "compensation_status") != "not_approved" {
		t.Errorf("Pilot participant/recruitment contract is not blocked: %+v", participants)
	}
	data := pilotObject(t, protocol["data_governance"])
	if pilotBool(t, data, "collection_enabled") || pilotString(t, data, "collection_endpoint") != "none" ||
		pilotString(t, data, "database_schema") != "none" || data["retention_days"] != nil ||
		pilotString(t, data, "retention_status") != "not_approved" ||
		pilotString(t, data, "training_use") != "separate_opt_in_required_default_false" {
		t.Errorf("Pilot data governance permits collection or unresolved reuse: %+v", data)
	}
	legacy := pilotObject(t, data["legacy_feedback_audit"])
	if pilotString(t, legacy, "status") != "required_before_approval" ||
		!reflect.DeepEqual(pilotStrings(t, legacy["known_legacy_storage_fields"]), []string{"fortune_feedbacks.event_year", "fortune_feedbacks.event_category"}) ||
		!pilotBool(t, legacy, "public_request_fields_removed") || !pilotBool(t, legacy, "runtime_persistence_assignments_removed") ||
		len(pilotArray(t, legacy["required_actions"])) != 5 {
		t.Errorf("Pilot legacy feedback audit is incomplete: %+v", legacy)
	}
}

func TestProspectivePilotRequiresEveryApprovalBeforeLaunch(t *testing.T) {
	protocol, _ := loadProspectivePilotProtocol(t)
	wantGates := map[string]bool{
		"ethics_review": true, "privacy_impact_assessment": true, "scoring_plan": true, "power_and_sample": true,
		"security_review": true, "consent_finalization": true, "profile_and_claim_freeze": true,
		"legacy_feedback_audit": true, "incident_and_support_plan": true, "public_preregistration": true,
	}
	seen := make(map[string]bool, len(wantGates))
	for _, value := range pilotArray(t, protocol["approval_gates"]) {
		gate := pilotObject(t, value)
		id := pilotString(t, gate, "gate_id")
		if !wantGates[id] || seen[id] || pilotString(t, gate, "status") != "not_approved" ||
			!pilotBool(t, gate, "blocking") || strings.TrimSpace(pilotString(t, gate, "required_evidence")) == "" {
			t.Errorf("invalid Pilot approval gate: %+v", gate)
		}
		seen[id] = true
	}
	if !reflect.DeepEqual(seen, wantGates) {
		t.Errorf("Pilot approval gates = %+v, want %+v", seen, wantGates)
	}
	launch := pilotObject(t, protocol["launch_gate"])
	if pilotString(t, launch, "policy") != "all_approval_gates_must_be_approved" ||
		pilotInt(t, launch, "approved_gate_count") != 0 || pilotInt(t, launch, "required_gate_count") != 10 ||
		pilotBool(t, launch, "can_recruit") || pilotBool(t, launch, "can_collect_event_data") ||
		pilotBool(t, launch, "can_publish_predictive_accuracy") {
		t.Errorf("Pilot launch gate is not fail-closed: %+v", launch)
	}
	reporting := pilotObject(t, protocol["reporting"])
	if !pilotBool(t, reporting, "negative_results_required") || !pilotBool(t, reporting, "all_arms_reported") ||
		!pilotBool(t, reporting, "all_preregistered_primary_metrics_reported") ||
		!pilotBool(t, reporting, "independent_replication_required_before_predictive_claim") {
		t.Errorf("Pilot reporting contract permits selective publication: %+v", reporting)
	}
}

func TestProspectivePilotConsentIsDraftAndNotUsableForRecruitment(t *testing.T) {
	consent, err := os.ReadFile(prospectivePilotConsentPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := prospectivePilotSHA256(consent); got != prospectivePilotConsentSHA256 {
		t.Fatalf("Pilot consent SHA-256 = %s, want %s", got, prospectivePilotConsentSHA256)
	}
	protocol, _ := loadProspectivePilotProtocol(t)
	consentRef := pilotObject(t, protocol["consent"])
	if pilotString(t, consentRef, "document_path") != "research/pilot/participant-consent-v1.zh-CN.md" ||
		pilotString(t, consentRef, "document_sha256") != prospectivePilotConsentSHA256 ||
		pilotString(t, consentRef, "status") != "draft_not_approved_not_for_recruitment" ||
		!pilotBool(t, consentRef, "withdrawal_required") || !pilotBool(t, consentRef, "deletion_required") ||
		len(pilotArray(t, consentRef["missing_final_fields"])) != 8 {
		t.Errorf("Pilot consent reference is incomplete: %+v", consentRef)
	}
	for _, fragment := range []string{
		"本文件仅用于伦理、隐私和研究设计审查，不得用于招募、签署或收集任何参与者数据",
		"只有年满 18 周岁", "研究期内不向参与者展示封存的个体事件声明",
		"两名不知道声明来源、Profile 和对照组身份的审阅者", "第三名审阅者裁决",
		"不保证获得任何直接收益", "参加完全自愿", "要求导出、更正或删除",
		"默认不同意", "当前版本不得填写姓名、签名、日期或联系方式",
		"现有产品中的“准确/不准确/有帮助”等段落反馈不是本研究事件数据",
	} {
		if !bytes.Contains(consent, []byte(fragment)) {
			t.Errorf("Pilot consent missing required disclosure %q", fragment)
		}
	}
	for _, forbidden := range []string{"TODO", "TBD", "预测准确率 95%", "保证准确", "点击即同意"} {
		if bytes.Contains(consent, []byte(forbidden)) {
			t.Errorf("Pilot consent contains unsafe placeholder or claim %q", forbidden)
		}
	}
}

func TestPublicFeedbackCannotIngestProspectiveEventFields(t *testing.T) {
	publicFiles := []string{
		"../../model/dto.go",
		"../../handler/feedback.go",
		"../../../../vue/src/api/feedback.ts",
		"../../../../vue/src/components/ClassicalInterpretationPanel.vue",
		"../../../../API.md",
	}
	for _, path := range publicFiles {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"EventYear", "EventCategory", "event_year", "event_category"} {
			if bytes.Contains(raw, []byte(forbidden)) {
				t.Errorf("public feedback surface %s still accepts or documents %s", path, forbidden)
			}
		}
	}

	modelSource, err := os.ReadFile("../../model/fortune_feedback.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"Legacy event fields remain mapped only for privacy audit and disposal",
		"The public feedback API neither accepts nor populates them",
		"EventYear", "EventCategory", `json:"-"`,
	} {
		if !bytes.Contains(modelSource, []byte(required)) {
			t.Errorf("legacy feedback model missing privacy boundary %q", required)
		}
	}
	productionGoFiles := []string{}
	err = filepath.Walk("../../../..", func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") ||
			strings.Contains(path, string(filepath.Separator)+"vendor"+string(filepath.Separator)) {
			return nil
		}
		productionGoFiles = append(productionGoFiles, path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range productionGoFiles {
		if strings.HasSuffix(filepath.ToSlash(path), "src/internal/model/fortune_feedback.go") {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(raw, []byte("EventYear")) || bytes.Contains(raw, []byte("EventCategory")) ||
			bytes.Contains(raw, []byte(`json:"event_year"`)) || bytes.Contains(raw, []byte(`json:"event_category"`)) {
			t.Errorf("production Go file %s exposes a legacy event feedback field", path)
		}
	}

	api, err := os.ReadFile("../../../../API.md")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(api, []byte("Pilot 伦理、隐私与评分方案批准前，公开请求不接受或持久化事件年份、类别或结果")) {
		t.Error("API documentation does not disclose the pre-approval event collection boundary")
	}
}

func TestProspectivePilotResearchPlansAreSynchronized(t *testing.T) {
	marker := "第一百四十一项完成前瞻性 Pilot 协议、同意草案与事件采集关闭治理"
	for _, path := range []string{
		"../../../../docs/fortune-accuracy-research-plan.md",
		"../../../../docs/fortune-accuracy-roadmap.md",
		"../../../../docs/precision-test-plan.md",
	} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(raw), marker) != 1 {
			t.Errorf("research document %s must contain exactly one phase-141 marker", path)
		}
		for _, fragment := range []string{
			"prospective_fortune_pilot_protocol_v1", "draft_not_approved", "10 个阻断门槛",
			"event_year/event_category", "不得招募", "不得采集事件数据", "不得发布准确率",
		} {
			if !bytes.Contains(raw, []byte(fragment)) {
				t.Errorf("research document %s missing %q", path, fragment)
			}
		}
	}
}

func loadProspectivePilotProtocol(t testing.TB) (map[string]any, []byte) {
	t.Helper()
	raw, err := os.ReadFile(prospectivePilotProtocolPath)
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	var protocol map[string]any
	if err := decoder.Decode(&protocol); err != nil {
		t.Fatal(err)
	}
	return protocol, raw
}

func pilotObject(t testing.TB, value any) map[string]any {
	t.Helper()
	object, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("Pilot value is %T, want object", value)
	}
	return object
}

func pilotArray(t testing.TB, value any) []any {
	t.Helper()
	array, ok := value.([]any)
	if !ok {
		t.Fatalf("Pilot value is %T, want array", value)
	}
	return array
}

func pilotString(t testing.TB, object map[string]any, key string) string {
	t.Helper()
	value, ok := object[key].(string)
	if !ok {
		t.Fatalf("Pilot %s is %T, want string", key, object[key])
	}
	return value
}

func pilotBool(t testing.TB, object map[string]any, key string) bool {
	t.Helper()
	value, ok := object[key].(bool)
	if !ok {
		t.Fatalf("Pilot %s is %T, want bool", key, object[key])
	}
	return value
}

func pilotInt(t testing.TB, object map[string]any, key string) int {
	t.Helper()
	value, ok := object[key].(float64)
	if !ok || value != float64(int(value)) {
		t.Fatalf("Pilot %s is %v (%T), want integer", key, object[key], object[key])
	}
	return int(value)
}

func pilotStrings(t testing.TB, value any) []string {
	t.Helper()
	array := pilotArray(t, value)
	result := make([]string, len(array))
	for index, item := range array {
		text, ok := item.(string)
		if !ok {
			t.Fatalf("Pilot array item %d is %T, want string", index, item)
		}
		result[index] = text
	}
	return result
}

func pilotStringSet(t testing.TB, value any) map[string]bool {
	t.Helper()
	result := map[string]bool{}
	for _, text := range pilotStrings(t, value) {
		if text == "" || result[text] {
			t.Fatalf("Pilot string array contains empty or duplicate value %q", text)
		}
		result[text] = true
	}
	return result
}

func prospectivePilotSHA256(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}
