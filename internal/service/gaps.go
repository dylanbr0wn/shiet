package service

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// Interval is a half-open UTC time span [Start, End).
type Interval struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// DayTimeline is one day's working window with what occupies it and the gaps
// left uncovered. All instants are UTC; Date/Tz record the local bucketing.
type DayTimeline struct {
	Date         string     `json:"date"` // YYYY-MM-DD, the local bucketing date
	Tz           string     `json:"tz"`   // active IANA segment for this date
	WindowStart  time.Time  `json:"windowStart"`
	WindowEnd    time.Time  `json:"windowEnd"`
	Events       []Interval `json:"events"`       // event-occupied spans, clipped to the window
	Filled       []Interval `json:"filled"`       // gap-fill spans, clipped to the window
	Gaps         []Interval `json:"gaps"`         // uncovered spans within the window
	CoveredHours float64    `json:"coveredHours"` // union of events+fills within the window
	GapHours     float64    `json:"gapHours"`
}

// ComputeGaps builds the per-day timelines for a period: for each date it
// resolves ExpectedTime working windows (schedule timezone), records the
// period's active timezone segment for local bucketing, and subtracts active
// timed events + gap fills to find the uncovered gaps.
//
// Dates with no expected minutes / empty windows produce an empty timeline
// (no unconditional daily target). Multiple windows are unioned.
//
// All-day events do not occupy time here (they are resolved via the review
// queue, not the timeline). Declined and soft-hidden events are excluded.
func (s *Service) ComputeGaps(ctx context.Context, periodID int64) ([]DayTimeline, error) {
	period, err := s.GetPeriod(ctx, periodID)
	if err != nil {
		return nil, err
	}
	segs, err := s.ListTzSegments(ctx, periodID)
	if err != nil {
		return nil, err
	}
	if len(segs) == 0 {
		return nil, fmt.Errorf("period %d has no timezone segment", periodID)
	}
	events, err := s.ListEvents(ctx, periodID)
	if err != nil {
		return nil, err
	}
	fills, err := s.ListTimeEntries(ctx, periodID)
	if err != nil {
		return nil, err
	}
	expected, err := s.ExpectedTimeForRange(ctx, period.StartDate, period.EndDate)
	if err != nil {
		return nil, err
	}

	// Collect occupying spans once: timed, active, non-declined events; and fills.
	var eventSpans, fillSpans []Interval
	for _, e := range events {
		if e.AllDay || e.Status == "declined" || e.Start == nil || e.End == nil {
			continue
		}
		eventSpans = append(eventSpans, Interval{Start: e.Start.UTC(), End: e.End.UTC()})
	}
	for _, f := range fills {
		fillSpans = append(fillSpans, Interval{Start: parseTime(f.Start), End: parseTime(f.End)})
	}

	locCache := map[string]*time.Location{}
	out := make([]DayTimeline, 0, len(expected))
	for _, et := range expected {
		dateStr := et.Date
		seg := activeSegment(segs, dateStr)

		working, err := workingWindowsUTC(et, locCache)
		if err != nil {
			return nil, err
		}
		if len(working) == 0 {
			out = append(out, DayTimeline{
				Date:         dateStr,
				Tz:           seg.IanaTz,
				Events:       []Interval{},
				Filled:       []Interval{},
				Gaps:         []Interval{},
				CoveredHours: 0,
				GapHours:     0,
			})
			continue
		}

		var dayEvents, dayFills []Interval
		for _, win := range working {
			dayEvents = append(dayEvents, clipAll(eventSpans, win)...)
			dayFills = append(dayFills, clipAll(fillSpans, win)...)
		}
		occupied := union(append(append([]Interval{}, dayEvents...), dayFills...))
		var gaps []Interval
		for _, win := range working {
			gaps = append(gaps, subtract(win, occupied)...)
		}

		out = append(out, DayTimeline{
			Date:         dateStr,
			Tz:           seg.IanaTz,
			WindowStart:  working[0].Start,
			WindowEnd:    working[len(working)-1].End,
			Events:       dayEvents,
			Filled:       dayFills,
			Gaps:         gaps,
			CoveredHours: totalHours(occupied),
			GapHours:     totalHours(gaps),
		})
	}
	return out, nil
}

