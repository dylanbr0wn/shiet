package appapi_test

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"

	"connectrpc.com/connect"
	appv1 "github.com/dylanbr0wn/shiet/gen/shiet/app/v1"
	"github.com/dylanbr0wn/shiet/gen/shiet/app/v1/appv1connect"
	"github.com/dylanbr0wn/shiet/internal/api/appapi"
	"github.com/dylanbr0wn/shiet/internal/db"
	"github.com/dylanbr0wn/shiet/internal/seed"
	"github.com/dylanbr0wn/shiet/internal/service"
)

func TestWorkScheduleService_ExpectedTimeAndException(t *testing.T) {
	t.Parallel()

	conn, err := db.Open(filepath.Join(t.TempDir(), "shiet.db"))
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

	client := appv1connect.NewWorkScheduleServiceClient(&http.Client{
		Transport: handlerTransport{handler: appapi.NewHandler(appapi.Dependencies{Service: service.New(conn)})},
	}, "http://shiet.test")

	weekday, err := client.ExpectedTimeForDate(context.Background(), connect.NewRequest(&appv1.ExpectedTimeForDateRequest{
		Date: "2026-06-15",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if weekday.Msg.ExpectedTime == nil || weekday.Msg.ExpectedTime.ExpectedMinutes != 480 {
		t.Fatalf("weekday expected = %#v", weekday.Msg.ExpectedTime)
	}
	if weekday.Msg.ExpectedTime.Source != "weekday" {
		t.Fatalf("source = %q", weekday.Msg.ExpectedTime.Source)
	}

	_, err = client.UpsertScheduleException(context.Background(), connect.NewRequest(&appv1.UpsertScheduleExceptionRequest{
		Date:            "2026-06-15",
		Kind:            "holiday",
		ExpectedMinutes: 0,
	}))
	if err != nil {
		t.Fatal(err)
	}

	holiday, err := client.ExpectedTimeForDate(context.Background(), connect.NewRequest(&appv1.ExpectedTimeForDateRequest{
		Date: "2026-06-15",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if holiday.Msg.ExpectedTime.Source != "exception" || holiday.Msg.ExpectedTime.ExceptionKind != "holiday" {
		t.Fatalf("holiday expected = %#v", holiday.Msg.ExpectedTime)
	}

	listed, err := client.ListWorkSchedules(context.Background(), connect.NewRequest(&appv1.ListWorkSchedulesRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Msg.Schedules) != 1 || listed.Msg.Schedules[0].Timezone != "America/Toronto" {
		t.Fatalf("schedules = %#v", listed.Msg.Schedules)
	}
}
