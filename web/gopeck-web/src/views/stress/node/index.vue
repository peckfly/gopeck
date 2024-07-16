<template>
    <div style="margin-top: 13px">
        <a-table
            :columns="columns"
            :dataSource="nodes"
            rowKey="addr">
            <template #bodyCell="{ column, record }">
                <template v-if="'actions' === column.key">
                    <a-button
                        type="primary"
                        style="background-color: peru"
                        @click="showNodeDetail(record)"
                        >{{ t('view_monitoring') }}</a-button
                    >
                    <a-button
                        type="primary"
                        @click="showUpdateQuotaModal(record)"
                        >{{ t('update_quota') }}</a-button
                    >
                </template>
                <template v-if="'rps_percentage' === column.key">
                    <a-progress
                        :percent="calculatePercentage(record.rps_cost, record.rps_quota)"
                        :status="getProgressStatus(record.rps_cost, record.rps_quota)" />
                </template>
                <template v-if="'goroutine_percentage' === column.key">
                    <a-progress
                        :percent="calculatePercentage(record.goroutine_cost, record.goroutine_quota)"
                        :status="getProgressStatus(record.goroutine_cost, record.goroutine_quota)" />
                </template>
            </template>
        </a-table>

        <a-modal
            v-model:open="detailModal.open"
            :title="detailModal.title"
            :width="1200"
            @cancel="handleCancel"
            :footer="null">
            <a-card :title="t('goroutine_num')">
                <XChart
                    :options="goroutineChartOptions"
                    :height="'350px'" />
            </a-card>
            <a-card :title="t('cpu_usage')">
                <XChart
                    :options="cpuChartOptions"
                    :height="'350px'" />
            </a-card>
            <a-card :title="t('memory_usage')">
                <XChart
                    :options="memChartOptions"
                    :height="'350px'" />
            </a-card>
            <a-card :title="t('loadavg_1_min')">
                <XChart
                    :options="load1ChartOptions"
                    :height="'350px'" />
            </a-card>
            <a-card :title="t('loadavg_5_min')">
                <XChart
                    :options="load5ChartOptions"
                    :height="'350px'" />
            </a-card>
            <a-card :title="t('loadavg_15_min')">
                <XChart
                    :options="load15ChartOptions"
                    :height="'350px'" />
            </a-card>
        </a-modal>

        <a-modal
            v-model:open="updateQuotaModal.open"
            :title="t('update_quota')"
            @ok="handleUpdateQuota"
            @cancel="handleCancel">
            <a-form
                :model="quotaForm"
                layout="vertical">
                <a-form-item :label="t('machine_address')">
                    <a-input
                        v-model:value="quotaForm.addr"
                        disabled />
                </a-form-item>
                <a-form-item :label="t('rps_quota')">
                    <a-input-number
                        v-model:value="quotaForm.rps_quota"
                        min="0"
                        style="width: 100%" />
                </a-form-item>
                <a-form-item :label="t('goroutine_quota')">
                    <a-input-number
                        v-model:value="quotaForm.goroutine_quota"
                        min="0"
                        style="width: 100%" />
                </a-form-item>
            </a-form>
        </a-modal>
    </div>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { message } from 'ant-design-vue'
import apis from '@/apis'

const { t } = useI18n()
console.log(`t`, t(''))
const nodes = ref([])

const columns = [
    { title: t('machine_address'), dataIndex: 'addr', key: 'addr', width: 200, fixed: 'left' },
    { title: t('rps_quota'), dataIndex: 'rps_quota', key: 'rps_quota' },
    { title: t('rps_cost'), dataIndex: 'rps_cost', key: 'rps_cost' },
    { title: t('rps_percentage'), dataIndex: 'rps_percentage', key: 'rps_percentage' },
    { title: t('goroutine_quota'), dataIndex: 'goroutine_quota', key: 'goroutine_quota' },
    { title: t('goroutine_cost'), dataIndex: 'goroutine_cost', key: 'goroutine_cost' },
    { title: t('goroutine_percentage'), dataIndex: 'goroutine_percentage', key: 'goroutine_percentage' },
    { title: t('running_task_count'), dataIndex: 'running_task_count', key: 'running_task_count' },
    { title: t('button.action'), key: 'actions', width: 260, fixed: 'right' },
]