// workingWindowsUTC converts ExpectedTime local windows into sorted, disjoint
// UTC intervals on that calendar date in the schedule timezone. When expected
// minutes are set but windows are omitted, derives a single window starting at
// 09:00 local with that duration.
func workingWindowsUTC(et ExpectedTime, locCache map[string]*time.Location) ([]Interval, error) {
	if et.ExpectedMinutes <= 0 {
		return nil, nil
	}
	windows := et.Windows
	if len(windows) == 0 {
		start := 9 * 60
		windows = []WorkingWindow{{
			StartMinutes: start,
			EndMinutes:   start + et.ExpectedMinutes,
		}}
	}
	iana := et.Timezone
	if iana == "" {
		return nil, fmt.Errorf("expected time for %s: missing schedule timezone", et.Date)
	}
	loc, err := loadLoc(locCache, iana)
	if err != nil {
		return nil, err
	}
	day, err := time.ParseInLocation("2006-01-02", et.Date, loc)
	if err != nil {
		return nil, fmt.Errorf("expected time date %q: %w", et.Date, err)
	}
	y, m, d := day.Date()
	raw := make([]Interval, 0, len(windows))
	for _, w := range windows {
		if w.EndMinutes <= w.StartMinutes {
			continue
		}
		ws := time.Date(y, m, d, 0, w.StartMinutes, 0, 0, loc)
		we := time.Date(y, m, d, 0, w.EndMinutes, 0, 0, loc)
		raw = append(raw, Interval{Start: ws.UTC(), End: we.UTC()})
	}
	return union(raw), nil
}

// activeSegment returns the segment governing dateStr: the latest segment whose
// effective_from_date <= dateStr, else the earliest (segs is date-ascending).
func activeSegment(segs []TzSegment, dateStr string) TzSegment {
	active := segs[0]
	for _, seg := range segs {
		if seg.EffectiveFromDate <= dateStr {
			active = seg
		} else {
			break
		}
	}
	return active
}

func loadLoc(cache map[string]*time.Location, iana string) (*time.Location, error) {
	if loc, ok := cache[iana]; ok {
		return loc, nil
	}
	loc, err := time.LoadLocation(iana)
	if err != nil {
		return nil, fmt.Errorf("load timezone %q: %w", iana, err)
	}
	cache[iana] = loc
	return loc, nil
}

func parseDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, time.UTC)
}

// ── interval algebra (all UTC, half-open) ─────────────────────────────

// clip returns the portion of iv inside [lo,hi), or ok=false if disjoint/empty.
func clip(iv Interval, lo, hi time.Time) (Interval, bool) {
	s := iv.Start
	if s.Before(lo) {
		s = lo
	}
	e := iv.End
	if e.After(hi) {
		e = hi
	}
	if !s.Before(e) {
		return Interval{}, false
	}
	return Interval{Start: s, End: e}, true
}

func clipAll(ivs []Interval, win Interval) []Interval {
	out := []Interval{}
	for _, iv := range ivs {
		if c, ok := clip(iv, win.Start, win.End); ok {
			out = append(out, c)
		}
	}
	return out
}

// union merges overlapping/adjacent intervals into sorted, disjoint spans.
func union(ivs []Interval) []Interval {
	if len(ivs) == 0 {
		return []Interval{}
	}
	sort.Slice(ivs, func(i, j int) bool { return ivs[i].Start.Before(ivs[j].Start) })
	out := []Interval{ivs[0]}
	for _, iv := range ivs[1:] {
		last := &out[len(out)-1]
		if !iv.Start.After(last.End) { // overlap or touch
			if iv.End.After(last.End) {
				last.End = iv.End
			}
			continue
		}
		out = append(out, iv)
	}
	return out
}

// subtract returns win minus the (assumed sorted, disjoint) occupied spans.
func subtract(win Interval, occupied []Interval) []Interval {
	out := []Interval{}
	cursor := win.Start
	for _, o := range occupied {
		if o.End.Before(win.Start) || o.Start.After(win.End) {
			continue
		}
		if o.Start.After(cursor) {
			out = append(out, Interval{Start: cursor, End: o.Start})
		}
		if o.End.After(cursor) {
			cursor = o.End
		}
	}
	if cursor.Before(win.End) {
		out = append(out, Interval{Start: cursor, End: win.End})
	}
	return out
}

func totalHours(ivs []Interval) float64 {
	var d time.Duration
	for _, iv := range ivs {
		d += iv.End.Sub(iv.Start)
	}
	return d.Hours()
}
