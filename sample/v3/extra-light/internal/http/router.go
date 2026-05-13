package http

import (
	"errors"
	"io/fs"
	"path"
	"strings"

	env "github.com/gofurry/fiberx/v3/extra-light/config"
	"github.com/gofurry/fiberx/v3/extra-light/internal/bootstrap"
	"github.com/gofurry/fiberx/v3/extra-light/internal/http/webui"
	"github.com/gofurry/fiberx/v3/extra-light/pkg/common"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func New() *fiber.App {
	appName := common.COMMON_PROJECT_NAME
	if name := env.GetServerConfig().Server.AppName; strings.TrimSpace(name) != "" {
		appName = name
	}

	app := fiber.New(fiber.Config{
		AppName:      appName,
		ServerHeader: appName,
		ErrorHandler: customErrorHandler,
	})

	app.Use(recover.New())
	registerHealthRoutes(app)
	api(app.Group("/api"))
	return app
}

func registerHealthRoutes(app *fiber.App) {
	app.Get(healthcheck.LivenessEndpoint, healthcheck.New(healthcheck.Config{
		Probe: func(c fiber.Ctx) bool { return bootstrap.Live() },
	}))
	app.Get(healthcheck.ReadinessEndpoint, healthcheck.New(healthcheck.Config{
		Probe: func(c fiber.Ctx) bool { return bootstrap.Ready() },
	}))
	app.Get(healthcheck.StartupEndpoint, healthcheck.New(healthcheck.Config{
		Probe: func(c fiber.Ctx) bool { return bootstrap.Started() },
	}))
	app.Get("/healthz", func(c fiber.Ctx) error {
		ready := bootstrap.Ready()
		statusCode := fiber.StatusOK
		status := "ok"
		if !ready {
			statusCode = fiber.StatusServiceUnavailable
			status = "not_ready"
		}
		return c.Status(statusCode).JSON(common.ResultData{
			Code:    common.RETURN_SUCCESS,
			Message: "success",
			Data: fiber.Map{
				"status":  status,
				"live":    bootstrap.Live(),
				"ready":   ready,
				"startup": bootstrap.Started(),
			},
		})
	})
}

func AttachEmbeddedUI(app *fiber.App) {
	uiFS, err := fs.Sub(webui.FS, "dist")
	if err != nil {
		return
	}
	index, err := fs.ReadFile(uiFS, "index.html")
	if err != nil {
		return
	}

	sendIndex := func(c fiber.Ctx) error {
		c.Type("html", "utf-8")
		return c.Send(index)
	}

	app.Use(func(c fiber.Ctx) error {
		if c.Method() != fiber.MethodGet && c.Method() != fiber.MethodHead {
			return fiber.ErrNotFound
		}

		reqPath := c.Path()
		if reqPath == "/api" || strings.HasPrefix(reqPath, "/api/") || reqPath == "/v1" || strings.HasPrefix(reqPath, "/v1/") {
			return fiber.ErrNotFound
		}
		if reqPath == "/" || reqPath == "" {
			return sendIndex(c)
		}

		cleaned := path.Clean(reqPath)
		cleaned = strings.TrimPrefix(cleaned, "/")
		if cleaned == "." || cleaned == "" {
			return sendIndex(c)
		}

		if stat, err := fs.Stat(uiFS, cleaned); err == nil && !stat.IsDir() {
			return c.SendFile(cleaned, fiber.SendFile{FS: uiFS})
		}

		return sendIndex(c)
	})
}

func Run() error {
	cfg := env.GetServerConfig().Server
	app := New()
	defer func() {
		_ = app.Shutdown()
		_ = bootstrap.Shutdown()
	}()
	return app.Listen(cfg.IPAddress + ":" + cfg.Port)
}

func customErrorHandler(c fiber.Ctx, err error) error {
	if appErr, ok := errors.AsType[common.Error](err); ok {
		return common.NewResponse(c).ErrorWithCode(appErr, appErr.GetHTTPStatus())
	}

	code := fiber.StatusInternalServerError
	if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
		code = fiberErr.Code
	}

	response := common.NewResponse(c)
	switch code {
	case fiber.StatusNotFound:
		return response.ErrorWithCode(common.NewError(common.RETURN_FAILED, code, "resource not found"), code)
	case fiber.StatusMethodNotAllowed:
		return response.ErrorWithCode(common.NewError(common.RETURN_FAILED, code, "method not allowed"), code)
	default:
		return response.ErrorWithCode(common.NewError(common.RETURN_FAILED, code, "internal server error"), code)
	}
}
