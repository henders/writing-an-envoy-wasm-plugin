package internal

import (
	"os"
	"time"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tidwall/gjson"
)

const (
	AuthTimeoutDefault = time.Second
)

type Config struct {
	AuthClusterName string
	AuthAuthority   string
	AuthTimeout     uint32
	Namespace       string
}

func NewConfig() Config {
	configuration := getPluginConfiguration()
	namespace := getNamespace()
	config := Config{
		AuthClusterName: getAuthClusterName(configuration),
		AuthAuthority:   getAuthAuthority(configuration),
		AuthTimeout:     uint32(getAuthTimeout(configuration)),
		Namespace:       namespace,
	}

	return config
}

func getPluginConfiguration() gjson.Result {
	proxywasm.LogInfof("Getting WASM plugin config...")
	configuration, err := proxywasm.GetPluginConfiguration()
	if err != nil {
		proxywasm.LogCriticalf("error reading plugin configuration: %v", err)
	}
	if len(configuration) == 0 {
		proxywasm.LogCritical("WASM plugin config was empty")
		return gjson.Result{}
	}
	if !gjson.ValidBytes(configuration) {
		proxywasm.LogCriticalf("WASM plugin config was invalid: %s", configuration)
		return gjson.Result{}
	}

	result := gjson.ParseBytes(configuration)
	return result
}

func getAuthClusterName(configuration gjson.Result) string {
	authClusterNameResult := configuration.Get("auth_cluster_name")
	if authClusterNameResult.Exists() {
		return authClusterNameResult.String()
	} else {
		proxywasm.LogCriticalf("Configuration for 'auth_cluster_name' wasn't set in config:%s", configuration)
	}
	return ""
}

func getAuthTimeout(configuration gjson.Result) int64 {
	authTimeoutResult := configuration.Get("auth_timeout_ms")
	if authTimeoutResult.Exists() {
		return authTimeoutResult.Int()
	} else {
		proxywasm.LogCriticalf("Configuration for 'auth_timeout_ms' wasn't set in config:%s", configuration)
	}
	return AuthTimeoutDefault.Milliseconds()
}

func getAuthAuthority(configuration gjson.Result) string {
	authAuthorityResult := configuration.Get("auth_authority")
	if authAuthorityResult.Exists() {
		return authAuthorityResult.String()
	} else {
		proxywasm.LogCriticalf("Configuration for 'auth_authority' wasn't set in config:%s", configuration)
	}
	return ""
}

func getNamespace() string {
	// Try reading Staging/Production ENV var
	if namespace, exists := os.LookupEnv("POD_NAMESPACE"); exists {
		return namespace
	}
	proxywasm.LogWarnf("Failed to determine the namespace")
	return ""
}
