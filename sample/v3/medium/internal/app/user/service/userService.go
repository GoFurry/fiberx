package service

import (
	"net/http"
	"strings"

	"github.com/gofurry/fiberx/v3/medium/internal/app/user/dao"
	"github.com/gofurry/fiberx/v3/medium/internal/app/user/models"
	"github.com/gofurry/fiberx/v3/medium/pkg/common"
	pkgmodels "github.com/gofurry/fiberx/v3/medium/pkg/models"
)

type userService struct{}

var userSingleton = new(userService)

func GetUserService() *userService { return userSingleton }

func (s *userService) Create(req models.CreateUserRequest) (*models.User, common.Error) {
	user, err := buildUser(req.Name, req.Email, req.Age, req.Status)
	if err != nil {
		return nil, err
	}

	if err := dao.GetUserDao().Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetByID(id int64) (*models.User, common.Error) {
	if id <= 0 {
		return nil, common.NewValidationError("id must be greater than 0")
	}

	return dao.GetUserDao().GetByID(id)
}

func (s *userService) Update(id int64, req models.UpdateUserRequest) (*models.User, common.Error) {
	if id <= 0 {
		return nil, common.NewValidationError("id must be greater than 0")
	}

	user, err := dao.GetUserDao().GetByID(id)
	if err != nil {
		return nil, err
	}

	updated, buildErr := buildUser(req.Name, req.Email, req.Age, req.Status)
	if buildErr != nil {
		return nil, buildErr
	}

	user.Name = updated.Name
	user.Email = updated.Email
	user.Age = updated.Age
	user.Status = updated.Status

	if err := dao.GetUserDao().Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Delete(id int64) common.Error {
	if id <= 0 {
		return common.NewValidationError("id must be greater than 0")
	}

	return dao.GetUserDao().Delete(id)
}

func (s *userService) List(req models.ListUsersRequest) (pkgmodels.PageResponse, common.Error) {
	if req.PageNum <= 0 {
		req.PageNum = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		return pkgmodels.PageResponse{}, common.NewError(common.RETURN_FAILED, http.StatusBadRequest, "page_size must be less than or equal to 100")
	}

	users, total, err := dao.GetUserDao().List(dao.UserListFilter{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Keyword:  strings.TrimSpace(req.Keyword),
	})
	if err != nil {
		return pkgmodels.PageResponse{}, err
	}

	return pkgmodels.PageResponse{
		Total: total,
		Data:  users,
	}, nil
}

func buildUser(name, email string, age int, status string) (*models.User, common.Error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	status = strings.TrimSpace(status)

	if name == "" {
		return nil, common.NewValidationError("name is required")
	}
	if email == "" {
		return nil, common.NewValidationError("email is required")
	}
	if !strings.Contains(email, "@") {
		return nil, common.NewValidationError("email format is invalid")
	}
	if age < 0 {
		return nil, common.NewValidationError("age must be greater than or equal to 0")
	}
	if status == "" {
		status = "active"
	}

	return &models.User{
		Name:   name,
		Email:  email,
		Age:    age,
		Status: status,
	}, nil
}
