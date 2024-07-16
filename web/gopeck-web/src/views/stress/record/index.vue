<template>
    <x-search-bar class="mb-8-2">
        <template #default="{ gutter, colSpan }">
            <a-form
                :label-col="{ style: { width: '100px' } }"
                :model="searchFormData"
                layout="inline">
                <a-row :gutter="gutter">
                    <a-col v-bind="colSpan">
                        <a-form-item
                            :label="$t('stress.plan.record.page.plan_id')"
                            name="username">
                            <a-input
                                :placeholder="$t('stress.plan.record.page.plan_id_search_placeholder')"
                                v-model:value="searchFormData.plan_id"></a-input>
                        </a-form-item>
                    </a-col>

                    <a-col v-bind="colSpan">
                        <a-form-item
                            name="name"
                            :label-col="{ span: 10 }"
                            :wrapper-col="{ span: 120 }">
                            <template #label>
                                {{ $t('stress.plan.record.page.plan_name') }}
                                <a-tooltip :title="$t('stress.plan.record.page.plan_name')">
                                    <question-circle-outlined class="ml-4-1 color-placeholder" />
                                </a-tooltip>
                            </template>
                            <a-input
                                :placeholder="$t('stress.plan.record.page.plan_name_search_placeholder')"
                                v-model:value="searchFormData.plan_name"></a-input>
                        </a-form-item>
                    </a-col>

                    <a-col
                        class="align-right"
                        v-bind="colSpan">
                        <a-space>
                            <a-button @click="handleResetSearch">{{ $t('button.reset') }}</a-button>
                            <a-button
                                ghost
                                type="primary"
                                @click="handleSearch">
                                {{ $t('button.search') }}
                            </a-button>
                        </a-space>
                    </a-col>
                </a-row>
            </a-form>
        </template>
    </x-search-bar>
    <a-row
        :gutter="8"
        :wrap="false">
        <a-col flex="auto">
            <a-card type="flex">
                <a-table
                    :columns="columns"
                    :data-source="listData"
                    :loading="loading"
                    :pagination="paginationState"
                    :scroll="{ x: 1000 }"
                    @change="onTableChange">
                    <template #bodyCell="{ column, record }">
                        <template v-if="'stressType' === column.key">
                            <a-tag
                                v-if="record.stress_type === 1"
                                color="success">
                                {{ $t('stress.rpsModeRecord') }}
                            </a-tag>
                            <a-tag
                                v-if="record.stress_type === 2"
                                color="processing">
                                {{ $t('stress.currencyMode') }}
                            </a-tag>
                        </template>

                        <template v-if="'stressMode' === column.key">
                            <a-tag
                                v-if="record.stress_mode === 1"
                                color="processing">
                                {{ $t('stress.normalMode') }}
                            </a-tag>
                            <a-tag
                                v-if="record.stress_mode === 2"
                                color="success">
                                {{ $t('stress.stepMode') }}
                            </a-tag>
                        </template>

                        <template v-if="'stress_progress' === column.key">
                            <a-progress
                                :percent="record.stress_progress"
                                type="circle"
                                :size="45" />
                        </template>

                        <template v-if="'create_time' === column.key">
                            {{ formatTimestampTime(record.create_time) }}
                        </template>

                        <template v-if="'overview_metrics_url' === column.key">
                            <a
                                :href="record.overview_metrics_url"
                                target="_blank">
                                <LinkOutlined />
                                {{ t('stress.plan.record.page.view_metrics') }}
                            </a>
                        </template>

                        <template v-if="'action' === column.key">
                            <a-button
                                type="primary"
                                @click="$refs.showReportRef.showReport(record)">
                                <FileDoneOutlined />
                                {{ t('stress.plan.record.page.task_record_report') }}
                            </a-button>
                            <a-button
                                type="primary"
                                v-if="record.stress_progress < 100"
                                danger
                                @click="stopStress(record.plan_id)">
                                <CloseOutlined />
                                {{ t('stress.plan.record.page.stop_stress') }}
                            </a-button>

                            <a-button
                                v-if="record.stress_progress >= 100"
                                type="primary"
                                style="background-color: #f56a00"
                                @click="copyStress(record.plan_id)">
                                <ExportOutlined />
                                {{ t('stress.plan.record.page.copy_stress') }}
                            </a-button>
                        </template>
                    </template>
                </a-table>
            </a-card>
        </a-col>
    </a-row>

    <report-dialog
        ref="showReportRef"
        @ok="onOk"></report-dialog>
