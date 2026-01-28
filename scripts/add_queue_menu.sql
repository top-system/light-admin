-- 添加队列管理菜单
-- 首先获取系统管理目录的ID (假设是 ID=1，根据实际情况调整)
-- 如果要添加为独立的顶级菜单，设置 parent_id = 0

-- 添加队列管理菜单 (作为系统管理的子菜单)
-- 注意: 需要根据实际的 parent_id 和 tree_path 进行调整
-- type: 1-菜单 2-目录 3-外链 4-按钮
-- visible: 1-显示 0-隐藏

-- 添加队列管理页面菜单
INSERT INTO `t_menu` (
    `parent_id`, `tree_path`, `name`, `type`, `route_name`, `route_path`,
    `component`, `perm`, `always_show`, `keep_alive`, `visible`, `sort`,
    `icon`, `redirect`, `params`, `create_time`, `update_time`
) VALUES (
    1,                          -- parent_id: 系统管理目录的ID (需要根据实际情况调整)
    '0,1',                      -- tree_path: 需要根据实际的parent_id调整
    '队列管理',                 -- name: 菜单名称
    1,                          -- type: 1-菜单
    'QueueList',                -- route_name: 路由名称
    'queue',                    -- route_path: 路由路径
    'queue/index',              -- component: 组件路径
    'sys:task:query',           -- perm: 权限标识
    0,                          -- always_show: 不总是显示
    1,                          -- keep_alive: 保持缓存
    1,                          -- visible: 显示
    100,                        -- sort: 排序值
    'menu',                     -- icon: 图标
    '',                         -- redirect: 重定向
    '',                         -- params: 参数
    NOW(),                      -- create_time
    NOW()                       -- update_time
);

-- 获取刚插入的菜单ID
SET @queue_menu_id = LAST_INSERT_ID();

-- 添加查询按钮权限
INSERT INTO `t_menu` (
    `parent_id`, `tree_path`, `name`, `type`, `route_name`, `route_path`,
    `component`, `perm`, `always_show`, `keep_alive`, `visible`, `sort`,
    `icon`, `redirect`, `params`, `create_time`, `update_time`
) VALUES (
    @queue_menu_id,
    CONCAT('0,1,', @queue_menu_id),
    '查询',
    4,                          -- type: 4-按钮
    '',
    '',
    '',
    'sys:task:query',
    0,
    0,
    1,
    1,
    '',
    '',
    '',
    NOW(),
    NOW()
);

-- 添加删除按钮权限
INSERT INTO `t_menu` (
    `parent_id`, `tree_path`, `name`, `type`, `route_name`, `route_path`,
    `component`, `perm`, `always_show`, `keep_alive`, `visible`, `sort`,
    `icon`, `redirect`, `params`, `create_time`, `update_time`
) VALUES (
    @queue_menu_id,
    CONCAT('0,1,', @queue_menu_id),
    '删除',
    4,                          -- type: 4-按钮
    '',
    '',
    '',
    'sys:task:delete',
    0,
    0,
    1,
    2,
    '',
    '',
    '',
    NOW(),
    NOW()
);

-- 将菜单分配给管理员角色 (假设管理员角色ID为1)
-- 注意: 需要根据实际的角色ID进行调整
INSERT INTO `t_role_menu` (`role_id`, `menu_id`)
SELECT 1, id FROM `t_menu` WHERE `perm` LIKE 'sys:task:%';

-- 输出提示
SELECT '队列管理菜单添加完成！请确认以下菜单已正确添加:' AS message;
SELECT id, parent_id, name, type, route_path, component, perm FROM `t_menu` WHERE `perm` LIKE 'sys:task:%';
