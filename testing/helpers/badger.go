package helpers

import (
	"github.com/onflow/flow-go/model/flow"
	"testing"

	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/require"
)

func InMemoryDB(t *testing.T) *badger.DB {
	t.Helper()

	opts := badger.DefaultOptions("")
	opts.InMemory = true
	opts.Logger = nil

	db, err := badger.Open(opts)
	require.NoError(t, err)

	return db
}

//accounts of interests for debug
var DebugAccounts = []string{
	"8624b52f9ddcd04a",
	"d796ff17107bbff6",
}

func IsDebugAccount(address flow.Address) bool {
	for _, account := range DebugAccounts {
		if address.Hex() == account {
			return true
		}
	}
	return false
}
