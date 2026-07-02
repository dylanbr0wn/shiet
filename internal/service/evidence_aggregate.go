package service

import (
	"context"
	"sort"
)

// AggregateEvidence collects activity evidence from all wired providers for a
// gap interval. Individual provider failures are best-effort and skipped so a
// single integration outage does not block suggestions.
func AggregateEvidence(ctx context.Context, providers []EvidenceProvider, window TimeWindow) ([]ActivityEvidence, error) {
	if len(providers) == 0 {
		return nil, nil
	}

	out := make([]ActivityEvidence, 0)
	for _, provider := range providers {
		items, err := provider.FetchEvidence(ctx, window)
		if err != nil {
			continue
		}
		out = append(out, items...)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Start.Equal(out[j].Start) {
			return out[i].End.Before(out[j].End)
		}
		return out[i].Start.Before(out[j].Start)
	})
	return out, nil
}

func (s *Service) fetchGapEvidence(ctx context.Context, window TimeWindow) ([]ActivityEvidence, error) {
	if s.evidence == nil || len(s.evidence.Providers) == 0 {
		return nil, nil
	}
	return AggregateEvidence(ctx, s.evidence.Providers, window)
}
