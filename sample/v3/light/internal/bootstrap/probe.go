package bootstrap

import (
	env "github.com/gofurry/fiberx/v3/light/config"
	"github.com/gofurry/fiberx/v3/light/internal/infra/db"
)

func Live() bool {
	return true
}

func Started() bool {
	return started.Load()
}

func Ready() bool {
	if !Started() {
		return false
	}

	cfg := env.GetServerConfig()
	if cfg.DataBase.Enabled && !db.Orm.Ready() {
		return false
	}
	return true
}
