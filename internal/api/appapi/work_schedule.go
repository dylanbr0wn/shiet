package appapi

import (
	"context"
	"time"

	"connectrpc.com/connect"
	appv1 "github.com/dylanbr0wn/shiet/gen/shiet/app/v1"
	"github.com/dylanbr0wn/shiet/internal/service"
)

type WorkScheduleService struct{ service *service.Service }

func (s *WorkScheduleService) ListWorkSchedules(ctx context.Context, _ *connect.Request[appv1.ListWorkSchedulesRequest]) (*connect.Response[appv1.ListWorkSchedulesResponse], error) {
	items, err := s.service.ListWorkSchedules(ctx)
	if err != nil {
		return nil, mapServiceError(err)
	}
	out := make([]*appv1.WorkSchedule, len(items))
	for i := range items {
		out[i] = toProtoWorkSchedule(items[i])
	}
	return connect.NewResponse(&appv1.ListWorkSchedulesResponse{Schedules: out}), nil
}

func (s *WorkScheduleService) GetWorkSchedule(ctx context.Context, req *connect.Request[appv1.GetWorkScheduleRequest]) (*connect.Response[appv1.GetWorkScheduleResponse], error) {
	if err := requireID(req.Msg.Id, "id"); err != nil {
		return nil, err
	}
	item, err := s.service.GetWorkSchedule(ctx, req.Msg.Id)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(&appv1.GetWorkScheduleResponse{Schedule: toProtoWorkSchedule(item)}), nil
}

