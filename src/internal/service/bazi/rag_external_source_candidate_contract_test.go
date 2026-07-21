package bazi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
)

const (
	ragExternalCandidateRegistryPath   = "../../../../research/rag/bazi-external-source-candidates-v1.json"
	ragExternalCandidateRegistrySHA256 = "758974aa243b415a23bddbde995a597d82b575afff05000cda536130375a4dd9"
)

type ragExternalCandidateRegistry struct {
	Schema          string                      `json:"schema"`
	Version         string                      `json:"version"`
	Status          string                      `json:"status"`
	ObservedAt      string                      `json:"observed_at"`
	Work            string                      `json:"work"`
	Purpose         string                      `json:"purpose"`
	GlobalPolicy    ragExternalCandidatePolicy  `json:"global_policy"`
	Candidates      []ragExternalCandidate      `json:"candidates"`
	RejectedSources []ragExternalRejectedSource `json:"rejected_sources"`
}

type ragExternalCandidatePolicy struct {
	DownloadDoesNotImplyIngestion            bool `json:"download_does_not_imply_ingestion"`
	RemoteHashDoesNotReplaceLocalSHA256      bool `json:"remote_hash_does_not_replace_local_sha256"`
	ProviderMetadataIsNotBibliographicReview bool `json:"provider_metadata_is_not_bibliographic_adjudication"`
	CrossSourceTextMatchRequired             bool `json:"cross_source_text_match_required"`
	TwoIndependentReviewersRequired          bool `json:"two_independent_reviewers_required"`
	RuntimeIngestionAllowed                  bool `json:"runtime_ingestion_allowed"`
	ClaimEligible                            bool `json:"claim_eligible"`
	PublishableAccuracy                      bool `json:"publishable_accuracy"`
}

type ragExternalCandidate struct {
	CandidateID          string                       `json:"candidate_id"`
	Provider             string                       `json:"provider"`
	SourceType           string                       `json:"source_type"`
	Edition              string                       `json:"edition"`
	Status               string                       `json:"status"`
	LandingURL           string                       `json:"landing_url"`
	RemoteManifestScheme string                       `json:"remote_manifest_scheme"`
	RemoteManifestSHA256 string                       `json:"remote_manifest_sha256"`
	License              ragExternalCandidateLicense  `json:"license"`
	Gates                ragExternalCandidateGates    `json:"gates"`
	Volumes              []ragExternalCandidateVolume `json:"volumes"`
}

type ragExternalCandidateLicense struct {
	Status               string   `json:"status"`
	UnderlyingWork       string   `json:"underlying_work"`
	DigitalContributions string   `json:"digital_contributions"`
	RightsField          string   `json:"rights_field"`
	LicenseURL           string   `json:"license_url"`
	EvidenceURLs         []string `json:"evidence_urls"`
}

type ragExternalCandidateGates struct {
	CompleteStructureObserved bool `json:"complete_structure_observed"`
	LicenseTermsResolved      bool `json:"license_terms_resolved"`
	LocalArtifactFrozen       bool `json:"local_artifact_frozen"`
	IndependenceVerified      bool `json:"independence_verified"`
	BibliographyAdjudicated   bool `json:"bibliography_adjudicated"`
	PageMappingVerified       bool `json:"page_mapping_verified"`
	ClaimSupportReviewed      bool `json:"claim_support_reviewed"`
	RuntimeIngestionAllowed   bool `json:"runtime_ingestion_allowed"`
	ClaimEligible             bool `json:"claim_eligible"`
}

type ragExternalCandidateVolume struct {
	Volume         int    `json:"volume"`
	Title          string `json:"title"`
	StableID       string `json:"stable_id"`
	StableRevision int    `json:"stable_revision"`
	Timestamp      string `json:"timestamp"`
	SizeBytes      int64  `json:"size_bytes"`
	RemoteSHA1     string `json:"remote_sha1"`
	RemoteMD5      string `json:"remote_md5"`
	SourceURL      string `json:"source_url"`
	DownloadURL    string `json:"download_url"`
	LocalSHA256    string `json:"local_sha256"`
	Retrieved      bool   `json:"retrieved"`
}

type ragExternalRejectedSource struct {
	SourceID   string `json:"source_id"`
	URL        string `json:"url"`
	RemoteSHA1 string `json:"remote_sha1"`
	Reason     string `json:"reason"`
}

