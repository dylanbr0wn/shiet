// Package appapi exposes portable application operations through Connect.
package appapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"connectrpc.com/connect"
	appv1 "github.com/dylanbr0wn/shiet/gen/shiet/app/v1"
	"github.com/dylanbr0wn/shiet/gen/shiet/app/v1/appv1connect"
	"github.com/dylanbr0wn/shiet/internal/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PeriodService struct {
	service *service.Service
}

func NewPeriodService(svc *service.Service) *PeriodService {
	return &PeriodService{service: svc}
}

func NewHandler(svc *service.Service) http.Handler {
	mux := http.NewServeMux()
	path, handler := appv1connect.NewPeriodServiceHandler(NewPeriodService(svc))
	mux.Handle(path, handler)
	return mux
}

func (s *PeriodService) ListPeriods(ctx context.Context, _ *connect.Request[appv1.ListPeriodsRequest]) (*connect.Response[appv1.ListPeriodsResponse], error) {
	periods, err := s.service.ListPeriods(ctx)
	if err != nil {
		return nil, mapServiceError(err)
	}
	out := &appv1.ListPeriodsResponse{Periods: make([]*appv1.Period, len(periods))}
	for i := range periods {
		out.Periods[i] = toProtoPeriod(periods[i])
	}
	return connect.NewResponse(out), nil
}

func (s *PeriodService) EnsureCurrentPeriod(ctx context.Context, req *connect.Request[appv1.EnsureCurrentPeriodRequest]) (*connect.Response[appv1.EnsureCurrentPeriodResponse], error) {
	if _, err := time.Parse("2006-01-02", req.Msg.Today); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("today must be a YYYY-MM-DD date"))
	}
	ianaTZ := req.Msg.IanaTz
	if ianaTZ == "" {
		ianaTZ = "UTC"
	}
	if _, err := time.LoadLocation(ianaTZ); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("iana_tz must be a valid IANA timezone"))
	}
	period, err := s.service.EnsureCurrentPeriod(ctx, req.Msg.Today, ianaTZ)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(&appv1.EnsureCurrentPeriodResponse{Period: toProtoPeriod(period)}), nil
}

func toProtoPeriod(period service.Period) *appv1.Period {
	out := &appv1.Period{
		Id:                period.ID,
		StartDate:         period.StartDate,
		EndDate:           period.EndDate,
		Cadence:           period.Cadence,
		AnchorDate:        period.AnchorDate,
		TargetHoursPerDay: period.TargetHoursPerDay,
	}
	if period.LastSyncedAt != nil {
		out.LastSyncedAt = timestamppb.New(*period.LastSyncedAt)
	}
	return out
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, context.Canceled):
		return connect.NewError(connect.CodeCanceled, errors.New("request canceled"))
	case errors.Is(err, context.DeadlineExceeded):
		return connect.NewError(connect.CodeDeadlineExceeded, errors.New("request deadline exceeded"))
	case errors.Is(err, service.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, errors.New("resource not found"))
	default:
		return connect.NewError(connect.CodeInternal, errors.New("internal service error"))
	}
}
