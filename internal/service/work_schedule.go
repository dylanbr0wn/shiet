package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dylanbr0wn/shiet/internal/db/sqlc"
)

var (
	validWeekdays = map[string]struct{}{
		"monday": {}, "tuesday": {}, "wednesday": {}, "thursday": {},
		"friday": {}, "saturday": {}, "sunday": {},
	}
	validExceptionKinds = map[string]struct{}{
		"holiday": {}, "leave": {}, "changed_hours": {},
	}
)

// WorkSchedule is an effective-dated weekday template.
type WorkSchedule struct {
	ID            int64              `json:"id"`
	Timezone      string             `json:"timezone"`
	WorkweekStart string             `json:"workweekStart"`
	EffectiveFrom string             `json:"effectiveFrom"`
	EffectiveTo   string             `json:"effectiveTo,omitempty"`
	Days          []WorkScheduleDay  `json:"days"`
}

// WorkScheduleDay is one weekday row on a schedule version.
type WorkScheduleDay struct {
	Weekday         string          `json:"weekday"`
	ExpectedMinutes int             `json:"expectedMinutes"`
	Windows         []WorkingWindow `json:"windows"`
}

// WorkScheduleDayInput is the writable shape for a weekday row.
type WorkScheduleDayInput struct {
	Weekday         string          `json:"weekday"`
	ExpectedMinutes int             `json:"expectedMinutes"`
	Windows         []WorkingWindow `json:"windows"`
}

// WorkScheduleInput creates a new schedule version (and closes the prior open range).
type WorkScheduleInput struct {
	Timezone      string                 `json:"timezone"`
	WorkweekStart string                 `json:"workweekStart"`
	EffectiveFrom string                 `json:"effectiveFrom"`
	Days          []WorkScheduleDayInput `json:"days"`
}

// ScheduleException is a date-keyed override of expected time.
type ScheduleException struct {
	ID              int64           `json:"id"`
	Date            string          `json:"date"`
	Kind            string          `json:"kind"`
	ExpectedMinutes int             `json:"expectedMinutes"`
	Windows         []WorkingWindow `json:"windows"`
}

// ScheduleExceptionInput creates or replaces an exception for a date.
type ScheduleExceptionInput struct {
	Date            string          `json:"date"`
	Kind            string          `json:"kind"`
	ExpectedMinutes int             `json:"expectedMinutes"`
	Windows         []WorkingWindow `json:"windows"`
}

func (s *Service) ListWorkSchedules(ctx context.Context) ([]WorkSchedule, error) {
	rows, err := s.q.ListWorkSchedules(ctx)
	if err != nil {
		return nil, mapErr("list work schedules", err)
	}
	out := make([]WorkSchedule, len(rows))
	for i, r := range rows {
		detail, err := s.loadWorkSchedule(ctx, r)
		if err != nil {
			return nil, err
		}
		out[i] = detail
	}
	return out, nil
}

func (s *Service) GetWorkSchedule(ctx context.Context, id int64) (WorkSchedule, error) {
	row, err := s.q.GetWorkSchedule(ctx, id)
	if err != nil {
		return WorkSchedule{}, mapErr("get work schedule", err)
	}
	return s.loadWorkSchedule(ctx, row)
}