func TestRAGExternalSourceCandidateRegistryContract(t *testing.T) {
	raw, err := os.ReadFile(ragExternalCandidateRegistryPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := ragExternalCandidateSHA256(raw); got != ragExternalCandidateRegistrySHA256 {
		t.Fatalf("external source candidate registry SHA-256 = %s, want %s", got, ragExternalCandidateRegistrySHA256)
	}
	var registry ragExternalCandidateRegistry
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&registry); err != nil {
		t.Fatalf("decode external source candidate registry: %v", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		t.Fatalf("external source registry must contain exactly one JSON document: %v", err)
	}
	if registry.Schema != "bazi_external_source_candidate_registry_v1" || registry.Version != "2026-07-17.1" ||
		registry.Status != "candidate_only_no_runtime_ingestion" || registry.ObservedAt != "2026-07-17" ||
		registry.Work != "三命通會" || strings.TrimSpace(registry.Purpose) == "" {
		t.Fatalf("unexpected external source registry identity: %+v", registry)
	}
	policy := registry.GlobalPolicy
	if !policy.DownloadDoesNotImplyIngestion || !policy.RemoteHashDoesNotReplaceLocalSHA256 ||
		!policy.ProviderMetadataIsNotBibliographicReview || !policy.CrossSourceTextMatchRequired ||
		!policy.TwoIndependentReviewersRequired || policy.RuntimeIngestionAllowed ||
		policy.ClaimEligible || policy.PublishableAccuracy {
		t.Fatalf("external source policy must fail closed: %+v", policy)
	}

	expectedManifests := map[string]string{
		"sanming-siku-wikisource-12vol-v1": "80d14200816a4bc1fd3b26c73facb73f3abd16d5e61f241f279e38fb8ddc2fcd",
		"sanming-siku-ia-060660-12scan-v1": "22d7abce9b76daed59a1a930b3c20b4194d7187866a555826e70fcc42025e300",
		"sanming-siku-ia-060564-12scan-v1": "1198c6231e9888f4793efcedc927f310c8216d3e2aeb5b38fa6837ebcd6286a6",
	}
	if len(registry.Candidates) != len(expectedManifests) {
		t.Fatalf("candidate count = %d, want %d", len(registry.Candidates), len(expectedManifests))
	}
	seenCandidates := map[string]bool{}
	for _, candidate := range registry.Candidates {
		expectedManifest, ok := expectedManifests[candidate.CandidateID]
		if !ok || seenCandidates[candidate.CandidateID] {
			t.Fatalf("unknown or duplicate candidate %q", candidate.CandidateID)
		}
		seenCandidates[candidate.CandidateID] = true
		if candidate.Edition != "欽定四庫全書本" || candidate.RemoteManifestSHA256 != expectedManifest ||
			!ragExternalValidSHA256(candidate.RemoteManifestSHA256) || len(candidate.Volumes) != 12 ||
			!candidate.Gates.CompleteStructureObserved || candidate.Gates.LocalArtifactFrozen ||
			candidate.Gates.IndependenceVerified || candidate.Gates.BibliographyAdjudicated ||
			candidate.Gates.PageMappingVerified || candidate.Gates.ClaimSupportReviewed ||
			candidate.Gates.RuntimeIngestionAllowed || candidate.Gates.ClaimEligible {
			t.Fatalf("candidate gates or identity invalid for %q: %+v", candidate.CandidateID, candidate)
		}
		ragExternalValidateVolumes(t, candidate)
		if got := ragExternalCandidateManifestSHA256(t, candidate); got != candidate.RemoteManifestSHA256 {
			t.Fatalf("candidate manifest SHA-256 for %q = %s, want %s", candidate.CandidateID, got, candidate.RemoteManifestSHA256)
		}
		if candidate.Provider == "zh.wikisource.org" {
			if candidate.SourceType != "versioned_digital_transcription" ||
				candidate.Status != "candidate_for_manual_textual_comparison" ||
				!candidate.Gates.LicenseTermsResolved || candidate.License.Status != "terms_identified_attribution_required" ||
				candidate.License.UnderlyingWork != "PD-old" ||
				candidate.License.DigitalContributions != "CC BY-SA 4.0 and GFDL" ||
				candidate.License.LicenseURL == "" {
				t.Fatalf("Wikisource candidate license mismatch: %+v", candidate)
			}
			continue
		}
		if candidate.Provider != "archive.org" || candidate.SourceType != "twelve_volume_text_pdf_scan_set" ||
			candidate.Status != "candidate_blocked_license_metadata_missing" || candidate.Gates.LicenseTermsResolved ||
			candidate.License.Status != "provider_rights_and_licenseurl_missing" ||
			candidate.License.RightsField != "" || candidate.License.LicenseURL != "" {
			t.Fatalf("Internet Archive candidate must remain license-blocked: %+v", candidate)
		}
	}

	if len(registry.RejectedSources) != 2 {
		t.Fatalf("rejected source count = %d, want 2", len(registry.RejectedSources))
	}
	rejectedText := string(raw)
	for _, required := range []string{
		"zhwikisource-sanming-standard-incomplete",
		"cabbbd89dfddb586db59ca2a9a4f5c5a642d871e",
		"repository-local-sanming-pdf-incomplete",
		"卷十至卷十二",
	} {
		if !strings.Contains(rejectedText, required) {
			t.Fatalf("rejected source evidence missing %q", required)
		}
	}
}

func TestRAGExternalCandidatesAreNotRuntimeRegistered(t *testing.T) {
	for _, sourcePath := range []string{
		"../localrag/index.go",
		"../localrag/retriever.go",
		"../interpretation/bazi.go",
		"../../model/dto.go",
		"../../../cmd/bazi-rag-index/main.go",
	} {
		raw, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatal(err)
		}
		for _, candidateID := range []string{
			"sanming-siku-wikisource-12vol-v1",
			"sanming-siku-ia-060660-12scan-v1",
			"sanming-siku-ia-060564-12scan-v1",
		} {
			if bytes.Contains(raw, []byte(candidateID)) {
				t.Fatalf("draft external source %q leaked into runtime source %s", candidateID, sourcePath)
			}
		}
	}
}

