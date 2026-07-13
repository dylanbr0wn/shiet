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

// ExpectedTimeSource identifies where a day's expected minutes came from.
const (
	ExpectedTimeSourceWeekday   = "weekday"
	ExpectedTimeSourceException = "exception"
)

// WorkingWindow is a local-time window within a day, in minutes from midnight.
type WorkingWindow struct {
	StartMinutes int `json:"startMinutes"`
	EndMinutes   int `json:"endMinutes"`
}

// ExpectedTime is the resolved expected work for one local calendar date.
type ExpectedTime struct {
	Date             string          `json:"date"`
	ExpectedMinutes  int             `json:"expectedMinutes"`
	Windows          []WorkingWindow `json:"windows"`
	Source           string          `json:"source"`
	ExceptionKind    string          `json:"exceptionKind,omitempty"`
	Timezone         string          `json:"timezone,omitempty"`
	WorkweekStart    string          `json:"workweekStart,omitempty"`
}

// ExpectedTimeForDate resolves expected minutes and windows for a local YYYY-MM-DD date.
func (s *Service) ExpectedTimeForDate(ctx context.Context, date string) (ExpectedTime, error) {
	date = strings.TrimSpace(date)
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return ExpectedTime{}, fmt.Errorf("expected time for date: invalid date %q", date)
	}

	empty := ExpectedTime{
		Date:            date,
		ExpectedMinutes: 0,
		Windows:         []WorkingWindow{},
		Source:          ExpectedTimeSourceWeekday,
	}

	sched, err := s.q.GetWorkScheduleForDate(ctx, date)
	if errors.Is(err, sql.ErrNoRows) {
		return empty, nil
	}
	if err != nil {
		return ExpectedTime{}, mapErr("expected time for date", err)
	}

	empty.Timezone = sched.Timezone
	empty.WorkweekStart = sched.WorkweekStart

	if exc, err := s.q.GetScheduleExceptionByDate(ctx, date); err == nil {
		wins, werr := s.listExceptionWindows(ctx, exc.ID)
		if werr != nil {
			return ExpectedTime{}, werr
		}
		return ExpectedTime{
			Date:            date,
			ExpectedMinutes: int(exc.ExpectedMinutes),
			Windows:         wins,
			Source:          ExpectedTimeSourceException,
			ExceptionKind:   exc.Kind,
			Timezone:        sched.Timezone,
			WorkweekStart:   sched.WorkweekStart,
		}, nil
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return ExpectedTime{}, mapErr("expected time for date", err)
	}

	weekday, err := weekdayName(date)
	if err != nil {
		return ExpectedTime{}, fmt.Errorf("expected time for date: %w", err)
	}

	days, err := s.q.ListWorkScheduleDays(ctx, sched.ID)
	if err != nil {
		return ExpectedTime{}, mapErr("expected time for date", err)
	}
	var day *sqlc.WorkScheduleDay
	for i := range days {
		if days[i].Weekday == weekday {
			day = &days[i]
			break
		}
	}
	if day == nil {
		return empty, nil
	}

	wins, err := s.listScheduleDayWindows(ctx, day.ID)
	if err != nil {
		return ExpectedTime{}, err
	}
	return ExpectedTime{
		Date:            date,
		ExpectedMinutes: int(day.ExpectedMinutes),
		Windows:         wins,
		Source:          ExpectedTimeSourceWeekday,
		Timezone:        sched.Timezone,
		WorkweekStart:   sched.WorkweekStart,
	}, nil
}

// ExpectedTimeForRange resolves each local date in [start, end] inclusive.
func (s *Service) ExpectedTimeForRange(ctx context.Context, start, end string) ([]ExpectedTime, error) {
	start = strings.TrimSpace(start)
	end = strings.TrimSpace(end)
	startDay, err := time.Parse("2006-01-02", start)
	if err != nil {
		return nil, fmt.Errorf("expected time for range: invalid start %q", start)
	}
	endDay, err := time.Parse("2006-01-02", end)
	if err != nil {
		return nil, fmt.Errorf("expected time for range: invalid end %q", end)
	}
	if endDay.Before(startDay) {
		return nil, fmt.Errorf("expected time for range: end before start")
	}

	out := make([]ExpectedTime, 0)
	for d := startDay; !d.After(endDay); d = d.AddDate(0, 0, 1) {
		got, err := s.ExpectedTimeForDate(ctx, d.Format("2006-01-02"))
		if err != nil {
			return nil, err
		}
		out = append(out, got)
	}
	return out, nil
}

func (s *Service) listScheduleDayWindows(ctx context.Context, dayID int64) ([]WorkingWindow, error) {
	rows, err := s.q.ListWorkScheduleWindows(ctx, dayID)
	if err != nil {
		return nil, mapErr("list schedule windows", err)
	}
	out := make([]WorkingWindow, len(rows))
	for i, r := range rows {
		out[i] = WorkingWindow{StartMinutes: int(r.StartMinutes), EndMinutes: int(r.EndMinutes)}
	}
	return out, nil
}

func (s *Service) listExceptionWindows(ctx context.Context, exceptionID int64) ([]WorkingWindow, error) {
	rows, err := s.q.ListScheduleExceptionWindows(ctx, exceptionID)
	if err != nil {
		return nil, mapErr("list exception windows", err)
	}
	out := make([]WorkingWindow, len(rows))
	for i, r := range rows {
		out[i] = WorkingWindow{StartMinutes: int(r.StartMinutes), EndMinutes: int(r.EndMinutes)}
	}
	return out, nil
}

func weekdayName(date string) (string, error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", err
	}
	switch t.Weekday() {
	case time.Monday:
		return "monday", nil
	case time.Tuesday:
		return "tuesday", nil
	case time.Wednesday:
		return "wednesday", nil
	case time.Thursday:
		return "thursday", nil
	case time.Friday:
		return "friday", nil
	case time.Saturday:
		return "saturday", nil
	case time.Sunday:
		return "sunday", nil
	default:
		return "", fmt.Errorf("unknown weekday")
	}
}
