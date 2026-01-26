package dto

// RouteVO 路由视图对象
type RouteVO struct {
	Name      string     `json:"name"`               // 路由名称
	Path      string     `json:"path"`               // 路由路径
	Redirect  string     `json:"redirect,omitempty"` // 重定向路径
	Component string     `json:"component"`          // 组件路径
	Meta      RouteMeta  `json:"meta"`               // 路由元信息
	Children  []*RouteVO `json:"children,omitempty"` // 子路由
}

// RouteMeta 路由元信息
type RouteMeta struct {
	Title      string            `json:"title"`                // 菜单标题
	Icon       string            `json:"icon"`                 // 菜单图标
	Hidden     bool              `json:"hidden"`               // 是否隐藏
	KeepAlive  bool              `json:"keepAlive,omitempty"`  // 是否缓存
	AlwaysShow bool              `json:"alwaysShow,omitempty"` // 是否总是显示
	Params     map[string]string `json:"params,omitempty"`     // 路由参数
}

// MenuOption 菜单下拉选项
type MenuOption struct {
	Value    uint64       `json:"value"`
	Label    string       `json:"label"`
	Children []MenuOption `json:"children,omitempty"`
}
