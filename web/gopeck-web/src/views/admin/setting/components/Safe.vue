<template>
    <a-row
        :gutter="48"
        type="flex">
        <a-col flex="0 0 480px">
            <a-form
                layout="vertical"
                :model="form"
                @finish="handleUpdatePassword">
                <a-form-item
                    :label="t('pages.user.profile.tab.security.form.old_password')"
                    name="oldPassword">
                    <a-input-password
                        v-model:value="form.oldPassword"
                        :placeholder="$t('pages.user.profile.tab.security.form.old_password.placeholder')">
                    </a-input-password>
                </a-form-item>
                <a-form-item
                    :label="t('pages.user.profile.tab.security.form.password')"
                    name="newPassword">
                    <a-input-password
                        v-model:value="form.newPassword"
                        :placeholder="$t('pages.user.profile.tab.security.form.password.placeholder')">
                    </a-input-password>
                </a-form-item>
                <a-form-item
                    :label="t('pages.user.profile.tab.security.form.confirm_password')"
                    name="confirmPassword">
                    <a-input-password
                        v-model:value="form.confirmPassword"
                        :placeholder="$t('pages.user.profile.tab.security.form.confirm_password.placeholder')">
                    </a-input-password>
                </a-form-item>
                <a-form-item>
                    <a-button
                        type="primary"
                        html-type="submit"
                        >{{ t('pages.user.profile.tab.security.form.update_password') }}</a-button
                    >
                </a-form-item>
            </a-form>
        </a-col>
    </a-row>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import apis from '@/apis'
import { message } from 'ant-design-vue'

defineOptions({
    name: 'Safe',
})
const { t } = useI18n()

const form = ref({
    oldPassword: '',
    newPassword: '',
    confirmPassword: '',
})

const handleUpdatePassword = async () => {
    try {
        const params = {
            old_password: form.value.oldPassword,
            new_password: form.value.newPassword,
            confirm_password: form.value.confirmPassword,
        }
        const { success } = await apis.user.updatePassword(null, params)
        if (success) {
            message.success(t('pages.user.profile.tab.security.form.password.update_success'))
        }
    } catch (error) {
        message.error(t('pages.user.profile.tab.security.form.password.update_failed'))
    }
}
</script>

<style lang="less" scoped></style>
