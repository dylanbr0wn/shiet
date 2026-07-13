package service_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/dylanbr0wn/shiet/internal/db"
	"github.com/dylanbr0wn/shiet/internal/seed"
	"github.com/dylanbr0wn/shiet/internal/service"
)

func TestExpectedTimeForDate_SeededWeekday(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()

	got, err := s.ExpectedTimeForDate(ctx, "2026-06-15") // Monday
	if err != nil {
		t.Fatal(err)
	}
	if got.ExpectedMinutes != 480 {
		t.Fatalf("expected_minutes = %d, want 480", got.ExpectedMinutes)
	}
	if got.Timezone != "America/Toronto" || got.WorkweekStart != "monday" {
		t.Fatalf("timezone/workweek = %q/%q", got.Timezone, got.WorkweekStart)
	}
	if got.Source != service.ExpectedTimeSourceWeekday {
		t.Fatalf("source = %q, want %q", got.Source, service.ExpectedTimeSourceWeekday)
	}
	if got.ExceptionKind != "" {
		t.Fatalf("exception_kind = %q, want empty", got.ExceptionKind)
	}
	if len(got.Windows) != 1 || got.Windows[0].StartMinutes != 9*60 || got.Windows[0].EndMinutes != 9*60+480 {
		t.Fatalf("windows = %+v, want [{540 1020}]", got.Windows)
	}
}

func TestExpectedTimeForDate_SeededWeekend(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()

	got, err := s.ExpectedTimeForDate(ctx, "2026-06-20") // Saturday
	if err != nil {
		t.Fatal(err)
	}
	if got.ExpectedMinutes != 0 {
		t.Fatalf("expected_minutes = %d, want 0", got.ExpectedMinutes)
	}
	if got.Source != service.ExpectedTimeSourceWeekday {
		t.Fatalf("source = %q, want %q", got.Source, service.ExpectedTimeSourceWeekday)
	}
	if len(got.Windows) != 0 {
		t.Fatalf("windows = %+v, want empty", got.Windows)
	}
}

