package internal

import (
	"reflect"
	"testing"
)

func TestNewConfig(t *testing.T) {
	vmContext := InitPlugin(t)

	tests := []struct {
		name         string
		pluginConfig string
		want         Config
	}{
		{
			name: "reads auth Settings",
			pluginConfig: `
			{
				"auth_cluster_name": "auth",
				"auth_authority": "auth",
				"auth_timeout_ms": 5
			}`,
			want: Config{
				AuthClusterName: "auth",
				AuthAuthority:   "auth",
				AuthTimeout:     5,
				Namespace:       "example-service",
			},
		},
		{
			name: "reads service-specific settings",
			pluginConfig: `
			{
				"auth_cluster_name": "auth",
				"auth_authority": "auth",
				"auth_timeout_ms": 5,
				"example-service": {
					"enabled": true,
					"lightweight_system_auth": true
				}
			}`,
			want: Config{
				AuthClusterName: "auth",
				AuthAuthority:   "auth",
				AuthTimeout:     5,
				Namespace:       "example-service",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize new plugin
			host, _, reset := NewContextWithConfig(t, vmContext, tt.pluginConfig)
			defer reset()

			_ = host.SetProperty([]string{"POD_NAMESPACE"}, []byte("example-service"))
			if got := NewConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
