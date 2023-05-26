package internal

import (
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
)

const (
	XAuthServiceAuthResponseHeader = "x-auth-response-status"
)

type AuthClient struct {
	Conf       Config
	Metrics    *Metrics
	XRequestID string
}

func (d *AuthClient) RequestJWT(reqHeaders [][2]string) {
	proxywasm.LogInfof("%s: Requesting JWT: %s", d.XRequestID, d.Conf.AuthClusterName)

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

	proxywasm.LogInfof("%s: Got response from AuthService", d.XRequestID)

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
		if h[0] == XAuthServiceAuthResponseHeader {
			if err := proxywasm.AddHttpRequestHeader(h[0], h[1]); err != nil {
				proxywasm.LogCriticalf("%s: failed to add header '%v' to request: %v", d.XRequestID, h, err)
			}
		}
	}
	d.Metrics.Increment("auth_called", [][2]string{})
}
