package internal

import (
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

type RequestHandler struct {
	// Bring in the callback functions
	types.DefaultHttpContext

	Conf      Config
	ContextID uint32
	Metrics   *Metrics
}

const (
	XRequestIdHeader = "x-request-id"
	AuthHeader       = "authorization"
)

// OnHttpRequestHeaders is called on every request we intercept with this WASM filter
func (r *RequestHandler) OnHttpRequestHeaders(_ int, _ bool) types.Action {
	proxywasm.LogDebugf("Handling request - context:%d", r.ContextID)
	r.Metrics.Increment("requests", [][2]string{})

	requestHeaders, err := proxywasm.GetHttpRequestHeaders()
	if err != nil {
		proxywasm.LogCriticalf("%d: failed to get request headers: %v", r.ContextID, err)
		return types.ActionContinue
	}

	// Making this a map makes accessing specific headers much easier later on
	reqHeaderMap := make(map[string]string)
	for _, header := range requestHeaders {
		reqHeaderMap[header[0]] = header[1]
	}

	xRequestID := reqHeaderMap[XRequestIdHeader] // Use the always-present xRequestID to help print contextual logs
	for _, h := range requestHeaders {
		proxywasm.LogInfof("  %s: request header --> %s: %s", xRequestID, h[0], h[1])
	}

	// if auth header exists, call out to auth-service to request JWT
	if _, exists := reqHeaderMap[AuthHeader]; exists {
		authClient := AuthClient{XRequestID: xRequestID, Conf: r.Conf, Metrics: r.Metrics}
		authClient.RequestJWT(requestHeaders)
		return types.ActionPause
	}

	return types.ActionContinue
}
