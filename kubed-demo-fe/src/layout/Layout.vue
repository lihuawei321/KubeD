<template>
    <a-layout>
        <a-affix>
            <a-layout-header>
                <div style="float:left;display:flex;align-items:center;height:64px;">
                    <img style="height:40px;display:block;" :src="kubeLogo"/>
                    <span style="font-size:25px;padding:0 50px 0 20px;font-weight:bold;color:white;line-height:1;">KubeD</span>
                </div>
                <!-- 集群列表 -->
                <a-menu
                    style="float: left;width:250px;line-height:64px;"
                    v-model:selected-keys="selectedKeys1"
                    theme="dark"
                    mode="horizontal">
                    <a-menu-item v-for="item in clusterList" :key="item">
                        {{ item }}
                    </a-menu-item>
               </a-menu>
               <!-- 用户信息 -->
               <div style="float: right;">
                    <img style="height:35px;border-radius: 50%;margin-right: 10px;line-height:64px;" :src="avator"/>
                    <a-dropdown :overlay-style="{paddingTop: '20px'}">
                        <a>
                            Admin
                            <down-outlined/>
                        </a>
                        <template #overlay>
                            <a-menu>
                                <a-menu-item>
                                    <a @click="logout()">Logout</a>
                                </a-menu-item>
                                <a-menu-item>
                                    <a>Change Password</a>
                                </a-menu-item>
                            </a-menu>
                        </template>
                    </a-dropdown>
               </div>
            </a-layout-header>
        </a-affix>
        <a-layout style="height: calc(100vh - 68px);">
            <!-- 左侧菜单 -->
            <a-layout-sider
                width="240"
                v-model:collapsed="collapsed"
                collapsible
            >
            <!-- selectedKeys表示点击选中的栏目,用于a-menu-item -->
            <!-- openKeys表示展开的栏目，用于a-sub-menu -->
            <!-- openChange事件监听 SubMenu 展开/关闭的回调 -->
            <a-menu
                :selected-keys="selectedKeys2"
                :open-keys="openKeys"
                @openChange="onOpenChange"
                mode="inline"
                :style="{height: '100%',boderRight: 0}">
            <!-- routers是router/index.js中的routes,用于生成侧边栏的菜单 -->
            <template v-for="menu in routers" :key="menu.path">
                <!-- 仅一个子路由：直接作为菜单项 -->
                <a-menu-item
                    v-if="menu.children && menu.children.length === 1"
                    :key="menu.children[0].path"
                    :index="menu.children[0].path"
                    @click="routeChange('item', menu.children[0].path)"
                >
                    <template #icon>
                        <component :is="menu.children[0].icon" v-if="menu.children[0].icon" />
                    </template>
                    <span>{{ menu.children[0].name }}</span>
                </a-menu-item>
                <!-- 多个子路由：折叠子菜单 -->
                <a-sub-menu
                    v-else-if="menu.children && menu.children.length > 1"
                    :key="menu.path"
                    :index="menu.path"
                >
                    <template #icon>
                        <component :is="menu.icon" v-if="menu.icon" />
                    </template>
                    <template #title>
                        <span>
                            <span :class="[collapsed ? 'is-collapse' : '']">{{ menu.name }}</span>
                        </span>
                    </template>
                    <a-menu-item
                        v-for="child in menu.children"
                        :key="child.path"
                        :index="child.path"
                        @click="routeChange('sub', child.path)"
                    >
                        <span>{{ child.name }}</span>
                    </a-menu-item>
                </a-sub-menu>
            </template>
            </a-menu>
            </a-layout-sider>
            <a-layout style="padding: 0 24px">
                <a-breadcrumb style="margin: 10px 0">
                    <a-breadcrumb-item>工作台</a-breadcrumb-item>
                    <!-- router.currentRoute.value.matched表示路由的match信息，能拿到父路由和子路由的信息 -->
                    <template v-for="(matched,index) in router.currentRoute.value.matched" :key="index">
                        <a-breadcrumb-item v-if="matched.name">
                            {{ matched.name }}
                        </a-breadcrumb-item>
                    </template>
                </a-breadcrumb>
                <a-layout-content
                    :style="{
                        background: 'rgb(31, 30, 30)',
                        margin: 0,
                        minHeight: '280px',
                        overflowY: 'auto'
                    }"
                >
                    <router-view></router-view>
                </a-layout-content>
                <a-layout-footer style="text-align: center;">
                    @2026 Created by vv
                </a-layout-footer>
            </a-layout>
        </a-layout>
    </a-layout>
