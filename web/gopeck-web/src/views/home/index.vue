<template>
    <a-row :gutter="16">
        <a-col :lg="24">
            <a-card>
                <a-row
                    class="align-center"
                    type="flex">
                    <p
                        class="display-flex align-items-center"
                        style="background-color: #f6ffec; padding: 10px; box-sizing: border-box; width: 100%">
                        <check-circle-filled class="mr-4-1 color-success" />
                        {{ $t('pages.welcome.alertMessage') }}
                    </p>
                </a-row>
            </a-card>
        </a-col>
    </a-row>
</template>

<script setup>
import { CheckCircleFilled } from '@ant-design/icons-vue'
defineOptions({
    name: 'home',
})
// const handleCopy = () => {
//     const clipboard = new ClipboardJS('.copy-btn')
//     clipboard.on('success', (e) => {
//         e.clearSelection()
//         message.success('copy success!')
//     })
//     clipboard.on('error', () => {
//         message.error('copy error!')
//     })
// }
</script>

<style lang="less" scoped>
p {
    box-shadow: 0 0 3px 3px #f9f9f9;
    padding: 20px;
}
</style>
