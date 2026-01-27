package dto

// CurrentUserInfo 当前登录用户信息
type CurrentUserInfo struct {
	UserID          uint64   `json:"userId"`
	Username        string   `json:"username"`
	Nickname        string   `json:"nickname"`
	Avatar          string   `json:"avatar"`
	Gender          int      `json:"gender"`
	Mobile          string   `json:"mobile"`
	Email           string   `json:"email"`
	DeptName        string   `json:"deptName"`
	CreateTime      DateTime `json:"createTime"`
	CanSwitchTenant bool     `json:"canSwitchTenant"`
	Roles           []string `json:"roles"`
	Perms           []string `json:"perms"`
}
