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
// Check out the types.HttpContext interface to see what other callbacks you can override
func (r *RequestHandler) OnHttpRequestHeaders(_ int, _ bool) types.Action {
	proxywasm.LogDebugf("Handling request - context:%d", r.ContextID)
	r.Metrics.Increment("requests", [][2]string{})

	// None of the parameters are useful here, so we have to ask the Envoy Sidecar for the actual request headers
	requestHeaders, err := proxywasm.GetHttpRequestHeaders()
	if err != nil {
		proxywasm.LogCriticalf("%d: failed to get request headers: %v", r.ContextID, err)
		// Allow Envoy Sidecar to forward this request to the upstream service
		return types.ActionContinue
	}

	// Making this a map makes accessing specific headers much easier later on
	reqHeaderMap := headerArrayToMap(requestHeaders)

	// Let's dump all the request headers to help debugging
	xRequestID := reqHeaderMap[XRequestIdHeader] // Use the always-present xRequestID to help print contextual logs
	for _, h := range requestHeaders {
		proxywasm.LogInfof("  %s: request header --> %s: %s", xRequestID, h[0], h[1])
	}

	// if auth header exists, call out to auth-service to request JWT
	if _, exists := reqHeaderMap[AuthHeader]; exists {
		authClient := AuthClient{XRequestID: xRequestID, Conf: r.Conf, Metrics: r.Metrics}
		authClient.RequestJWT(requestHeaders)
		// Tell the Envoy Sidecar to not forward this request to upstream service yet
		return types.ActionPause
	}

	// If there was no authentication header to operate on, then just forward request to upstream service to let
	// it make the decision on what to do.
	return types.ActionContinue
}

// headerArrayToMap is a simple function to convert from array of headers to a Map
func headerArrayToMap(requestHeaders [][2]string) map[string]string {
	headerMap := make(map[string]string)
	for _, header := range requestHeaders {
		headerMap[header[0]] = header[1]
	}
	return headerMap
}
