<template>
    <a-row
        :gutter="48"
        type="flex">
        <a-col flex="0 0 480px">
            <a-form layout="vertical">
                <a-form-item :label="t('pages.user.profile.tab.basic.form.username')">
                    <a-input></a-input>
                </a-form-item>
                <a-form-item :label="t('pages.user.profile.tab.basic.form.email')">
                    <a-input></a-input>
                </a-form-item>
            </a-form>
        </a-col>
    </a-row>
</template>

<script setup>
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
// todo call api to update and show
defineOptions({
    name: 'Basic',
})
</script>

<style lang="less" scoped></style>
