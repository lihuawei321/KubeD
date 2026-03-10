<template>
    <a-config-provider :theme="themeConfig">
        <router-view></router-view>
    </a-config-provider>
</template>

<script>
import { theme } from 'ant-design-vue'

export default {
    name: 'App',
    setup() {
        // 可选: defaultAlgorithm(亮色) | theme.darkAlgorithm(暗色) | theme.compactAlgorithm(紧凑)
        const themeConfig = {
            algorithm: theme.darkAlgorithm
        }
        return { themeConfig }
    }
}
</script>

<style>
    html, body {
        width: 100vw;
        height: 100vh;
        min-width: 1425px;
        overflow-y: hidden;
    }
    #nprogress .bar {
        /*自定义进度条颜色*/
        background: #2186c0 !important;
    }
</style>