package dao

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gofurry/fiberx/v3/heavy/internal/app/user/models"
	"github.com/gofurry/fiberx/v3/heavy/internal/infra/db"
	"github.com/gofurry/fiberx/v3/heavy/pkg/common"
	"gorm.io/gorm"
)

type userDao struct{}

var userDaoSingleton = new(userDao)

func GetUserDao() *userDao { return userDaoSingleton }

type UserListFilter struct {
	PageNum  int
	PageSize int
	Keyword  string
}

func (dao *userDao) Create(user *models.User) common.Error {
	engine, err := engine()
	if err != nil {
		return err
	}

	if err := engine.Create(user).Error; err != nil {
		return mapDatabaseError(err)
	}
	return nil
}

func (dao *userDao) GetByID(id int64) (*models.User, common.Error) {
	engine, err := engine()
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := engine.First(&user, id).Error; err != nil {
		return nil, mapDatabaseError(err)
	}
	return &user, nil
}

func (dao *userDao) Update(user *models.User) common.Error {
	engine, err := engine()
	if err != nil {
		return err
	}

	if err := engine.Save(user).Error; err != nil {
		return mapDatabaseError(err)
	}
	return nil
}

func (dao *userDao) Delete(id int64) common.Error {
	engine, err := engine()
	if err != nil {
		return err
	}

	result := engine.Delete(&models.User{}, id)
	if result.Error != nil {
		return mapDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		return common.NewError(common.RETURN_FAILED, http.StatusNotFound, "user not found")
	}
	return nil
}

func (dao *userDao) List(filter UserListFilter) ([]models.User, int64, common.Error) {
	engine, err := engine()
	if err != nil {
		return nil, 0, err
	}

	query := engine.Model(&models.User{})
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR email LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, mapDatabaseError(err)
	}
	if total == 0 {
		return []models.User{}, 0, nil
	}

	var users []models.User
	offset := (filter.PageNum - 1) * filter.PageSize
	if err := query.Order("id DESC").Offset(offset).Limit(filter.PageSize).Find(&users).Error; err != nil {
		return nil, 0, mapDatabaseError(err)
	}

	return users, total, nil
}

func engine() (*gorm.DB, common.Error) {
	engine := db.Orm.DB()
	if engine == nil {
		return nil, common.NewDaoError("database is not initialized")
	}
	return engine, nil
}

func mapDatabaseError(err error) common.Error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return common.NewError(common.RETURN_FAILED, http.StatusNotFound, "user not found")
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return common.NewError(common.RETURN_FAILED, http.StatusConflict, "user email already exists")
	case strings.Contains(strings.ToLower(err.Error()), "unique"):
		return common.NewError(common.RETURN_FAILED, http.StatusConflict, "user email already exists")
	default:
		return common.NewDaoError(err.Error())
	}
}