// ReplaceActiveWorkSchedule closes any open schedule at EffectiveFrom (half-open)
// and inserts a new version with the given weekday template.
func (s *Service) ReplaceActiveWorkSchedule(ctx context.Context, input WorkScheduleInput) (WorkSchedule, error) {
	tz := strings.TrimSpace(input.Timezone)
	if tz == "" {
		return WorkSchedule{}, fmt.Errorf("replace work schedule: timezone is required")
	}
	if _, err := time.LoadLocation(tz); err != nil {
		return WorkSchedule{}, fmt.Errorf("replace work schedule: invalid timezone %q", tz)
	}
	ww := strings.ToLower(strings.TrimSpace(input.WorkweekStart))
	if _, ok := validWeekdays[ww]; !ok {
		return WorkSchedule{}, fmt.Errorf("replace work schedule: invalid workweek_start %q", input.WorkweekStart)
	}
	from := strings.TrimSpace(input.EffectiveFrom)
	if _, err := time.Parse("2006-01-02", from); err != nil {
		return WorkSchedule{}, fmt.Errorf("replace work schedule: invalid effective_from %q", input.EffectiveFrom)
	}
	if err := validateScheduleDays(input.Days); err != nil {
		return WorkSchedule{}, fmt.Errorf("replace work schedule: %w", err)
	}
	for i := range input.Days {
		input.Days[i].Weekday = strings.ToLower(strings.TrimSpace(input.Days[i].Weekday))
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return WorkSchedule{}, mapErr("replace work schedule", err)
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.q.WithTx(tx)

	existing, err := qtx.ListWorkSchedules(ctx)
	if err != nil {
		return WorkSchedule{}, mapErr("replace work schedule", err)
	}
	for _, row := range existing {
		// Close open-ended or ranges that extend past the new start.
		if !row.EffectiveTo.Valid || row.EffectiveTo.String > from {
			if row.EffectiveFrom >= from {
				return WorkSchedule{}, fmt.Errorf("replace work schedule: existing schedule %d starts on or after %s", row.ID, from)
			}
			if _, err := qtx.UpdateWorkScheduleEffectiveTo(ctx, sqlc.UpdateWorkScheduleEffectiveToParams{
				EffectiveTo: sql.NullString{String: from, Valid: true},
				ID:          row.ID,
			}); err != nil {
				return WorkSchedule{}, mapErr("replace work schedule", err)
			}
		}
	}

	created, err := qtx.CreateWorkSchedule(ctx, sqlc.CreateWorkScheduleParams{
		Timezone:      tz,
		WorkweekStart: ww,
		EffectiveFrom: from,
		EffectiveTo:   sql.NullString{},
	})
	if err != nil {
		return WorkSchedule{}, mapErr("replace work schedule", err)
	}
	if err := insertScheduleDays(ctx, qtx, created.ID, input.Days); err != nil {
		return WorkSchedule{}, err
	}
	if err := tx.Commit(); err != nil {
		return WorkSchedule{}, mapErr("replace work schedule", err)
	}
	return s.GetWorkSchedule(ctx, created.ID)
}

func (s *Service) ListScheduleExceptions(ctx context.Context) ([]ScheduleException, error) {
	rows, err := s.q.ListScheduleExceptions(ctx)
	if err != nil {
		return nil, mapErr("list schedule exceptions", err)
	}
	out := make([]ScheduleException, len(rows))
	for i, r := range rows {
		detail, err := s.loadScheduleException(ctx, r)
		if err != nil {
			return nil, err
		}
		out[i] = detail
	}
	return out, nil
}

func (s *Service) GetScheduleExceptionByDate(ctx context.Context, date string) (ScheduleException, error) {
	date = strings.TrimSpace(date)
	row, err := s.q.GetScheduleExceptionByDate(ctx, date)
	if err != nil {
		return ScheduleException{}, mapErr("get schedule exception", err)
	}
	return s.loadScheduleException(ctx, row)
}

func (s *Service) UpsertScheduleException(ctx context.Context, input ScheduleExceptionInput) (ScheduleException, error) {
	date := strings.TrimSpace(input.Date)
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return ScheduleException{}, fmt.Errorf("upsert schedule exception: invalid date %q", input.Date)
	}
	kind := strings.ToLower(strings.TrimSpace(input.Kind))
	if _, ok := validExceptionKinds[kind]; !ok {
		return ScheduleException{}, fmt.Errorf("upsert schedule exception: invalid kind %q", input.Kind)
	}
	if input.ExpectedMinutes < 0 {
		return ScheduleException{}, fmt.Errorf("upsert schedule exception: expected_minutes must be >= 0")
	}
	if kind != "changed_hours" && len(input.Windows) > 0 {
		return ScheduleException{}, fmt.Errorf("upsert schedule exception: windows only allowed for changed_hours")
	}
	if err := validateWindows(input.Windows); err != nil {
		return ScheduleException{}, fmt.Errorf("upsert schedule exception: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ScheduleException{}, mapErr("upsert schedule exception", err)
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.q.WithTx(tx)

	existing, err := qtx.GetScheduleExceptionByDate(ctx, date)
	var id int64
	if errors.Is(err, sql.ErrNoRows) {
		row, cerr := qtx.CreateScheduleException(ctx, sqlc.CreateScheduleExceptionParams{
			Date:            date,
			Kind:            kind,
			ExpectedMinutes: int64(input.ExpectedMinutes),
		})
		if cerr != nil {
			return ScheduleException{}, mapErr("upsert schedule exception", cerr)
		}
		id = row.ID
	} else if err != nil {
		return ScheduleException{}, mapErr("upsert schedule exception", err)
	} else {
		id = existing.ID
		if _, uerr := qtx.UpdateScheduleException(ctx, sqlc.UpdateScheduleExceptionParams{
			Kind:            kind,
			ExpectedMinutes: int64(input.ExpectedMinutes),
			ID:              id,
		}); uerr != nil {
			return ScheduleException{}, mapErr("upsert schedule exception", uerr)
		}
		if derr := qtx.DeleteScheduleExceptionWindows(ctx, id); derr != nil {
			return ScheduleException{}, mapErr("upsert schedule exception", derr)
		}
	}

	for _, w := range input.Windows {
		if _, err := qtx.CreateScheduleExceptionWindow(ctx, sqlc.CreateScheduleExceptionWindowParams{
			ScheduleExceptionID: id,
			StartMinutes:        int64(w.StartMinutes),
			EndMinutes:          int64(w.EndMinutes),
		}); err != nil {
			return ScheduleException{}, mapErr("upsert schedule exception", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return ScheduleException{}, mapErr("upsert schedule exception", err)
	}
	return s.GetScheduleExceptionByDate(ctx, date)
}

func (s *Service) DeleteScheduleException(ctx context.Context, date string) error {
	date = strings.TrimSpace(date)
	row, err := s.q.GetScheduleExceptionByDate(ctx, date)
	if err != nil {
		return mapErr("delete schedule exception", err)
	}
	if err := s.q.DeleteScheduleException(ctx, row.ID); err != nil {
		return mapErr("delete schedule exception", err)
	}
	return nil
}

func (s *Service) loadWorkSchedule(ctx context.Context, row sqlc.WorkSchedule) (WorkSchedule, error) {
	days, err := s.q.ListWorkScheduleDays(ctx, row.ID)
	if err != nil {
		return WorkSchedule{}, mapErr("load work schedule days", err)
	}
	outDays := make([]WorkScheduleDay, len(days))
	for i, d := range days {
		wins, err := s.listScheduleDayWindows(ctx, d.ID)
		if err != nil {
			return WorkSchedule{}, err
		}
		outDays[i] = WorkScheduleDay{
			Weekday:         d.Weekday,
			ExpectedMinutes: int(d.ExpectedMinutes),
			Windows:         wins,
		}
	}
	effTo := ""
	if row.EffectiveTo.Valid {
		effTo = row.EffectiveTo.String
	}
	return WorkSchedule{
		ID:            row.ID,
		Timezone:      row.Timezone,
		WorkweekStart: row.WorkweekStart,
		EffectiveFrom: row.EffectiveFrom,
		EffectiveTo:   effTo,
		Days:          outDays,
	}, nil
}

func (s *Service) loadScheduleException(ctx context.Context, row sqlc.ScheduleException) (ScheduleException, error) {
	wins, err := s.listExceptionWindows(ctx, row.ID)
	if err != nil {
		return ScheduleException{}, err
	}
	return ScheduleException{
		ID:              row.ID,
		Date:            row.Date,
		Kind:            row.Kind,
		ExpectedMinutes: int(row.ExpectedMinutes),
		Windows:         wins,
	}, nil
}

func insertScheduleDays(ctx context.Context, q *sqlc.Queries, scheduleID int64, days []WorkScheduleDayInput) error {
	for _, d := range days {
		row, err := q.CreateWorkScheduleDay(ctx, sqlc.CreateWorkScheduleDayParams{
			WorkScheduleID:  scheduleID,
			Weekday:         d.Weekday,
			ExpectedMinutes: int64(d.ExpectedMinutes),
		})
		if err != nil {
			return mapErr("create work schedule day", err)
		}
		for _, w := range d.Windows {
			if _, err := q.CreateWorkScheduleWindow(ctx, sqlc.CreateWorkScheduleWindowParams{
				WorkScheduleDayID: row.ID,
				StartMinutes:      int64(w.StartMinutes),
				EndMinutes:        int64(w.EndMinutes),
			}); err != nil {
				return mapErr("create work schedule window", err)
			}
		}
	}
	return nil
}

func validateScheduleDays(days []WorkScheduleDayInput) error {
	if len(days) != 7 {
		return fmt.Errorf("days must include all 7 weekdays")
	}
	seen := map[string]struct{}{}
	for _, d := range days {
		wd := strings.ToLower(strings.TrimSpace(d.Weekday))
		if _, ok := validWeekdays[wd]; !ok {
			return fmt.Errorf("invalid weekday %q", d.Weekday)
		}
		if _, ok := seen[wd]; ok {
			return fmt.Errorf("duplicate weekday %q", wd)
		}
		seen[wd] = struct{}{}
		if d.ExpectedMinutes < 0 {
			return fmt.Errorf("expected_minutes must be >= 0 for %s", wd)
		}
		if err := validateWindows(d.Windows); err != nil {
			return fmt.Errorf("%s: %w", wd, err)
		}
		d.Weekday = wd
	}
	return nil
}

func validateWindows(windows []WorkingWindow) error {
	for _, w := range windows {
		if w.StartMinutes < 0 || w.StartMinutes >= 1440 {
			return fmt.Errorf("window start_minutes out of range")
		}
		if w.EndMinutes <= 0 || w.EndMinutes > 1440 {
			return fmt.Errorf("window end_minutes out of range")
		}
		if w.EndMinutes <= w.StartMinutes {
			return fmt.Errorf("window end_minutes must be > start_minutes")
		}
	}
	return nil
}
