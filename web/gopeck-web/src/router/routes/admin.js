export default [
    {
        path: '/setting',
        name: 'setting',
        component: 'admin/setting/index.vue',
        meta: {
            title: '个人设置',
            isMenu: false,
            keepAlive: false,
            permission: '*',
            active: 'setting',
            openKeys: 'setting',
        },
    },
]
