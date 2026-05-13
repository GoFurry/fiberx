package router

import (
	user "github.com/gofurry/fiberx/v3/heavy/internal/app/user/controller"
	"github.com/gofiber/fiber/v3"
)

func api(root fiber.Router) {
	v1(root.Group("/v1"))
}

func v1(root fiber.Router) {
	userRoutes(root.Group("/user"))
}

func userRoutes(root fiber.Router) {
	root.Post("/", user.UserApi.CreateUser)
	root.Get("/", user.UserApi.ListUsers)
	root.Get("/:id", user.UserApi.GetUser)
	root.Put("/:id", user.UserApi.UpdateUser)
	root.Delete("/:id", user.UserApi.DeleteUser)
}
