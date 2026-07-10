package appapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
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

func TestPeriodServiceEnsuresAndListsCurrentPeriod(t *testing.T) {
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

	client := appv1connect.NewPeriodServiceClient(&http.Client{
		Transport: handlerTransport{handler: appapi.NewHandler(service.New(conn))},
	}, "http://shiet.test")

	ensured, err := client.EnsureCurrentPeriod(context.Background(), connect.NewRequest(&appv1.EnsureCurrentPeriodRequest{
		Today:  "2026-07-09",
		IanaTz: "America/Vancouver",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if ensured.Msg.Period == nil || ensured.Msg.Period.Id <= 0 {
		t.Fatalf("expected persisted period, got %#v", ensured.Msg.Period)
	}

	listed, err := client.ListPeriods(context.Background(), connect.NewRequest(&appv1.ListPeriodsRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if got := len(listed.Msg.Periods); got != 1 {
		t.Fatalf("expected one period, got %d", got)
	}
	if got, want := listed.Msg.Periods[0].Id, ensured.Msg.Period.Id; got != want {
		t.Fatalf("period id = %d, want %d", got, want)
	}
}

type handlerTransport struct {
	handler http.Handler
}

func (t handlerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	recorder := httptest.NewRecorder()
	t.handler.ServeHTTP(recorder, req)
	return recorder.Result(), nil
}
