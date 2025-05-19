package v0_test

import (
	"context"
	"os"
	"testing"

	"github.com/reddit/baseplate.go/log"
	"github.com/reddit/baseplate.go/secrets"
	"github.com/reddit/edgecontext/lib/go/internal/v0"
)

var globalTestImpl *v0.Impl

func TestMain(m *testing.M) {
	store, _, err := secrets.NewTestSecrets(
		context.Background(),
		make(map[string]secrets.GenericSecret),
	)
	if err != nil {
		log.Panic(err)
	}
	defer store.Close()

	globalTestImpl = v0.Init(v0.Config{Store: store})
	os.Exit(m.Run())
}
