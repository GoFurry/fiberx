package fiberx

import "github.com/gofurry/fiberx/internal/core"

func Generate(req Request) error {
	return core.Generate(req)
}
