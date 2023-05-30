package internal

import (
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

type RequestHandler struct {
	// Bring in the callback functions
	types.DefaultHttpContext
}

const (
	XRequestIdHeader = "x-request-id"
	AuthHeader       = "authorization"
)

// OnHttpRequestHeaders is called on every request we intercept with this WASM filter
// Check out the types.HttpContext interface to see what other callbacks you can override
//
// Note: Parameters are not needed here, but a brief description:
//   - numHeaders = fairly self-explanatory, the number of request headers
//   - endOfStream = only set to false when there is a request body (e.g. in a POST/PATCH/PUT request)
func (r *RequestHandler) OnHttpRequestHeaders(_ int, _ bool) types.Action {
	proxywasm.LogInfof("WASM plugin Handling request")

	// None of the parameters are useful here, so we have to ask the Envoy Sidecar for the actual request headers
	requestHeaders, err := proxywasm.GetHttpRequestHeaders()
	if err != nil {
		proxywasm.LogCriticalf("failed to get request headers: %v", err)
		// Allow Envoy Sidecar to forward this request to the upstream service
		return types.ActionContinue
	}

	// Making this a map makes accessing specific headers much easier later on
	reqHeaderMap := headerArrayToMap(requestHeaders)

	// Grab the always-present xRequestID to help grouping logs belonging to same request
	xRequestID := reqHeaderMap[XRequestIdHeader]

	// Now we can take action on this request
	return r.doSomethingWithRequest(reqHeaderMap, xRequestID)
}

// headerArrayToMap is a simple function to convert from array of headers to a Map
func headerArrayToMap(requestHeaders [][2]string) map[string]string {
	headerMap := make(map[string]string)
	for _, header := range requestHeaders {
		headerMap[header[0]] = header[1]
	}
	return headerMap
}

func (r *RequestHandler) doSomethingWithRequest(reqHeaderMap map[string]string, xRequestID string) types.Action {
	// for now, let's just log all the request headers to we get an idea of what we have to work with
	for key, value := range reqHeaderMap {
		proxywasm.LogInfof("  %s: request header --> %s: %s", xRequestID, key, value)
	}

	// if auth header exists, call out to auth-service to request JWT
	if _, exists := reqHeaderMap[AuthHeader]; exists {
		authClient := AuthClient{XRequestID: xRequestID}
		authClient.RequestJWT(reqHeaderMap)
		// We need to tell Envoy to block this request until we get a response from the Auth Service
		return types.ActionPause
	}

	// If there was no authentication header to operate on, then
	// forward request to upstream service, i.e. unblock request
	return types.ActionContinue
}