const detailModal = ref({
    open: false,
    title: '',
})
const updateQuotaModal = ref({
    open: false,
    title: '',
})
const quotaForm = ref({
    addr: '',
    rps_quota: 0,
    goroutine_quota: 0,
})

const cpuChartOptions = ref({})
const memChartOptions = ref({})
const load1ChartOptions = ref({})
const load5ChartOptions = ref({})
const load15ChartOptions = ref({})
const goroutineChartOptions = ref({})

async function fetchNodes() {
    const { success, data } = await apis.stress.getNodeList().catch(() => {
        throw new Error(t('error.get_node_list_failed'))
    })
    if (success) {
        nodes.value = data.data
    } else {
        message.error(t('error.get_node_list_failed'))
    }
}

function showNodeDetail(record) {
    detailModal.value.open = true
    detailModal.value.title = `${t('machine_monitoring')} (${record.addr})`
    fetchNodeDetail(record.addr)
}

function showUpdateQuotaModal(record) {
    updateQuotaModal.value.open = true
    quotaForm.value = { addr: record.addr, rps_quota: record.rps_quota, goroutine_quota: record.goroutine_quota }
}

async function fetchNodeDetail(addr) {
    const { success, data } = await apis.stress.getNodeDetail({ addr: addr })
    if (success) {
        const detailData = data.data

        // 使用detailData中的Timestamp生成labels
        const labels = detailData.map((d) => new Date(d.Timestamp * 1000).toISOString())

        const cpuData = detailData.map((d) => d.ProgressInfo.CPUPercent)
        const memData = detailData.map((d) => d.ProgressInfo.MemPercent)
        const load1Data = detailData.map((d) => d.LoadInfo.load1)
        const load5Data = detailData.map((d) => d.LoadInfo.load5)
        const load15Data = detailData.map((d) => d.LoadInfo.load15)
        const goroutineData = detailData.map((d) => d.RunTimeInfo.GoRoutineNum)

        cpuChartOptions.value = {
            xAxis: {
                type: 'category',
                data: labels,
                axisLabel: {
                    formatter: function (value, index) {
                        if (index % 12 === 0) {
                            return new Date(value).toLocaleTimeString()
                        } else {
                            return ''
                        }
                    },
                },
            },
            yAxis: { type: 'value' },
            series: [{ data: cpuData, type: 'line', smooth: true, name: t('stress.nodeDetail.cpuUsage') }],
            tooltip: {
                trigger: 'axis',
                formatter: function (params) {
                    let date = new Date(params[0].name)
                    let formattedTime = date.toLocaleTimeString()
                    return `${params[0].seriesName}: ${params[0].value}<br/>${t('stress.nodeDetail.time')}: ${formattedTime}`
                },
            },
        }
        memChartOptions.value = {
            xAxis: {
                type: 'category',
                data: labels,
                axisLabel: {
                    formatter: function (value, index) {
                        if (index % 12 === 0) {
                            return new Date(value).toLocaleTimeString()
                        } else {
                            return ''
                        }
                    },
                },
            },
            yAxis: { type: 'value' },
            series: [{ data: memData, type: 'line', smooth: true, name: t('stress.nodeDetail.memUsage') }],
            tooltip: {
                trigger: 'axis',
                formatter: function (params) {
                    let date = new Date(params[0].name)
                    let formattedTime = date.toLocaleTimeString()
                    return `${params[0].seriesName}: ${params[0].value}<br/>${t('stress.nodeDetail.time')}: ${formattedTime}`
                },
            },
        }
        load1ChartOptions.value = {
            xAxis: {
                type: 'category',
                data: labels,
                axisLabel: {
                    formatter: function (value, index) {
                        if (index % 12 === 0) {
                            return new Date(value).toLocaleTimeString()
                        } else {
                            return ''
                        }
                    },
                },
            },
            yAxis: { type: 'value' },
            series: [{ data: load1Data, type: 'line', smooth: true, name: t('stress.nodeDetail.load1') }],
            tooltip: {
                trigger: 'axis',
                formatter: function (params) {
                    let date = new Date(params[0].name)
                    let formattedTime = date.toLocaleTimeString()
                    return `${params[0].seriesName}: ${params[0].value}<br/>${t('stress.nodeDetail.time')}: ${formattedTime}`
                },
            },
        }
        load5ChartOptions.value = {
            xAxis: {
                type: 'category',
                data: labels,
                axisLabel: {
                    formatter: function (value, index) {
                        if (index % 12 === 0) {
                            return new Date(value).toLocaleTimeString()
                        } else {
                            return ''
                        }
                    },
                },
            },
            yAxis: { type: 'value' },
            series: [{ data: load5Data, type: 'line', smooth: true, name: t('stress.nodeDetail.load5') }],
            tooltip: {
                trigger: 'axis',
                formatter: function (params) {
                    let date = new Date(params[0].name)
                    let formattedTime = date.toLocaleTimeString()
                    return `${params[0].seriesName}: ${params[0].value}<br/>${t('stress.nodeDetail.time')}: ${formattedTime}`
                },
            },
        }
        load15ChartOptions.value = {
            xAxis: {
                type: 'category',
                data: labels,
                axisLabel: {
                    formatter: function (value, index) {
                        if (index % 12 === 0) {
                            return new Date(value).toLocaleTimeString()
                        } else {
                            return ''
                        }
                    },
                },
            },
            yAxis: { type: 'value' },
            series: [{ data: load15Data, type: 'line', smooth: true, name: t('stress.nodeDetail.load15') }],
            tooltip: {
                trigger: 'axis',
                formatter: function (params) {
                    let date = new Date(params[0].name)
                    let formattedTime = date.toLocaleTimeString()
                    return `${params[0].seriesName}: ${params[0].value}<br/>${t('stress.nodeDetail.time')}: ${formattedTime}`
                },
            },
        }
        goroutineChartOptions.value = {
            xAxis: {
                type: 'category',
                data: labels,
                axisLabel: {
                    formatter: function (value, index) {
                        if (index % 12 === 0) {
                            return new Date(value).toLocaleTimeString()
                        } else {
                            return ''
                        }
                    },
                },
            },
            yAxis: { type: 'value' },
            series: [{ data: goroutineData, type: 'line', smooth: true, name: t('stress.nodeDetail.goroutineNum') }],
            tooltip: {
                trigger: 'axis',
                formatter: function (params) {
                    let date = new Date(params[0].name)
                    let formattedTime = date.toLocaleTimeString()
                    return `${params[0].seriesName}: ${params[0].value}<br/>${t('stress.nodeDetail.time')}: ${formattedTime}`
                },
            },
        }
    }
}

function handleCancel() {
    detailModal.value.open = false
    updateQuotaModal.value.open = false
}

function calculatePercentage(cost, quota) {
    return quota === 0 ? 0 : ((cost / quota) * 100).toFixed(1)
}

function getProgressStatus(cost, quota) {
    return cost > quota ? 'exception' : 'normal'
}

async function handleUpdateQuota() {
    const { success } = await apis.stress.updateNodeQuota(quotaForm.value)
    if (success) {
        message.success(t('stress.nodeDetail.quotaUpdateSuccess'))
        fetchNodes()
        updateQuotaModal.value.open = false
    } else {
        message.error(t('stress.nodeDetail.quotaUpdateFailed'))
    }
}

fetchNodes()
</script>
