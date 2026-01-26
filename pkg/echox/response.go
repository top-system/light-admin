package echox

import (
	"net/http"

	"github.com/top-system/light-admin/errors"

	"github.com/labstack/echo/v4"
)

// Response in order to unify the returned response structure
type Response struct {
	Code    int         `json:"-"`
	Pretty  bool        `json:"-"`
	BizCode string      `json:"code"`
	Data    interface{} `json:"data,omitempty"`
	Page    *PageInfo   `json:"page,omitempty"`
	Message interface{} `json:"message"`
}

// PageInfo 分页信息
type PageInfo struct {
	Total    int64 `json:"total"`
	PageNum  int   `json:"pageNum"`
	PageSize int   `json:"pageSize"`
}

// sends a JSON response with status code.
func (a Response) JSON(ctx echo.Context) error {
	if a.Message == "" || a.Message == nil {
		a.Message = http.StatusText(a.Code)
	}

	if err, ok := a.Message.(error); ok {
		if errors.Is(err, errors.DatabaseInternalError) {
			a.Code = http.StatusInternalServerError
		}

		if errors.Is(err, errors.DatabaseRecordNotFound) {
			a.Code = http.StatusNotFound
		}

		a.Message = err.Error()
	}

	// 自动设置业务码：成功=00000，失败=A
	if a.BizCode == "" {
		if a.Code == http.StatusOK {
			a.BizCode = "00000"
		} else {
			a.BizCode = "A"
		}
	}

	if a.Pretty {
		return ctx.JSONPretty(a.Code, a, "\t")
	}

	return ctx.JSON(a.Code, a)
}
