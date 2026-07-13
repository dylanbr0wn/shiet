package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dylanbr0wn/shiet/internal/service"
)

func defaultDays(minutes int, withWindow bool) []service.WorkScheduleDayInput {
	days := make([]service.WorkScheduleDayInput, 0, 7)
	for _, wd := range []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"} {
		d := service.WorkScheduleDayInput{Weekday: wd, ExpectedMinutes: 0}
		if wd != "saturday" && wd != "sunday" {
			d.ExpectedMinutes = minutes
			if withWindow {
				d.Windows = []service.WorkingWindow{{StartMinutes: 9 * 60, EndMinutes: 9*60 + minutes}}
			}
		}
		days = append(days, d)
	}
	return days
}

func TestWorkScheduleCRUD_ListAndReplace(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()

	list, err := s.ListWorkSchedules(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Timezone != "America/Toronto" || list[0].EffectiveTo != "" {
		t.Fatalf("seeded schedule unexpected: %+v", list[0])
	}

	got, err := s.GetWorkSchedule(ctx, list[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Days) != 7 {
		t.Fatalf("want 7 days, got %d", len(got.Days))
	}

	next, err := s.ReplaceActiveWorkSchedule(ctx, service.WorkScheduleInput{
		Timezone:      "America/Vancouver",
		WorkweekStart: "monday",
		EffectiveFrom: "2026-08-01",
		Days:          defaultDays(300, true),
	})
	if err != nil {
		t.Fatal(err)
	}
	if next.Timezone != "America/Vancouver" || next.EffectiveFrom != "2026-08-01" {
		t.Fatalf("new schedule: %+v", next)
	}

	list, err = s.ListWorkSchedules(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2 schedule versions, got %d", len(list))
	}
	if list[0].EffectiveTo != "2026-08-01" {
		t.Fatalf("prior effective_to = %q, want 2026-08-01", list[0].EffectiveTo)
	}
}

func TestScheduleExceptionCRUD_UpsertAndDelete(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()

	exc, err := s.UpsertScheduleException(ctx, service.ScheduleExceptionInput{
		Date:            "2026-06-18",
		Kind:            "leave",
		ExpectedMinutes: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if exc.Kind != "leave" || exc.Date != "2026-06-18" {
		t.Fatalf("exception: %+v", exc)
	}

	exc, err = s.UpsertScheduleException(ctx, service.ScheduleExceptionInput{
		Date:            "2026-06-18",
		Kind:            "changed_hours",
		ExpectedMinutes: 120,
		Windows:         []service.WorkingWindow{{StartMinutes: 13 * 60, EndMinutes: 15 * 60}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if exc.Kind != "changed_hours" || exc.ExpectedMinutes != 120 || len(exc.Windows) != 1 {
		t.Fatalf("updated exception: %+v", exc)
	}

	all, err := s.ListScheduleExceptions(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Fatalf("want 1 exception, got %d", len(all))
	}

	if err := s.DeleteScheduleException(ctx, "2026-06-18"); err != nil {
		t.Fatal(err)
	}
	_, err = s.GetScheduleExceptionByDate(ctx, "2026-06-18")
	if !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("want ErrNotFound after delete, got %v", err)
	}
}
