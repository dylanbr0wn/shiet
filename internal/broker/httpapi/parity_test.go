package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	brokerv1 "github.com/dylanbr0wn/shiet/gen/shiet/broker/v1"
	"github.com/dylanbr0wn/shiet/gen/shiet/broker/v1/brokerv1connect"
	"github.com/dylanbr0wn/shiet/internal/broker/codes"
	"github.com/dylanbr0wn/shiet/internal/broker/observe"
	"github.com/dylanbr0wn/shiet/internal/broker/store"
)

func TestBrokerRESTAndConnectSuccessParity(t *testing.T) {
	t.Parallel()

	t.Run("start", func(t *testing.T) {
		now := time.Date(2026, 7, 9, 12, 0, 0, 0, time.UTC)
		restMem, connectMem := &memoryStore{}, &memoryStore{}
		restMetrics, connectMetrics := observe.NewMetrics(), observe.NewMetrics()
		restServer := Server{Config: testConfig(), Store: restMem, Clock: func() time.Time { return now }, Metrics: restMetrics}
		connectServer := Server{Config: testConfig(), Store: connectMem, Clock: func() time.Time { return now }, Metrics: connectMetrics}

		rest := httptest.NewRecorder()
		restServer.Handler().ServeHTTP(rest, httptest.NewRequest(http.MethodPost, "/v1/google/oauth/start", strings.NewReader(`{
			"desktop_session_id":"desktop-1","handoff_challenge":"challenge-1","app_version":"1.2.3","platform":"test"
		}`)))
		if rest.Code != http.StatusCreated {
			t.Fatalf("REST status = %d: %s", rest.Code, rest.Body.String())
		}
		var restResponse startResponse
		if err := json.NewDecoder(rest.Body).Decode(&restResponse); err != nil {
			t.Fatal(err)
		}

		connectResponse, err := connectTestClient(connectServer).StartAuthorization(context.Background(), connect.NewRequest(&brokerv1.StartAuthorizationRequest{
			Provider:         brokerv1.Provider_PROVIDER_GOOGLE,
			DesktopSessionId: "desktop-1",
			HandoffChallenge: "challenge-1",
			Application:      &brokerv1.ApplicationMetadata{AppVersion: "1.2.3", Platform: "test"},
		}))
		if err != nil {
			t.Fatal(err)
		}
		if restResponse.AuthURL == "" || connectResponse.Msg.AuthUrl == "" || len(restMem.states) != 1 || len(connectMem.states) != 1 {
			t.Fatalf("start parity: REST=%+v Connect=%+v states=%d/%d", restResponse, connectResponse.Msg, len(restMem.states), len(connectMem.states))
		}
		if restMem.states[0].AppVersion != connectMem.states[0].AppVersion || metricsText(restMetrics) != metricsText(connectMetrics) {
			t.Fatalf("start side effects diverged: REST=%+v Connect=%+v\n%s\n%s", restMem.states[0], connectMem.states[0], metricsText(restMetrics), metricsText(connectMetrics))
		}
	})

	t.Run("handoff and replay", func(t *testing.T) {
		now := time.Date(2026, 7, 9, 12, 0, 0, 0, time.UTC)
		restMetrics, connectMetrics := observe.NewMetrics(), observe.NewMetrics()
		restServer, restMem := handoffParityServer(t, now, restMetrics)
		connectServer, connectMem := handoffParityServer(t, now, connectMetrics)
		body := `{"desktop_session_id":"desktop-1","broker_state":"state-1","handoff_code":"code-1","handoff_verifier":"verifier-1"}`

		rest := httptest.NewRecorder()
		restServer.Handler().ServeHTTP(rest, httptest.NewRequest(http.MethodPost, "/v1/google/oauth/handoff", strings.NewReader(body)))
		var restResponse handoffResponse
		if rest.Code != http.StatusOK || json.NewDecoder(rest.Body).Decode(&restResponse) != nil {
			t.Fatalf("REST handoff = %d %s", rest.Code, rest.Body.String())
		}
		request := &brokerv1.ExchangeHandoffRequest{
			Provider: brokerv1.Provider_PROVIDER_GOOGLE, DesktopSessionId: "desktop-1", BrokerState: "state-1", HandoffCode: "code-1", HandoffVerifier: "verifier-1",
			Application: &brokerv1.ApplicationMetadata{AppVersion: "1.2.3", Platform: "test"},
		}
		connectResponse, err := connectTestClient(connectServer).ExchangeHandoff(context.Background(), connect.NewRequest(request))
		if err != nil {
			t.Fatal(err)
		}
		if restResponse.Token.AccessToken != connectResponse.Msg.Token.AccessToken || restMem.handoffs[0].UsedAt == nil || connectMem.handoffs[0].UsedAt == nil {
			t.Fatalf("handoff response/side effect diverged")
		}

		replay := httptest.NewRecorder()
		restServer.Handler().ServeHTTP(replay, httptest.NewRequest(http.MethodPost, "/v1/google/oauth/handoff", strings.NewReader(body)))
		if replay.Code != http.StatusBadRequest || !strings.Contains(replay.Body.String(), codes.HandoffAlreadyUsed) {
			t.Fatalf("REST replay = %d %s", replay.Code, replay.Body.String())
		}
		_, err = connectTestClient(connectServer).ExchangeHandoff(context.Background(), connect.NewRequest(request))
		assertConnectBrokerError(t, err, connect.CodeAlreadyExists, codes.HandoffAlreadyUsed)
		if metricsText(restMetrics) != metricsText(connectMetrics) {
			t.Fatalf("handoff metrics diverged:\n%s\n%s", metricsText(restMetrics), metricsText(connectMetrics))
		}
	})

	t.Run("refresh", func(t *testing.T) {
		restMetrics, connectMetrics := observe.NewMetrics(), observe.NewMetrics()
		restServer := refreshParityServer(restMetrics)
		connectServer := refreshParityServer(connectMetrics)
		rest := httptest.NewRecorder()
		restServer.Handler().ServeHTTP(rest, httptest.NewRequest(http.MethodPost, "/v1/google/oauth/refresh", strings.NewReader(`{"refresh_token":"refresh-1","app_version":"1.2.3"}`)))
		var restResponse refreshResponse
		if rest.Code != http.StatusOK || json.NewDecoder(rest.Body).Decode(&restResponse) != nil {
			t.Fatalf("REST refresh = %d %s", rest.Code, rest.Body.String())
		}
		connectResponse, err := connectTestClient(connectServer).RefreshToken(context.Background(), connect.NewRequest(&brokerv1.RefreshTokenRequest{
			Provider: brokerv1.Provider_PROVIDER_GOOGLE, RefreshToken: "refresh-1", Application: &brokerv1.ApplicationMetadata{AppVersion: "1.2.3"},
		}))
		if err != nil || restResponse.AccessToken != connectResponse.Msg.Token.AccessToken || metricsText(restMetrics) != metricsText(connectMetrics) {
			t.Fatalf("refresh parity: REST=%+v Connect=%+v err=%v", restResponse, connectResponse, err)
		}
	})

	t.Run("revoke", func(t *testing.T) {
		restMetrics, connectMetrics := observe.NewMetrics(), observe.NewMetrics()
		restServer := revokeParityServer(restMetrics)
		connectServer := revokeParityServer(connectMetrics)
		rest := httptest.NewRecorder()
		restServer.Handler().ServeHTTP(rest, httptest.NewRequest(http.MethodPost, "/v1/google/oauth/revoke", strings.NewReader(`{"refresh_token":"refresh-1"}`)))
		var restResponse revokeResponse
		if rest.Code != http.StatusOK || json.NewDecoder(rest.Body).Decode(&restResponse) != nil {
			t.Fatalf("REST revoke = %d %s", rest.Code, rest.Body.String())
		}
		connectResponse, err := connectTestClient(connectServer).RevokeToken(context.Background(), connect.NewRequest(&brokerv1.RevokeTokenRequest{
			Provider: brokerv1.Provider_PROVIDER_GOOGLE, Credential: &brokerv1.RevokeTokenRequest_RefreshToken{RefreshToken: "refresh-1"},
		}))
		if err != nil || !restResponse.Revoked || !connectResponse.Msg.Revoked || metricsText(restMetrics) != metricsText(connectMetrics) {
			t.Fatalf("revoke parity: REST=%+v Connect=%+v err=%v", restResponse, connectResponse, err)
		}
	})
}

