package internal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/proxytest"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

const DefaultTestConfig = `
{
	"auth_cluster_name": "auth", 
	"auth_authority": "auth", 
	"auth_timeout_ms": 5,
	"example-service": {
		"enabled": true,
		"lightweight_system_auth": true
	}
}`

func InitPlugin(t *testing.T) proxytest.WasmVMContext {
	wasm, err := os.ReadFile("../main.wasm")
	if err != nil {
		t.Fatalf("wasm not found")
	}
	vmContext, err := proxytest.NewWasmVMContext(wasm)
	require.NoError(t, err)
	return vmContext
}

func NewContextWithConfig(t *testing.T, vmContext proxytest.WasmVMContext, config string) (proxytest.HostEmulator, uint32, func()) {
	opt := proxytest.
		NewEmulatorOption().
		WithPluginConfiguration([]byte(config)).
		WithVMContext(vmContext)
	host, reset := proxytest.NewHostEmulator(opt)

	// Call OnVMStart.
	require.Equal(t, types.OnVMStartStatusOK, host.StartVM())

	// Set POD_NAMESPACE and call OnPluginStart to read config
	_ = host.SetProperty([]string{"POD_NAMESPACE"}, []byte("example-service"))
	require.Equal(t, host.StartPlugin(), types.OnPluginStartStatusOK)

	// Initialize http context.
	return host, host.InitializeHttpContext(), reset
}

func NewContext(t *testing.T, vmContext proxytest.WasmVMContext) (proxytest.HostEmulator, uint32, func()) {
	return NewContextWithConfig(t, vmContext, DefaultTestConfig)
}