func (s *WorkScheduleService) ReplaceActiveWorkSchedule(ctx context.Context, req *connect.Request[appv1.ReplaceActiveWorkScheduleRequest]) (*connect.Response[appv1.ReplaceActiveWorkScheduleResponse], error) {
	if req.Msg.Timezone == "" {
		return nil, invalidArgument("timezone is required")
	}
	if req.Msg.WorkweekStart == "" {
		return nil, invalidArgument("workweek_start is required")
	}
	if _, err := time.Parse("2006-01-02", req.Msg.EffectiveFrom); err != nil {
		return nil, invalidArgument("effective_from must be a YYYY-MM-DD date")
	}
	days := make([]service.WorkScheduleDayInput, len(req.Msg.Days))
	for i, d := range req.Msg.Days {
		days[i] = service.WorkScheduleDayInput{
			Weekday:         d.Weekday,
			ExpectedMinutes: int(d.ExpectedMinutes),
			Windows:         toServiceWindows(d.Windows),
		}
	}
	item, err := s.service.ReplaceActiveWorkSchedule(ctx, service.WorkScheduleInput{
		Timezone:      req.Msg.Timezone,
		WorkweekStart: req.Msg.WorkweekStart,
		EffectiveFrom: req.Msg.EffectiveFrom,
		Days:          days,
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(&appv1.ReplaceActiveWorkScheduleResponse{Schedule: toProtoWorkSchedule(item)}), nil
}

func (s *WorkScheduleService) ListScheduleExceptions(ctx context.Context, _ *connect.Request[appv1.ListScheduleExceptionsRequest]) (*connect.Response[appv1.ListScheduleExceptionsResponse], error) {
	items, err := s.service.ListScheduleExceptions(ctx)
	if err != nil {
		return nil, mapServiceError(err)
	}
	out := make([]*appv1.ScheduleException, len(items))
	for i := range items {
		out[i] = toProtoScheduleException(items[i])
	}
	return connect.NewResponse(&appv1.ListScheduleExceptionsResponse{Exceptions: out}), nil
}

func (s *WorkScheduleService) UpsertScheduleException(ctx context.Context, req *connect.Request[appv1.UpsertScheduleExceptionRequest]) (*connect.Response[appv1.UpsertScheduleExceptionResponse], error) {
	if _, err := time.Parse("2006-01-02", req.Msg.Date); err != nil {
		return nil, invalidArgument("date must be a YYYY-MM-DD date")
	}
	if req.Msg.Kind == "" {
		return nil, invalidArgument("kind is required")
	}
	item, err := s.service.UpsertScheduleException(ctx, service.ScheduleExceptionInput{
		Date:            req.Msg.Date,
		Kind:            req.Msg.Kind,
		ExpectedMinutes: int(req.Msg.ExpectedMinutes),
		Windows:         toServiceWindows(req.Msg.Windows),
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(&appv1.UpsertScheduleExceptionResponse{Exception: toProtoScheduleException(item)}), nil
}

func (s *WorkScheduleService) DeleteScheduleException(ctx context.Context, req *connect.Request[appv1.DeleteScheduleExceptionRequest]) (*connect.Response[appv1.DeleteScheduleExceptionResponse], error) {
	if _, err := time.Parse("2006-01-02", req.Msg.Date); err != nil {
		return nil, invalidArgument("date must be a YYYY-MM-DD date")
	}
	if err := s.service.DeleteScheduleException(ctx, req.Msg.Date); err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(&appv1.DeleteScheduleExceptionResponse{}), nil
}

func (s *WorkScheduleService) ExpectedTimeForDate(ctx context.Context, req *connect.Request[appv1.ExpectedTimeForDateRequest]) (*connect.Response[appv1.ExpectedTimeForDateResponse], error) {
	if _, err := time.Parse("2006-01-02", req.Msg.Date); err != nil {
		return nil, invalidArgument("date must be a YYYY-MM-DD date")
	}
	item, err := s.service.ExpectedTimeForDate(ctx, req.Msg.Date)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(&appv1.ExpectedTimeForDateResponse{ExpectedTime: toProtoExpectedTime(item)}), nil
}

func (s *WorkScheduleService) ExpectedTimeForRange(ctx context.Context, req *connect.Request[appv1.ExpectedTimeForRangeRequest]) (*connect.Response[appv1.ExpectedTimeForRangeResponse], error) {
	if _, err := time.Parse("2006-01-02", req.Msg.StartDate); err != nil {
		return nil, invalidArgument("start_date must be a YYYY-MM-DD date")
	}
	if _, err := time.Parse("2006-01-02", req.Msg.EndDate); err != nil {
		return nil, invalidArgument("end_date must be a YYYY-MM-DD date")
	}
	items, err := s.service.ExpectedTimeForRange(ctx, req.Msg.StartDate, req.Msg.EndDate)
	if err != nil {
		return nil, mapServiceError(err)
	}
	out := make([]*appv1.ExpectedTime, len(items))
	for i := range items {
		out[i] = toProtoExpectedTime(items[i])
	}
	return connect.NewResponse(&appv1.ExpectedTimeForRangeResponse{Days: out}), nil
}

func toProtoWorkSchedule(s service.WorkSchedule) *appv1.WorkSchedule {
	days := make([]*appv1.WorkScheduleDay, len(s.Days))
	for i, d := range s.Days {
		days[i] = &appv1.WorkScheduleDay{
			Weekday:         d.Weekday,
			ExpectedMinutes: int32(d.ExpectedMinutes),
			Windows:         toProtoWindows(d.Windows),
		}
	}
	return &appv1.WorkSchedule{
		Id:            s.ID,
		Timezone:      s.Timezone,
		WorkweekStart: s.WorkweekStart,
		EffectiveFrom: s.EffectiveFrom,
		EffectiveTo:   s.EffectiveTo,
		Days:          days,
	}
}

func toProtoScheduleException(e service.ScheduleException) *appv1.ScheduleException {
	return &appv1.ScheduleException{
		Id:              e.ID,
		Date:            e.Date,
		Kind:            e.Kind,
		ExpectedMinutes: int32(e.ExpectedMinutes),
		Windows:         toProtoWindows(e.Windows),
	}
}

func toProtoExpectedTime(e service.ExpectedTime) *appv1.ExpectedTime {
	return &appv1.ExpectedTime{
		Date:            e.Date,
		ExpectedMinutes: int32(e.ExpectedMinutes),
		Windows:         toProtoWindows(e.Windows),
		Source:          e.Source,
		ExceptionKind:   e.ExceptionKind,
		Timezone:        e.Timezone,
		WorkweekStart:   e.WorkweekStart,
	}
}

func toProtoWindows(windows []service.WorkingWindow) []*appv1.WorkingWindow {
	out := make([]*appv1.WorkingWindow, len(windows))
	for i, w := range windows {
		out[i] = &appv1.WorkingWindow{StartMinutes: int32(w.StartMinutes), EndMinutes: int32(w.EndMinutes)}
	}
	return out
}

func toServiceWindows(windows []*appv1.WorkingWindow) []service.WorkingWindow {
	out := make([]service.WorkingWindow, 0, len(windows))
	for _, w := range windows {
		if w == nil {
			continue
		}
		out = append(out, service.WorkingWindow{StartMinutes: int(w.StartMinutes), EndMinutes: int(w.EndMinutes)})
	}
	return out
}
