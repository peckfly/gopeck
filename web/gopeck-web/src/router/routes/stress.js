import { PlaySquareOutlined } from '@ant-design/icons-vue'

export default [
    {
        path: 'stress',
        name: 'stress',
        component: 'RouteViewLayout',
        meta: {
            icon: PlaySquareOutlined,
            title: '性能测试',
            isMenu: true,
            keepAlive: true,
            permission: '*',
        },
        children: [
            {
                path: 'start',
                name: 'start',
                component: 'stress/start/index.vue',
                meta: {
                    title: '发起压测',
                    isMenu: true,
                    keepAlive: true,
                    permission: '*',
                },
            },
            {
                path: 'record',
                name: 'record',
                component: 'stress/record/index.vue',
                meta: {
                    title: '压测记录',
                    isMenu: true,
                    keepAlive: true,
                    permission: '*',
                },
            },
            {
                path: 'node',
                name: 'node',
                component: 'stress/node/index.vue',
                meta: {
                    title: '机器管理',
                    isMenu: true,
                    keepAlive: true,
                    permission: '*',
                },
            },
        ],
    },
]
