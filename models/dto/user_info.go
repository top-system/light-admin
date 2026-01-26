package dto

// CurrentUserInfo 当前登录用户信息
type CurrentUserInfo struct {
	UserID          uint64   `json:"userId"`
	Username        string   `json:"username"`
	Nickname        string   `json:"nickname"`
	Avatar          string   `json:"avatar"`
	CanSwitchTenant bool     `json:"canSwitchTenant"`
	Roles           []string `json:"roles"`
	Perms           []string `json:"perms"`
}