func ragExternalValidateVolumes(t *testing.T, candidate ragExternalCandidate) {
	t.Helper()
	seen := map[int]bool{}
	for index, volume := range candidate.Volumes {
		wantVolume := index + 1
		if volume.Volume != wantVolume || seen[volume.Volume] || strings.TrimSpace(volume.Title) == "" ||
			strings.TrimSpace(volume.StableID) == "" || strings.TrimSpace(volume.Timestamp) == "" || volume.SizeBytes <= 0 ||
			!ragExternalValidHex(volume.RemoteSHA1, 40) || volume.LocalSHA256 != "" || volume.Retrieved {
			t.Fatalf("invalid volume %d for %q: %+v", wantVolume, candidate.CandidateID, volume)
		}
		seen[volume.Volume] = true
		sourceURL, err := url.Parse(volume.SourceURL)
		if err != nil || sourceURL.Scheme != "https" || sourceURL.Host != candidate.Provider {
			t.Fatalf("invalid source URL for %q volume %d: %q", candidate.CandidateID, volume.Volume, volume.SourceURL)
		}
		if candidate.Provider == "zh.wikisource.org" {
			if volume.StableRevision <= 0 || volume.RemoteMD5 != "" || volume.DownloadURL != "" {
				t.Fatalf("invalid versioned transcription volume: %+v", volume)
			}
			continue
		}
		downloadURL, err := url.Parse(volume.DownloadURL)
		if err != nil || downloadURL.Scheme != "https" || downloadURL.Host != "archive.org" ||
			path.Base(downloadURL.Path) != volume.StableID+".pdf" || volume.StableRevision != 0 ||
			!ragExternalValidHex(volume.RemoteMD5, 32) {
			t.Fatalf("invalid scan volume: %+v", volume)
		}
	}
}

func ragExternalCandidateManifestSHA256(t *testing.T, candidate ragExternalCandidate) string {
	t.Helper()
	var manifest strings.Builder
	for _, volume := range candidate.Volumes {
		if candidate.Provider == "zh.wikisource.org" {
			manifest.WriteString(volume.Title)
			manifest.WriteByte('\t')
			manifest.WriteString(volume.StableID)
			manifest.WriteByte('\t')
			manifest.WriteString(strconv.Itoa(volume.StableRevision))
			manifest.WriteByte('\t')
			manifest.WriteString(strconv.FormatInt(volume.SizeBytes, 10))
			manifest.WriteByte('\t')
			manifest.WriteString(volume.Timestamp)
			manifest.WriteByte('\t')
			manifest.WriteString(volume.RemoteSHA1)
			manifest.WriteByte('\n')
			continue
		}
		manifest.WriteString(volume.StableID)
		manifest.WriteByte('\t')
		manifest.WriteString(volume.Title)
		manifest.WriteByte('\t')
		downloadURL, err := url.Parse(volume.DownloadURL)
		if err != nil {
			t.Fatal(err)
		}
		manifest.WriteString(path.Base(downloadURL.Path))
		manifest.WriteByte('\t')
		manifest.WriteString(strconv.FormatInt(volume.SizeBytes, 10))
		manifest.WriteByte('\t')
		manifest.WriteString(volume.RemoteSHA1)
		manifest.WriteByte('\t')
		manifest.WriteString(volume.RemoteMD5)
		manifest.WriteByte('\n')
	}
	return ragExternalCandidateSHA256([]byte(manifest.String()))
}

func ragExternalValidSHA256(value string) bool {
	return ragExternalValidHex(value, 64)
}

func ragExternalValidHex(value string, length int) bool {
	if len(value) != length || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func ragExternalCandidateSHA256(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
