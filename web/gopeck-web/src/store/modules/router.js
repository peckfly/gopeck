import { defineStore } from 'pinia'

import router from '@/router'
import { notFoundRoute } from '@/router/config'
import { addWebPage, formatRoutes, generateMenuList, generateRoutes, getFirstValidRoute } from '@/router/util'
import { findTree } from '@/utils/util'
import { config } from '@/config'
import apis from '@/apis'
import { formatApiData } from '../../router/util'

const useRouterStore = defineStore('router', {
    state: () => ({
        routes: [],
        menuList: [],
        indexRoute: null,
    }),
    getters: {},
    actions: {
        /**
         * 获取路由列表
         * @returns {Promise}
         */
        getRouterList() {
            return new Promise((resolve, reject) => {
                ;(async () => {
                    try {
                        const { success, data } = await apis.user.getUserMenu().catch(() => {
                            throw new Error()
                        })
                        if (config('http.code.success') === success) {
                            const list = formatApiData(data)

                            list.push(...addWebPage())

                            const validRoutes = formatRoutes(list)

                            const menuList = generateMenuList(validRoutes)
                            const routes = [...generateRoutes(validRoutes), notFoundRoute]
                            const indexRoute = getFirstValidRoute(menuList)
                            routes.forEach((route) => {
                                router.addRoute(route)
                            })
                            this.routes = routes
                            this.menuList = menuList
                            this.indexRoute = indexRoute
                            resolve()
                        }
                    } catch (error) {
                        console.log(error)
                        reject()
                    }
                })()
            })
        },
        /**
         * 设置徽标
         * @param {string} name 名称
         * @param {number} count 数量
         */
        setBadge({ name, count } = {}) {
            let menuInfo = null
            findTree(
                this.menuList,
                name,
                (item) => {
                    menuInfo = item
                },
                { key: 'name', children: 'children' }
            )
            if (menuInfo) {
                menuInfo.meta.badge = count
            }
        },
    },
})

export default useRouterStore