</template>

<script setup>
import { message } from 'ant-design-vue'
import { ref } from 'vue'
import apis from '@/apis'
import { config } from '@/config'
import { usePagination } from '@/hooks'

import { formatTimestampTime } from '@/utils/util'

import ReportDialog from './components/ReportDialog.vue'
import { useI18n } from 'vue-i18n'
import { LinkOutlined, ExportOutlined, FileDoneOutlined, CloseOutlined } from '@ant-design/icons-vue'
import { useRouter } from 'vue-router'

defineOptions({
    name: 'planRecord',
})
const router = useRouter()

const { t } = useI18n() // 解构出t方法
const columns = [
    { title: t('stress.plan.record.page.plan_id'), dataIndex: 'plan_id', width: 145, fixed: 'left' },
    { title: t('stress.plan.record.page.plan_name'), dataIndex: 'plan_name', key: 'name', width: 130, fixed: 'left' },
    { title: t('stress.plan.record.page.stress_type'), dataIndex: 'stress_type', key: 'stressType', width: 90 },
    { title: t('stress.plan.record.page.stress_mode'), dataIndex: 'stress_mode', key: 'stressMode', width: 90 },
    { title: t('stress.plan.record.page.stress_time'), dataIndex: 'stress_time', width: 60 },
    {
        title: t('stress.plan.record.page.stress_progress'),
        dataIndex: 'stress_progress',
        key: 'stress_progress',
        width: 60,
    },
    {
        title: t('stress.plan.record.page.overview_metrics_url'),
        dataIndex: 'overview_metrics_url',
        key: 'overview_metrics_url',
        width: 80,
    },
    { title: t('stress.plan.record.page.create_time'), dataIndex: 'create_time', key: 'create_time', width: 160 },
    { title: t('button.action'), key: 'action', width: 160, fixed: 'right' },
]

const { listData, loading, showLoading, hideLoading, paginationState, resetPagination, searchFormData } =
    usePagination()

const showReportRef = ref()

getPlanList()

function copyStress(planId) {
    router.push({ name: 'start', query: { plan_id: planId } })
}

async function getPlanList() {
    try {
        showLoading()
        const { pageSize, current } = paginationState
        const { success, data, total } = await apis.stress
            .getPlanList({
                pageSize,
                page: current,
                ...searchFormData.value,
            })
            .catch(() => {
                throw new Error()
            })
        hideLoading()
        if (config('http.code.success') === success) {
            listData.value = data
            paginationState.total = total
        }
    } catch (error) {
        hideLoading()
    }
}

async function stopStress(planId) {
    const { success } = await apis.stress
        .stopStress({
            plan_id: planId,
        })
        .catch(() => {
            throw new Error()
        })
    if (true === success) {
        message.success(t('stress.plan.record.page.stop_stress_success'))
    }
    getPlanList()
}

/**
 * 分页
 */
function onTableChange({ current, pageSize }) {
    paginationState.current = current
    paginationState.pageSize = pageSize
    getPlanList()
}

/**
 * 搜索
 */
function handleSearch() {
    resetPagination()
    getPlanList()
}

/**
 * 重置
 */
function handleResetSearch() {
    searchFormData.value = {}
    resetPagination()
    getPlanList()
}

/**
 * 编辑完成
 */
async function onOk() {
    message.success(t('component.message.success.delete'))
    await getPlanList()
}
</script>

<style lang="less" scoped>
.show-report-button {
    background-color: #add8e6;
}

.stop-stress-button {
    background-color: #dd4a68;
}
</style>