func TestExpectedTimeForDate_HolidayException(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()

	_, err := s.UpsertScheduleException(ctx, service.ScheduleExceptionInput{
		Date:            "2026-06-15",
		Kind:            "holiday",
		ExpectedMinutes: 0,
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := s.ExpectedTimeForDate(ctx, "2026-06-15")
	if err != nil {
		t.Fatal(err)
	}
	if got.ExpectedMinutes != 0 {
		t.Fatalf("expected_minutes = %d, want 0", got.ExpectedMinutes)
	}
	if got.Source != service.ExpectedTimeSourceException {
		t.Fatalf("source = %q, want exception", got.Source)
	}
	if got.ExceptionKind != "holiday" {
		t.Fatalf("exception_kind = %q, want holiday", got.ExceptionKind)
	}
	if len(got.Windows) != 0 {
		t.Fatalf("windows = %+v, want empty", got.Windows)
	}
}

func TestExpectedTimeForDate_ChangedHoursException(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()

	_, err := s.UpsertScheduleException(ctx, service.ScheduleExceptionInput{
		Date:            "2026-06-16",
		Kind:            "changed_hours",
		ExpectedMinutes: 240,
		Windows:         []service.WorkingWindow{{StartMinutes: 10 * 60, EndMinutes: 14 * 60}},
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := s.ExpectedTimeForDate(ctx, "2026-06-16")
	if err != nil {
		t.Fatal(err)
	}
	if got.ExpectedMinutes != 240 {
		t.Fatalf("expected_minutes = %d, want 240", got.ExpectedMinutes)
	}
	if got.Source != service.ExpectedTimeSourceException || got.ExceptionKind != "changed_hours" {
		t.Fatalf("got source=%q kind=%q", got.Source, got.ExceptionKind)
	}
	if len(got.Windows) != 1 || got.Windows[0].StartMinutes != 600 || got.Windows[0].EndMinutes != 840 {
		t.Fatalf("windows = %+v", got.Windows)
	}
}

func TestExpectedTimeForDate_MissingCoveringTemplate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	conn, err := db.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if err := db.Migrate(conn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	// Core seeds a default schedule — wipe it so no covering template exists.
	if err := seed.Core(context.Background(), conn); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if _, err := conn.Exec(`DELETE FROM work_schedule`); err != nil {
		t.Fatalf("wipe schedules: %v", err)
	}
	s := service.New(conn)

	got, err := s.ExpectedTimeForDate(context.Background(), "2026-06-15")
	if err != nil {
		t.Fatal(err)
	}
	if got.ExpectedMinutes != 0 || len(got.Windows) != 0 {
		t.Fatalf("want zero expected with no windows, got %+v", got)
	}
	if got.Source != service.ExpectedTimeSourceWeekday {
		t.Fatalf("source = %q", got.Source)
	}
}

func TestExpectedTimeForDate_EffectiveRangeBoundary(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()

	scheds, err := s.ListWorkSchedules(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(scheds) != 1 {
		t.Fatalf("want 1 seeded schedule, got %d", len(scheds))
	}

	// Close default at 2026-07-01, open a new template with 240 Mon minutes.
	_, err = s.ReplaceActiveWorkSchedule(ctx, service.WorkScheduleInput{
		Timezone:      "America/Toronto",
		WorkweekStart: "monday",
		EffectiveFrom: "2026-07-01",
		Days: []service.WorkScheduleDayInput{
			{Weekday: "monday", ExpectedMinutes: 240, Windows: []service.WorkingWindow{{StartMinutes: 9 * 60, EndMinutes: 13 * 60}}},
			{Weekday: "tuesday", ExpectedMinutes: 0},
			{Weekday: "wednesday", ExpectedMinutes: 0},
			{Weekday: "thursday", ExpectedMinutes: 0},
			{Weekday: "friday", ExpectedMinutes: 0},
			{Weekday: "saturday", ExpectedMinutes: 0},
			{Weekday: "sunday", ExpectedMinutes: 0},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	before, err := s.ExpectedTimeForDate(ctx, "2026-06-30") // Tuesday under old schedule
	if err != nil {
		t.Fatal(err)
	}
	if before.ExpectedMinutes != 480 {
		t.Fatalf("before boundary minutes = %d, want 480", before.ExpectedMinutes)
	}

	onBoundary, err := s.ExpectedTimeForDate(ctx, "2026-07-01") // Wednesday under new schedule
	if err != nil {
		t.Fatal(err)
	}
	if onBoundary.ExpectedMinutes != 0 {
		t.Fatalf("new schedule Wed minutes = %d, want 0", onBoundary.ExpectedMinutes)
	}

	monday, err := s.ExpectedTimeForDate(ctx, "2026-07-06") // Monday under new
	if err != nil {
		t.Fatal(err)
	}
	if monday.ExpectedMinutes != 240 {
		t.Fatalf("new Mon minutes = %d, want 240", monday.ExpectedMinutes)
	}
}

func TestExpectedTimeForRange_InclusiveDays(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()

	got, err := s.ExpectedTimeForRange(ctx, "2026-06-15", "2026-06-21")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 7 {
		t.Fatalf("len = %d, want 7", len(got))
	}
	var sum int
	for _, d := range got {
		sum += d.ExpectedMinutes
	}
	if sum != 480*5 {
		t.Fatalf("week sum = %d, want %d", sum, 480*5)
	}
	if got[0].Date != "2026-06-15" || got[6].Date != "2026-06-21" {
		t.Fatalf("range ends = %s..%s", got[0].Date, got[6].Date)
	}
}

func TestExpectedTimeForRange_MissingCoveringGap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gap.db")
	conn, err := db.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if err := db.Migrate(conn); err != nil {
		t.Fatal(err)
	}
	if err := seed.Core(context.Background(), conn); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	svc := service.New(conn)

	list, err := svc.ListWorkSchedules(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("want 1 schedule, got %d", len(list))
	}
	// Half-open: close at 2026-06-20 so Fri 19 is last covered day.
	if _, err := conn.ExecContext(ctx, `UPDATE work_schedule SET effective_to = ? WHERE id = ?`, "2026-06-20", list[0].ID); err != nil {
		t.Fatal(err)
	}

	got, err := svc.ExpectedTimeForRange(ctx, "2026-06-18", "2026-06-22")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 5 {
		t.Fatalf("len = %d, want 5", len(got))
	}
	// Thu 18 + Fri 19 covered; Sat–Mon uncovered (no template).
	if got[0].ExpectedMinutes != 480 || got[1].ExpectedMinutes != 480 {
		t.Fatalf("covered days = %+v", got[:2])
	}
	for i := 2; i < 5; i++ {
		if got[i].ExpectedMinutes != 0 || len(got[i].Windows) != 0 {
			t.Fatalf("gap day[%d] = %+v, want zero", i, got[i])
		}
	}
}
