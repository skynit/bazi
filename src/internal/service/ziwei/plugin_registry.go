package ziwei

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// PluginRequirement is the exact plugin identity and version embedded in a
// calculation profile. Slice order is execution order and part of the hash.
type PluginRequirement struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

// PluginDescriptor declares dependency and conflict constraints. Dependencies
// must appear earlier in the profile manifest; no runtime sorting is allowed.
type PluginDescriptor struct {
	ID            string
	Version       string
	DependsOn     []string
	ConflictsWith []string
}

type StarPlugin interface {
	Descriptor() PluginDescriptor
	AdditionalStars(palace PalaceInfo, palaceIdx int) []StarOutput
	MutatePalace(palace *PalaceInfo) bool
}

type pluginExecutionPlan struct {
	manifest []PluginRequirement
	hash     string
	plugins  []StarPlugin
}

func newPluginCatalog(plugins []StarPlugin) (map[string]StarPlugin, error) {
	catalog := make(map[string]StarPlugin, len(plugins))
	for _, plugin := range plugins {
		if plugin == nil {
			return nil, fmt.Errorf("ziwei plugin is nil")
		}
		descriptor := plugin.Descriptor()
		descriptor.ID = strings.TrimSpace(descriptor.ID)
		descriptor.Version = strings.TrimSpace(descriptor.Version)
		if descriptor.ID == "" || descriptor.Version == "" {
			return nil, fmt.Errorf("ziwei plugin ID and version are required")
		}
		if _, exists := catalog[descriptor.ID]; exists {
			return nil, fmt.Errorf("duplicate ziwei plugin %q", descriptor.ID)
		}
		catalog[descriptor.ID] = plugin
	}
	return catalog, nil
}

func buildPluginExecutionPlan(profile CalculationProfile, catalog map[string]StarPlugin) (*pluginExecutionPlan, error) {
	manifest, err := normalizePluginManifest(profile.PluginManifest)
	if err != nil {
		return nil, fmt.Errorf("profile %s plugin manifest: %w", profile.ID, err)
	}
	allSelected := make(map[string]struct{}, len(manifest))
	for _, requirement := range manifest {
		allSelected[requirement.ID] = struct{}{}
	}

	resolved := make(map[string]struct{}, len(manifest))
	plugins := make([]StarPlugin, 0, len(manifest))
	for _, requirement := range manifest {
		plugin, ok := catalog[requirement.ID]
		if !ok {
			return nil, fmt.Errorf("profile %s requires missing ziwei plugin %s@%s", profile.ID, requirement.ID, requirement.Version)
		}
		descriptor := plugin.Descriptor()
		if strings.TrimSpace(descriptor.ID) != requirement.ID || strings.TrimSpace(descriptor.Version) != requirement.Version {
			return nil, fmt.Errorf("profile %s requires ziwei plugin %s@%s but catalog provides %s@%s",
				profile.ID, requirement.ID, requirement.Version, descriptor.ID, descriptor.Version)
		}
		for _, dependency := range descriptor.DependsOn {
			dependency = strings.TrimSpace(dependency)
			if _, ok := resolved[dependency]; !ok {
				return nil, fmt.Errorf("ziwei plugin %s requires %s earlier in the profile manifest", requirement.ID, dependency)
			}
		}
		for _, conflict := range descriptor.ConflictsWith {
			conflict = strings.TrimSpace(conflict)
			if _, ok := allSelected[conflict]; ok {
				return nil, fmt.Errorf("ziwei plugin %s conflicts with %s", requirement.ID, conflict)
			}
		}
		resolved[requirement.ID] = struct{}{}
		plugins = append(plugins, plugin)
	}

	return &pluginExecutionPlan{
		manifest: manifest,
		hash:     pluginManifestHash(manifest),
		plugins:  plugins,
	}, nil
}

func normalizePluginManifest(manifest []PluginRequirement) ([]PluginRequirement, error) {
	normalized := make([]PluginRequirement, 0, len(manifest))
	seen := make(map[string]struct{}, len(manifest))
	for _, requirement := range manifest {
		requirement.ID = strings.TrimSpace(requirement.ID)
		requirement.Version = strings.TrimSpace(requirement.Version)
		if requirement.ID == "" || requirement.Version == "" {
			return nil, fmt.Errorf("plugin ID and version are required")
		}
		if _, ok := seen[requirement.ID]; ok {
			return nil, fmt.Errorf("duplicate plugin %q", requirement.ID)
		}
		seen[requirement.ID] = struct{}{}
		normalized = append(normalized, requirement)
	}
	return normalized, nil
}

func pluginManifestHash(manifest []PluginRequirement) string {
	parts := make([]string, 0, len(manifest))
	for _, requirement := range manifest {
		parts = append(parts, requirement.ID+"@"+requirement.Version)
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\n")))
	return hex.EncodeToString(sum[:])
}

func (p *pluginExecutionPlan) Apply(chart *ZiWeiChart) {
	if p == nil || chart == nil {
		return
	}
	for _, plugin := range p.plugins {
		for i := range chart.Palaces {
			for _, star := range plugin.AdditionalStars(chart.Palaces[i], i) {
				star.Name = strings.TrimSpace(star.Name)
				if star.Name == "" {
					continue
				}
				if star.Type == "" {
					star.Type = "soft"
				}
				if star.Scope == "" {
					star.Scope = "origin"
				}
				chart.Palaces[i].Stars = append(chart.Palaces[i].Stars, star)
			}
			plugin.MutatePalace(&chart.Palaces[i])
		}
	}
}

func clonePluginManifest(manifest []PluginRequirement) []PluginRequirement {
	cloned := make([]PluginRequirement, len(manifest))
	copy(cloned, manifest)
	return cloned
}
