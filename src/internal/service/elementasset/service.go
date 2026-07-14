package elementasset

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strings"

	"bazi/internal/model"
)

type Repository interface {
	ListActive() ([]model.ElementAsset, error)
}

type Service struct {
	repo Repository
}

type Query struct {
	Element     string
	Scene       string
	Orientation string
	Season      string
	TimePeriod  string
	Seed        string
	Limit       int
	ExcludeIDs  map[uint]struct{}
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// Select returns deterministic weighted matches. Strict presentation filters
// are relaxed in stages so a page always has a usable fallback when the library
// is still small.
func (s *Service) Select(query Query) ([]model.ElementAsset, error) {
	if s == nil || s.repo == nil || query.Limit <= 0 {
		return []model.ElementAsset{}, nil
	}
	assets, err := s.repo.ListActive()
	if err != nil {
		return nil, err
	}

	tiers := []func(model.ElementAsset) bool{
		func(a model.ElementAsset) bool { return matches(a, query, true, true, true) },
		func(a model.ElementAsset) bool { return matches(a, query, true, true, false) },
		func(a model.ElementAsset) bool { return matches(a, query, true, false, false) },
		func(a model.ElementAsset) bool { return matches(a, query, false, false, false) },
	}

	selected := make([]model.ElementAsset, 0, query.Limit)
	seen := make(map[uint]struct{}, query.Limit)
	for _, tier := range tiers {
		candidates := make([]model.ElementAsset, 0)
		for _, asset := range assets {
			if _, excluded := query.ExcludeIDs[asset.ID]; excluded {
				continue
			}
			if _, exists := seen[asset.ID]; exists || !tier(asset) {
				continue
			}
			candidates = append(candidates, asset)
		}
		rank(candidates, query.Seed)
		for _, asset := range candidates {
			selected = append(selected, asset)
			seen[asset.ID] = struct{}{}
			if len(selected) >= query.Limit {
				return selected, nil
			}
		}
	}
	return selected, nil
}

func (s *Service) BuildBlessingSet(chartID uint, date, primary, secondary, avoid string, actionElements []string) (model.BlessingAssetSet, error) {
	seed := fmt.Sprintf("%d:%s", chartID, date)
	set := model.BlessingAssetSet{
		Ritual:  []model.ElementAsset{},
		Actions: []model.ElementAsset{},
		Gallery: []model.ElementAsset{},
	}

	heroes, err := s.Select(Query{Element: primary, Scene: "hero", Orientation: "landscape", Seed: seed + ":hero", Limit: 1})
	if err != nil {
		return set, err
	}
	if len(heroes) > 0 {
		set.Hero = &heroes[0]
	}

	excluded := map[uint]struct{}{}
	for index, element := range uniqueElements(primary, secondary, avoid) {
		items, selectErr := s.Select(Query{Element: element, Scene: "object", Seed: fmt.Sprintf("%s:ritual:%d", seed, index), Limit: 1, ExcludeIDs: excluded})
		if selectErr != nil {
			return set, selectErr
		}
		if len(items) > 0 {
			set.Ritual = append(set.Ritual, items[0])
			excluded[items[0].ID] = struct{}{}
		}
	}

	for index, element := range actionElements {
		if strings.TrimSpace(element) == "" {
			element = primary
		}
		items, selectErr := s.Select(Query{Element: element, Scene: "object", Orientation: "square", Seed: fmt.Sprintf("%s:action:%d", seed, index), Limit: 1})
		if selectErr != nil {
			return set, selectErr
		}
		if len(items) > 0 {
			set.Actions = append(set.Actions, items[0])
		}
	}

	gallery, err := s.Select(Query{Seed: seed + ":gallery", Limit: 10})
	if err != nil {
		return set, err
	}
	set.Gallery = gallery
	return set, nil
}

func matches(asset model.ElementAsset, query Query, scene, orientation, context bool) bool {
	if query.Element != "" && asset.Element != query.Element && asset.SecondaryElement != query.Element {
		return false
	}
	if scene && query.Scene != "" && asset.Scene != query.Scene && asset.Scene != "general" {
		return false
	}
	if orientation && query.Orientation != "" && asset.Orientation != query.Orientation {
		return false
	}
	if context {
		if query.Season != "" && asset.Season != "all" && asset.Season != query.Season {
			return false
		}
		if query.TimePeriod != "" && asset.TimePeriod != "all" && asset.TimePeriod != query.TimePeriod {
			return false
		}
	}
	return true
}

func rank(assets []model.ElementAsset, seed string) {
	sort.SliceStable(assets, func(i, j int) bool {
		return weightedHash(seed, assets[i]) < weightedHash(seed, assets[j])
	})
}

func weightedHash(seed string, asset model.ElementAsset) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(seed + ":" + asset.Key))
	weight := asset.Weight
	if weight < 1 {
		weight = 1
	}
	return h.Sum64() / uint64(weight)
}

func uniqueElements(elements ...string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(elements))
	for _, element := range elements {
		element = strings.TrimSpace(element)
		if element == "" {
			continue
		}
		if _, exists := seen[element]; exists {
			continue
		}
		seen[element] = struct{}{}
		out = append(out, element)
	}
	return out
}
