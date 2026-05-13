package controller

import (
	"encoding/json"
	"strconv"

	"github.com/gofurry/fiberx/v3/light/internal/app/user/models"
	"github.com/gofurry/fiberx/v3/light/internal/app/user/service"
	"github.com/gofurry/fiberx/v3/light/pkg/common"
	"github.com/gofiber/fiber/v3"
)

type userAPI struct{}

var UserApi = &userAPI{}

func (api *userAPI) CreateUser(c fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := decodeJSONBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}

	data, err := service.GetUserService().Create(req)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}

	return common.NewResponse(c).SuccessWithData(data)
}

func (api *userAPI) GetUser(c fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}

	data, serviceErr := service.GetUserService().GetByID(id)
	if serviceErr != nil {
		return common.NewResponse(c).Error(serviceErr)
	}

	return common.NewResponse(c).SuccessWithData(data)
}

func (api *userAPI) UpdateUser(c fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}

	var req models.UpdateUserRequest
	if err := decodeJSONBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}

	data, serviceErr := service.GetUserService().Update(id, req)
	if serviceErr != nil {
		return common.NewResponse(c).Error(serviceErr)
	}

	return common.NewResponse(c).SuccessWithData(data)
}

func (api *userAPI) DeleteUser(c fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}

	if serviceErr := service.GetUserService().Delete(id); serviceErr != nil {
		return common.NewResponse(c).Error(serviceErr)
	}

	return common.NewResponse(c).SuccessWithData(fiber.Map{
		"deleted": true,
		"id":      id,
	})
}

func (api *userAPI) ListUsers(c fiber.Ctx) error {
	req := models.ListUsersRequest{
		PageNum:  parseQueryInt(c, "page_num", 1),
		PageSize: parseQueryInt(c, "page_size", 10),
		Keyword:  c.Query("keyword", ""),
	}

	data, err := service.GetUserService().List(req)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}

	return common.NewResponse(c).SuccessWithData(data)
}

func decodeJSONBody(c fiber.Ctx, target any) common.Error {
	if err := json.Unmarshal(c.Body(), target); err != nil {
		return common.NewValidationError("request body must be valid json")
	}

	return nil
}

func parseIDParam(c fiber.Ctx) (int64, common.Error) {
	id, err := strconv.ParseInt(c.Params("id", "0"), 10, 64)
	if err != nil || id <= 0 {
		return 0, common.NewValidationError("id must be a positive integer")
	}

	return id, nil
}

func parseQueryInt(c fiber.Ctx, key string, fallback int) int {
	raw := c.Query(key, "")
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return value
}
