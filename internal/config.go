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

// Config is used to extract any WASMPlugin configuration defined in the deployed YML
type Config struct {
	AuthClusterName string
	AuthAuthority   string
	AuthTimeout     uint32
	Namespace       string
}

func NewConfig() *Config {
	configuration := getPluginConfiguration()
	namespace := getNamespace()
	config := Config{
		AuthClusterName: getStringFromConfig(configuration, "auth_cluster_name"),
		AuthAuthority:   getStringFromConfig(configuration, "auth_authority"),
		AuthTimeout:     uint32(getInt64FromConfig(configuration, "auth_timeout_ms", AuthTimeoutDefault.Milliseconds())),
		Namespace:       namespace,
	}

	return &config
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

func getNamespace() string {
	// Try reading Staging/Production ENV var
	if namespace, exists := os.LookupEnv("POD_NAMESPACE"); exists {
		return namespace
	}
	// Test fallback
	if namespace, err := proxywasm.GetProperty([]string{"POD_NAMESPACE"}); err == nil {
		return string(namespace)
	}

	proxywasm.LogWarnf("Failed to determine the namespace")
	return ""
}

func getStringFromConfig(configuration gjson.Result, key string) string {
	result := configuration.Get(key)
	if result.Exists() {
		return result.String()
	}
	proxywasm.LogCriticalf("Configuration for '%s' wasn't set in config:%s", key, configuration)
	return ""
}

func getInt64FromConfig(configuration gjson.Result, key string, defaultResult int64) int64 {
	result := configuration.Get(key)
	if result.Exists() {
		return result.Int()
	}
	proxywasm.LogCriticalf("Configuration for '%s' wasn't set in config:%s", key, configuration)
	return defaultResult
}