func TestBrokerRESTAndConnectRateLimitParity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, restPath, restBody, surface string
		connectCall                       func(context.Context, brokerv1connect.OAuthBrokerServiceClient) error
	}{
		{"start", "/v1/google/oauth/start", `{}`, codes.SurfaceStart, func(ctx context.Context, c brokerv1connect.OAuthBrokerServiceClient) error {
			_, err := c.StartAuthorization(ctx, connect.NewRequest(&brokerv1.StartAuthorizationRequest{Provider: brokerv1.Provider_PROVIDER_GOOGLE}))
			return err
		}},
		{"handoff", "/v1/google/oauth/handoff", `{}`, codes.SurfaceHandoff, func(ctx context.Context, c brokerv1connect.OAuthBrokerServiceClient) error {
			_, err := c.ExchangeHandoff(ctx, connect.NewRequest(&brokerv1.ExchangeHandoffRequest{Provider: brokerv1.Provider_PROVIDER_GOOGLE}))
			return err
		}},
		{"refresh", "/v1/google/oauth/refresh", `{}`, codes.SurfaceRefresh, func(ctx context.Context, c brokerv1connect.OAuthBrokerServiceClient) error {
			_, err := c.RefreshToken(ctx, connect.NewRequest(&brokerv1.RefreshTokenRequest{Provider: brokerv1.Provider_PROVIDER_GOOGLE}))
			return err
		}},
		{"revoke", "/v1/google/oauth/revoke", `{}`, codes.SurfaceRevoke, func(ctx context.Context, c brokerv1connect.OAuthBrokerServiceClient) error {
			_, err := c.RevokeToken(ctx, connect.NewRequest(&brokerv1.RevokeTokenRequest{Provider: brokerv1.Provider_PROVIDER_GOOGLE}))
			return err
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			restMetrics, connectMetrics := observe.NewMetrics(), observe.NewMetrics()
			restServer := Server{Config: testConfig(), Store: &memoryStore{}, Metrics: restMetrics, Limiter: denyLimiter{}}
			connectServer := Server{Config: testConfig(), Store: &memoryStore{}, Metrics: connectMetrics, Limiter: denyLimiter{}}
			rest := httptest.NewRecorder()
			restServer.Handler().ServeHTTP(rest, httptest.NewRequest(http.MethodPost, tc.restPath, strings.NewReader(tc.restBody)))
			if rest.Code != http.StatusTooManyRequests || !strings.Contains(rest.Body.String(), codes.RateLimited) {
				t.Fatalf("REST limit = %d %s", rest.Code, rest.Body.String())
			}
			err := tc.connectCall(context.Background(), connectTestClient(connectServer))
			assertConnectBrokerError(t, err, connect.CodeResourceExhausted, codes.RateLimited)
			if restMetrics.RateLimitedCount(tc.surface) != 1 || connectMetrics.RateLimitedCount(tc.surface) != 1 || metricsText(restMetrics) != metricsText(connectMetrics) {
				t.Fatalf("rate-limit metrics diverged:\n%s\n%s", metricsText(restMetrics), metricsText(connectMetrics))
			}
		})
	}
}

