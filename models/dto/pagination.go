package dto

type Pagination struct {
	Total    int64 `json:"total"`
	PageNum  int   `json:"pageNum"`
	PageSize int   `json:"pageSize"`
}

type PaginationParam struct {
	PageNum  int `query:"pageNum"`
	PageSize int `query:"pageSize" validate:"max=128"`
}

func (a *PaginationParam) GetPageNum() int {
	return a.PageNum
}

func (a *PaginationParam) GetPageSize() int {
	pageSize := a.PageSize
	if a.PageSize == 0 {
		pageSize = 15
	}

	return pageSize
}