</template>

<script>
import { ref,onMounted } from 'vue'
import kubeLogo from '@/assets/k8s-metrics.png'
import avator from '@/assets/avator.png'
import { useRouter} from 'vue-router'

export default {
    setup() {
        const collapsed = ref(false)
        const selectedKeys1 = ref([])
        const clusterList = ref(['TST-1', 'TST-2'])

        // 侧边栏的属性
        // 路由信息
        const routers = ref([])
        const selectedKeys2 = ref([])
        const openKeys = ref([])

        // 通过useRouter获取路由配置以及当前界面的路由信息
        const router = useRouter()
 


        // 退出登录
        function logout() {
            // 移除用户名
            localStorage.removeItem('username')
            // 移除token
            localStorage.removeItem('token')
            // 跳转登录页
            // router.push('/login')
        }

        // 导航栏点击切换页面以及处理选中的情况
        function routeChange(type,path) {
            // 判断点击是否为sub栏目（也就是单独的item），如果不是，则关闭其他父栏目
            if (type != 'sub') {
                openKeys.value = []
            } 
            // 选中当前path对应的栏目，单独item或者子item
            selectedKeys2.value  = [path]
            // 跳转页面
            // router.currentRoute.value.path获取当前路由的path
            if (router.currentRoute.value.path != path) {
                router.push(path)
            }
            
        }

        // 展开/关闭子栏目
        function onOpenChange(val) {
            // 匹配这个val是否在openKeys中，如果不在，则添加到openKeys中
            const latestOpenKey = val.find(key => openKeys.value.indexOf(key) == -1)
            // 如果latestOpenKey存在，则添加到openKeys中，否则清空openKeys
            openKeys.value = latestOpenKey ? [latestOpenKey] : []
        }

        // 用于从浏览器地址直接打开后的选中
        function getRouter(val){
            selectedKeys2.value = [val[1].path]
            openKeys.value = [val[0].path]
            
        }

        // 生命周期钩子，在组件挂载后执行
        onMounted(() => {
            // 只保留有 children 的布局路由，排除 redirect 等，避免菜单项/icon 为 null 导致卸载报错
            routers.value = router.options.routes 
            console.log(router.currentRoute.value.matched)
            getRouter(router.currentRoute.value.matched)
        })

        return {
            collapsed,
            kubeLogo,
            selectedKeys1,
            clusterList,
            avator,
            routers,
            selectedKeys2,
            openKeys,
            router,
            logout,
            routeChange,
            onOpenChange
        }
    }
}
</script>
<style scoped>
    /* 头部样式 */
    .ant-layout-header {
        padding: 0 30px !important;
    }
    /* 内容区域样式 */
    .ant-layout-content::-webkit-scrollbar {
        width:6px;
    }
    .ant-layout-content::-webkit-scrollbar-track {
        background-color:rgb(164, 162, 162);
    }
    .ant-layout-content::-webkit-scrollbar-thumb {
        background-color:#666;
    }
    .ant-layout-footer {
        padding: 5px 50px !important;
        color: rgb(239, 239, 239);
    }
    .is-collapse {
        display: none;
    }
    .ant-layout-sider {
        background: #141414 !important;
        overflow-y: auto;
    }
    .ant-layout-sider::-webkit-scrollbar {
        display: none;
    }
    .ant-menu-item {
        margin: 0 !important;
    }
</style>