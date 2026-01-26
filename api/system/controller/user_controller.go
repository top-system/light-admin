package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/top-system/light-admin/api/system/service"
	"github.com/top-system/light-admin/constants"
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/pkg/echox"
	"github.com/labstack/echo/v4"

	"gorm.io/gorm"
)

type UserController struct {
	userService service.UserService
	logger      lib.Logger
}

// NewUserController creates new user controller
func NewUserController(userService service.UserService, logger lib.Logger) UserController {
	return UserController{
		userService: userService,
		logger:      logger,
	}
}

// @tags User
// @summary Get Current User Info
// @produce application/json
// @success 200 {object} echox.Response{data=dto.CurrentUserInfo} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/users/me [get]
func (a UserController) Me(ctx echo.Context) error {
	claims, ok := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	if !ok || claims == nil {
		return echox.Response{Code: http.StatusUnauthorized, Message: errors.AuthTokenInvalid}.JSON(ctx)
	}

	userInfo, err := a.userService.GetCurrentUserInfo(claims.ID, claims.Username)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: userInfo}.JSON(ctx)
}

// @tags User
// @summary User Query
// @produce application/json
// @param data query system.UserQueryParam true "UserQueryParam"
// @success 200 {object} echox.Response{data=system.Users} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/users [get]
func (a UserController) Query(ctx echo.Context) error {
	param := new(system.UserQueryParam)
	if err := ctx.Bind(param); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	if v := ctx.QueryParam("role_ids"); v != "" {
		strIDs := strings.Split(v, ",")
		roleIDs := make([]uint64, 0, len(strIDs))
		for _, s := range strIDs {
			if id, err := strconv.ParseUint(s, 10, 64); err == nil {
				roleIDs = append(roleIDs, id)
			}
		}
		param.RoleIDs = roleIDs
	}

	qr, err := a.userService.Query(param)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{
		Code: http.StatusOK,
		Data: qr.List,
		Page: &echox.PageInfo{
			Total:    qr.Pagination.Total,
			PageNum:  qr.Pagination.PageNum,
			PageSize: qr.Pagination.PageSize,
		},
	}.JSON(ctx)
}

// @tags User
// @summary User Options
// @produce application/json
// @success 200 {object} echox.Response{data=[]system.UserOption} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/users/options [get]
func (a UserController) GetOptions(ctx echo.Context) error {
	options, err := a.userService.ListUserOptions()
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: options}.JSON(ctx)
}

// @tags User
// @summary User Create
// @produce application/json
// @param data body system.User true "User"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/users [post]
func (a UserController) Create(ctx echo.Context) error {
	user := new(system.User)
	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)

	if err := ctx.Bind(user); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	} else if user.Password == "" {
		return echox.Response{Code: http.StatusBadRequest, Message: errors.UserPasswordRequired}.JSON(ctx)
	}

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	user.CreateBy = claims.ID

	qr, err := a.userService.WithTrx(trxHandle).Create(user)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: qr}.JSON(ctx)
}

// @tags User
// @summary Get User Form By ID
// @produce application/json
// @param id path int true "user id"
// @success 200 {object} echox.Response{data=system.UserForm} "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/users/{id}/form [get]
func (a UserController) GetForm(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	form, err := a.userService.GetUserForm(id)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK, Data: form}.JSON(ctx)
}

// @tags User
// @summary User Update By ID
// @produce application/json
// @param id path int true "user id"
// @param data body system.User true "User"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/users/{id} [put]
func (a UserController) Update(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	user := new(system.User)
	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)

	if err := ctx.Bind(user); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	claims, _ := ctx.Get(constants.CurrentUser).(*dto.JwtClaims)
	user.UpdateBy = claims.ID

	err = a.userService.WithTrx(trxHandle).Update(id, user)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// @tags User
// @summary User Delete By ID
// @produce application/json
// @param id path int true "user id"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/users/{id} [delete]
func (a UserController) Delete(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	trxHandle := ctx.Get(constants.DBTransaction).(*gorm.DB)
	err = a.userService.WithTrx(trxHandle).Delete(id)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// @tags User
// @summary Reset User Password
// @produce application/json
// @param id path int true "user id"
// @param password query string true "new password"
// @success 200 {object} echox.Response "ok"
// @failure 400 {object} echox.Response "bad request"
// @failure 500 {object} echox.Response "internal error"
// @router /api/v1/users/{id}/password/reset [put]
func (a UserController) ResetPassword(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	password := ctx.QueryParam("password")
	if password == "" {
		return echox.Response{Code: http.StatusBadRequest, Message: errors.UserPasswordRequired}.JSON(ctx)
	}

	err = a.userService.ResetPassword(id, password)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}

	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}
