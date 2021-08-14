package testutils

import (
	"io"
	"net"
	"testing"

	"github.com/hashicorp/go-hclog"
	vaultAPI "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/builtin/logical/transit"
	vaultHTTP "github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/hashicorp/vault/vault"
	"github.com/stretchr/testify/require"
)

func SetupVault(t *testing.T) (net.Listener, *vaultAPI.Client) {
	vaultLogger := hclog.New(&hclog.LoggerOptions{
		Output: io.Discard,
	})
	coreConfig := &vault.CoreConfig{
		LogicalBackends: map[string]logical.Factory{
			"transit": transit.Factory,
		},
		Logger: vaultLogger,
	}
	core, _, rootToken := vault.TestCoreUnsealedWithConfig(t, coreConfig)
	ln, addr := vaultHTTP.TestServer(t, core)
	conf := vaultAPI.DefaultConfig()
	conf.Address = addr
	vaultClient, err := vaultAPI.NewClient(conf)
	require.Nil(t, err)
	vaultClient.SetToken(rootToken)
	err = vaultClient.Sys().Mount("transit", &vaultAPI.MountInput{
		Type: "transit",
	})
	require.Nil(t, err)
	_, err = vaultClient.Logical().Write("transit/keys/ledger_translator", map[string]interface{}{})
	require.Nil(t, err)
	return ln, vaultClient
}
