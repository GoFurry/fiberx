package bootstrap

import "github.com/gofurry/fiberx/v3/extra-light/internal/db"

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
	return db.Ready()
}
