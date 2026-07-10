package oauth_test

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	brokerv1 "github.com/dylanbr0wn/shiet/gen/shiet/broker/v1"
	"github.com/dylanbr0wn/shiet/internal/integration/oauth"
	"google.golang.org/protobuf/proto"
)

func TestLegacyFallbackRequiresDefinitiveUnsupportedProcedure(t *testing.T) {
	t.Parallel()

	unsupported := connect.NewError(connect.CodeUnimplemented, errors.New("procedure not found"))
	if !oauth.ShouldFallbackToLegacy(unsupported) {
		t.Fatal("unsupported procedure should fall back to the released REST API")
	}

	for _, code := range []connect.Code{connect.CodeUnavailable, connect.CodeInternal, connect.CodeUnknown} {
		if oauth.ShouldFallbackToLegacy(connect.NewError(code, errors.New("ambiguous failure"))) {
			t.Fatalf("ambiguous %s failure must not replay through REST", code)
		}
	}

	withDetail := connect.NewError(connect.CodeUnimplemented, errors.New("broker rejected operation"))
	detail, err := connect.NewErrorDetail(proto.Message(&brokerv1.BrokerErrorDetail{Code: "operation_not_supported"}))
	if err != nil {
		t.Fatal(err)
	}
	withDetail.AddDetail(detail)
	if oauth.ShouldFallbackToLegacy(withDetail) {
		t.Fatal("a broker error detail proves Connect is supported; do not fall back")
	}
}
