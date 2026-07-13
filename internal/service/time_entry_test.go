package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dylanbr0wn/shiet/internal/service"
)

func TestCreateTimeEntry_IsConfirmedAndListable(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()

	periods, err := s.ListPeriods(ctx)
	if err != nil {
		t.Fatal(err)
	}
	pid := periods[0].ID

	entry, err := s.CreateTimeEntry(ctx, service.TimeEntryInput{
		PeriodID:     pid,
		Day:          "2026-06-01",
		StartMinutes: 9 * 60,
		EndMinutes:   10*60 + 30,
		Description:  "New block",
	})
	if err != nil {
		t.Fatal(err)
	}
	if entry.PeriodID != pid || entry.LocalWorkDate != "2026-06-01" || entry.Description != "New block" {
		t.Fatalf("unexpected entry: %+v", entry)
	}
	if entry.Attestation != "confirmed" {
		t.Fatalf("want attestation confirmed, got %q", entry.Attestation)
	}
	if entry.Method != "" {
		t.Fatalf("user create should not stamp method, got %q", entry.Method)
	}
	if entry.DurationMinutes != 90 {
		t.Fatalf("want duration 90, got %d", entry.DurationMinutes)
	}

	wantStart := time.Date(2026, 6, 1, 13, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, 6, 1, 14, 30, 0, 0, time.UTC)
	gotStart, err := time.Parse(time.RFC3339, entry.Start)
	if err != nil {
		t.Fatal(err)
	}
	gotEnd, err := time.Parse(time.RFC3339, entry.End)
	if err != nil {
		t.Fatal(err)
	}
	if !gotStart.Equal(wantStart) || !gotEnd.Equal(wantEnd) {
		t.Fatalf("unexpected UTC span: %s to %s", entry.Start, entry.End)
	}

	listed, err := s.ListTimeEntries(ctx, pid)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].ID != entry.ID {
		t.Fatalf("entry was not listed: %+v", listed)
	}

	got, err := s.GetTimeEntry(ctx, entry.ID, pid)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != entry.ID || got.Attestation != "confirmed" {
		t.Fatalf("get mismatch: %+v", got)
	}
}

func TestCreateGapTimeEntry_StampsMethod(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()

	periods, err := s.ListPeriods(ctx)
	if err != nil {
		t.Fatal(err)
	}
	pid := periods[0].ID

	entry, err := s.CreateGapTimeEntry(ctx, service.TimeEntryInput{
		PeriodID:     pid,
		Day:          "2026-06-01",
		StartMinutes: 9 * 60,
		EndMinutes:   10 * 60,
		Description:  "Confirmed gap",
	})
	if err != nil {
		t.Fatal(err)
	}
	if entry.Attestation != "confirmed" {
		t.Fatalf("want confirmed, got %q", entry.Attestation)
	}
	if entry.Method != "gap_fill" {
		t.Fatalf("want method gap_fill, got %q", entry.Method)
	}
}

func TestCreateTimeEntryValidation(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()

	periods, err := s.ListPeriods(ctx)
	if err != nil {
		t.Fatal(err)
	}
	pid := periods[0].ID

	_, err = s.CreateTimeEntry(ctx, service.TimeEntryInput{
		PeriodID:     pid,
		Day:          "2026-06-01",
		StartMinutes: 10 * 60,
		EndMinutes:   10 * 60,
	})
	if err == nil {
		t.Fatal("expected invalid range error")
	}

	_, err = s.CreateTimeEntry(ctx, service.TimeEntryInput{
		PeriodID:     pid,
		Day:          "2026-06-15",
		StartMinutes: 9 * 60,
		EndMinutes:   10 * 60,
	})
	if err == nil {
		t.Fatal("expected out-of-period error")
	}
}

func TestUpdateAndDeleteTimeEntry(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()

	periods, err := s.ListPeriods(ctx)
	if err != nil {
		t.Fatal(err)
	}
	pid := periods[0].ID

	entry, err := s.CreateTimeEntry(ctx, service.TimeEntryInput{
		PeriodID:     pid,
		Day:          "2026-06-01",
		StartMinutes: 9 * 60,
		EndMinutes:   10 * 60,
		Description:  "Temp",
	})
	if err != nil {
		t.Fatal(err)
	}

	updated, err := s.UpdateTimeEntry(ctx, service.TimeEntryUpdateInput{
		ID: entry.ID,
		TimeEntryInput: service.TimeEntryInput{
			PeriodID:     pid,
			Day:          "2026-06-02",
			StartMinutes: 11 * 60,
			EndMinutes:   12 * 60,
			Description:  "Moved",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.LocalWorkDate != "2026-06-02" || updated.Description != "Moved" || updated.Attestation != "confirmed" {
		t.Fatalf("unexpected update: %+v", updated)
	}

	if err := s.DeleteTimeEntry(ctx, service.TimeEntryDeleteInput{ID: entry.ID, PeriodID: pid}); err != nil {
		t.Fatal(err)
	}
	listed, err := s.ListTimeEntries(ctx, pid)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 0 {
		t.Fatalf("not deleted: %+v", listed)
	}
	err = s.DeleteTimeEntry(ctx, service.TimeEntryDeleteInput{ID: entry.ID, PeriodID: pid})
	if !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
