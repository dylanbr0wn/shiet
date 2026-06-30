package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dylanbr0wn/clockr/internal/ai"
	"github.com/dylanbr0wn/clockr/internal/db/sqlc"
)

// applyAISuggestion auto-categorizes a new event when memory did not match and
// a local/cloud model is configured. Failures are best-effort and never abort sync.
func (s *Service) applyAISuggestion(ctx context.Context, q *sqlc.Queries, periodID int64, inc IncomingEvent) error {
	has, err := hasCategory(ctx, q, periodID, inc.GoogleEventID, inc.InstanceID)
	if err != nil || has {
		return err
	}

	baseURL, model, ok := s.loadAIConfig(ctx)
	if !ok {
		return nil
	}

	categories, err := s.ListCategories(ctx)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(categories))
	byName := make(map[string]int64, len(categories))
	for _, category := range categories {
		names = append(names, category.Name)
		byName[category.Name] = category.ID
	}

	local, _ := ai.ClassifyEndpoint(baseURL)
	privacy, err := s.loadPrivacyFields(ctx)
	if err != nil {
		return err
	}

	probeCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	client := ai.NewClient(baseURL, "")
	suggested, err := ai.SuggestCategory(
		probeCtx,
		client,
		model,
		names,
		buildEventContext(inc),
		local,
		privacy,
	)
	if err != nil {
		return nil
	}

	categoryID, ok := byName[suggested]
	if !ok {
		return nil
	}

	if _, err := q.UpsertOverlay(ctx, sqlc.UpsertOverlayParams{
		PeriodID:      periodID,
		GoogleEventID: inc.GoogleEventID,
		InstanceID:    inc.InstanceID,
		CategoryID:    sql.NullInt64{Int64: categoryID, Valid: true},
		Kind:          overlayKindCategory,
	}); err != nil {
		return fmt.Errorf("apply ai overlay: %w", err)
	}
	return nil
}