func handoffParityServer(t *testing.T, now time.Time, metrics *observe.Metrics) (Server, *memoryStore) {
	t.Helper()
	payload, err := encryptTokenPayload(testConfig().GoogleClientSecret, handoffAAD("state-1", "desktop-1", pkceS256("verifier-1")), tokenPayload{AccessToken: "access-1", RefreshToken: "refresh-1", TokenType: "Bearer", Expiry: now.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	mem := &memoryStore{}
	_ = mem.SaveHandoff(context.Background(), store.HandoffRecord{CodeHash: hashHandoffCode("code-1"), Provider: "google", StateID: "state-1", DesktopSessionID: "desktop-1", HandoffChallenge: pkceS256("verifier-1"), EncryptedTokenPayload: payload, ExpiresAt: now.Add(time.Minute)})
	return Server{Config: testConfig(), Store: mem, Clock: func() time.Time { return now }, Metrics: metrics}, mem
}

func refreshParityServer(metrics *observe.Metrics) Server {
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"access_token":"access-2","token_type":"Bearer","expires_in":3600}`))}, nil
	})}
	return Server{Config: testConfig(), Store: &memoryStore{}, HTTPClient: client, Metrics: metrics}
}

func revokeParityServer(metrics *observe.Metrics) Server {
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{}`))}, nil
	})}
	return Server{Config: testConfig(), Store: &memoryStore{}, HTTPClient: client, GoogleRevokeURL: "https://provider.test/revoke", Metrics: metrics}
}

func metricsText(metrics *observe.Metrics) string {
	var out strings.Builder
	metrics.WritePrometheus(&out)
	return out.String()
}

type denyLimiter struct{}

func (denyLimiter) Allow(string, int) bool { return false }
