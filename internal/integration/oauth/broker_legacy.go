package oauth

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	brokerv1 "github.com/dylanbr0wn/shiet/gen/shiet/broker/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type LegacyBrokerError struct {
	Status int
	Code   string
	Op     string
}

func (e *LegacyBrokerError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("legacy broker %s error %s", e.Op, e.Code)
	}
	return fmt.Sprintf("legacy broker %s returned %d", e.Op, e.Status)
}

func LegacyStartAuthorization(ctx context.Context, client *http.Client, base, provider string, req *brokerv1.StartAuthorizationRequest) (*brokerv1.StartAuthorizationResponse, error) {
	payload := &brokerv1.LegacyStartAuthorizationRequest{
		DesktopSessionId:       req.DesktopSessionId,
		HandoffChallenge:       req.HandoffChallenge,
		DesktopHandoffRedirect: req.DesktopHandoffRedirect,
	}
	if req.Application != nil {
		payload.AppVersion = req.Application.AppVersion
		payload.Platform = req.Application.Platform
	}
	response := &brokerv1.StartAuthorizationResponse{}
	if err := postLegacyJSON(ctx, client, base+"/v1/"+provider+"/oauth/start", payload, response, "start"); err != nil {
		return nil, err
	}
	return response, nil
}

func LegacyExchangeHandoff(ctx context.Context, client *http.Client, base, provider string, req *brokerv1.ExchangeHandoffRequest) (*brokerv1.ExchangeHandoffResponse, error) {
	payload := &brokerv1.ExchangeHandoffRequest{
		DesktopSessionId: req.DesktopSessionId,
		BrokerState:      req.BrokerState,
		HandoffCode:      req.HandoffCode,
		HandoffVerifier:  req.HandoffVerifier,
	}
	response := &brokerv1.LegacyHandoffResponse{}
	if err := postLegacyJSON(ctx, client, base+"/v1/"+provider+"/oauth/handoff", payload, response, "handoff"); err != nil {
		return nil, err
	}
	providerValue := brokerv1.Provider_PROVIDER_GOOGLE
	if response.Provider == "github" {
		providerValue = brokerv1.Provider_PROVIDER_GITHUB
	}
	if response.Token == nil {
		return nil, fmt.Errorf("decode legacy broker handoff response: token is required")
	}
	return &brokerv1.ExchangeHandoffResponse{
		Provider:    providerValue,
		AccountHint: response.AccountHint,
		Scopes:      response.Scope,
		Token: &brokerv1.TokenMaterial{
			AccessToken:  response.Token.AccessToken,
			RefreshToken: response.Token.GetRefreshToken(),
			TokenType:    response.Token.TokenType,
			Expiry:       response.Token.Expiry,
		},
	}, nil
}

func LegacyRefreshToken(ctx context.Context, client *http.Client, base string, req *brokerv1.RefreshTokenRequest) (*brokerv1.RefreshTokenResponse, error) {
	payload := &brokerv1.LegacyRefreshTokenRequest{RefreshToken: req.RefreshToken, Scope: req.Scopes}
	if req.Application != nil {
		payload.AppVersion = req.Application.AppVersion
		payload.Platform = req.Application.Platform
	}
	response := &brokerv1.LegacyRefreshTokenResponse{}
	if err := postLegacyJSON(ctx, client, base+"/v1/google/oauth/refresh", payload, response, "refresh"); err != nil {
		return nil, err
	}
	return &brokerv1.RefreshTokenResponse{Token: &brokerv1.TokenMaterial{
		AccessToken:  response.AccessToken,
		RefreshToken: response.GetRefreshToken(),
		TokenType:    response.TokenType,
		Expiry:       response.Expiry,
	}}, nil
}

func LegacyRevokeToken(ctx context.Context, client *http.Client, base, provider string, req *brokerv1.RevokeTokenRequest) (*brokerv1.RevokeTokenResponse, error) {
	payload := &brokerv1.LegacyRevokeTokenRequest{Reason: req.Reason}
	if provider == "github" {
		payload.AccessToken = req.GetAccessToken()
	} else {
		payload.RefreshToken = req.GetRefreshToken()
	}
	response := &brokerv1.RevokeTokenResponse{}
	if err := postLegacyJSON(ctx, client, base+"/v1/"+provider+"/oauth/revoke", payload, response, "revoke"); err != nil {
		return nil, err
	}
	return response, nil
}

func postLegacyJSON(ctx context.Context, client *http.Client, endpoint string, payload, response proto.Message, op string) error {
	body, err := (protojson.MarshalOptions{UseProtoNames: true}).Marshal(payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	if client == nil {
		client = http.DefaultClient
	}
	result, err := client.Do(request)
	if err != nil {
		return &LegacyBrokerError{Status: 0, Op: op}
	}
	defer result.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(result.Body, 1<<20))
	if result.StatusCode < 200 || result.StatusCode >= 300 {
		brokerError := &brokerv1.LegacyErrorResponse{}
		_ = (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(raw, brokerError)
		return &LegacyBrokerError{Status: result.StatusCode, Code: strings.TrimSpace(brokerError.Error), Op: op}
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: false}).Unmarshal(raw, response); err != nil {
		return fmt.Errorf("decode legacy broker %s response: %w", op, err)
	}
	return nil
}
