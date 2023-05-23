package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func TestRequestContext_OnHttpRequestHeaders(t *testing.T) {
	tests := []struct {
		name         string
		headers      [][2]string
		conf         string
		wantResponse types.Action
		wantHeaders  [][2]string
		wantAuthCall bool
	}{
		{
			name:         "Empty headers",
			headers:      [][2]string{},
			conf:         DefaultTestConfig,
			wantAuthCall: false,
			wantHeaders:  [][2]string{},
			wantResponse: types.ActionContinue,
		},
		{
			name: "No auth or mTLS headers",
			headers: [][2]string{
				{XRequestIdHeader, "abc"},
			},
			conf:         DefaultTestConfig,
			wantAuthCall: false,
			wantHeaders:  [][2]string{{"x-request-id", "abc"}},
			wantResponse: types.ActionContinue,
		},
		{
			name: "authorization header present with no mTLS headers",
			headers: [][2]string{
				{XRequestIdHeader, "abc"},
				{AuthHeader, "MAC <some mac digest>"},
			},
			conf:         DefaultTestConfig,
			wantAuthCall: true,
			wantHeaders: [][2]string{
				{"x-request-id", "abc"},
				{AuthHeader, "MAC <some mac digest>"},
			},
			wantResponse: types.ActionPause,
		},
	}

	vmContext := InitPlugin(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize new plugin
			host, contextID, reset := NewContextWithConfig(t, vmContext, tt.conf)
			defer reset()

			// Call OnRequestHeaders
			action := host.CallOnRequestHeaders(contextID, tt.headers, true)
			require.Equal(t, tt.wantResponse, action)
			require.Equal(t, tt.wantHeaders, host.GetCurrentRequestHeaders(contextID))
			if tt.wantAuthCall {
				// Verify Auth is called.
				require.Equal(t, 1, len(host.GetCalloutAttributesFromContext(contextID)))
			} else {
				require.Equal(t, 0, len(host.GetCalloutAttributesFromContext(contextID)))
			}
		})
	}
}
