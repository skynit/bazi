package handler

import (
	"fmt"
	"strings"

	"bazi/internal/model"
	"bazi/internal/service/bazi"
)

func resolveZiWeiRequestBirth(req model.ChartRequest) (*bazi.NormalizedBirth, bool, error) {
	gender, ok := parseGender(req.Gender)
	if !ok {
		return nil, false, fmt.Errorf("gender must be male or female")
	}
	req.Gender = gender
	set, err := bazi.CalculateBirthCandidates(&bazi.BaziService{}, birthInputFromChartRequest(req))
	if err != nil {
		return nil, false, err
	}
	preparedSet := &preparedCandidateSet{
		center:     &preparedChart{normalized: set.Center, result: set.CenterResult},
		candidates: set,
	}
	prepared, _, err := selectPreparedCandidate(preparedSet, strings.TrimSpace(req.CandidateID))
	if err != nil {
		return nil, set.RequiresCandidateSelection, err
	}
	return prepared.normalized, set.RequiresCandidateSelection, nil
}

func resolveStoredZiWeiBirth(chart *model.BirthChart) (*bazi.NormalizedBirth, error) {
	if chart == nil {
		return nil, fmt.Errorf("birth chart is nil")
	}
	normalized, _, _, err := calculateStoredChartBirth(&bazi.BaziService{}, chart)
	if err != nil {
		return nil, fmt.Errorf("resolve normalized birth for ziwei: %w", err)
	}
	return normalized, nil
}
