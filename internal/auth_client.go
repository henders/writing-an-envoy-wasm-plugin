package internal

import (
	"fmt"
	"time"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
)

const (
	XAuthServiceHeader             = "x-auth-jwt"
	XAuthServiceAuthResponseHeader = "x-auth-response-status"
)

type AuthClient struct {
	Conf       Config
	Metrics    *Metrics
	StartTime  time.Time
	XRequestID string
}

func (d *AuthClient) RequestJWT(reqHeaders [][2]string) {
	proxywasm.LogInfof("%s: Requesting JWT: %s", d.XRequestID, d.Conf.AuthClusterName)
	d.StartTime = time.Now()

	_, err := proxywasm.DispatchHttpCall(d.Conf.AuthClusterName, reqHeaders, nil, nil, d.Conf.AuthTimeout, d.authResponse)
	if err != nil {
		proxywasm.LogCriticalf("%s: failed to call AuthService: %v", d.XRequestID, err)
		_ = proxywasm.ResumeHttpRequest()
		d.Metrics.Increment("auth_called", [][2]string{{"result", "failed"}})
	}
}

func (d *AuthClient) authResponse(_, bodySize, _ int) {
	// We want to always resume the intercepted request regardless of success/fail to not indefinitely block anything
	defer func() {
		if err := proxywasm.ResumeHttpRequest(); err != nil {
			proxywasm.LogCriticalf("%s: failed to ResumeHttpRequest after calling auth: %v", d.XRequestID, err)
		}
	}()

	// Define all the headers we'll copy from AuthService to the original request headers
	headersToCopyToOrigReq := map[string]bool{
		XAuthServiceAuthResponseHeader: true,
		XAuthServiceHeader:             true,
	}

	proxywasm.LogInfof("%s: Successfully got response from AuthService", d.XRequestID)
	if bodySize > 0 {
		// Any 'body' returned by AuthService indicates an error
		proxywasm.LogCriticalf("  %s: auth response body --> %s", d.XRequestID, d.responseBody())
	}

	// Get the response headers from our call to AuthService
	headers, err := proxywasm.GetHttpCallResponseHeaders()
	if err != nil {
		proxywasm.LogCriticalf("%s: failed to GetHttpCallResponseHeaders from auth response: %v", d.XRequestID, err)
		return
	}

	// Now add specific headers from the response onto our originally intercepted request
	for _, h := range headers {
		proxywasm.LogInfof("  %s: auth response header --> %s: %s", d.XRequestID, h[0], h[1])
		// Copy auth header onto original request headers
		if headersToCopyToOrigReq[h[0]] {
			if err := proxywasm.AddHttpRequestHeader(h[0], h[1]); err != nil {
				proxywasm.LogCriticalf("%s: failed to add header '%v' to request: %v", d.XRequestID, h, err)
			}
		}
	}
	d.Metrics.Increment("auth_called", [][2]string{})
	d.Metrics.Histogram("auth_called_latency", [][2]string{}, uint64(time.Since(d.StartTime).Milliseconds()))
}

func (d *AuthClient) responseBody() string {
	body, err := proxywasm.GetHttpCallResponseBody(0, 1024)
	if err != nil {
		return fmt.Sprintf("%s: failed to GetHttpCallResponseBody for auth response: %v", d.XRequestID, err)
	}
	return string(body)
}
