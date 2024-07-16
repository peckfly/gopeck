<template>
    <a-modal
        :open="modal.open"
        :title="modal.title"
        :width="1200"
        :confirm-loading="modal.confirmLoading"
        :after-close="onAfterClose"
        :cancel-text="cancelText"
        :ok-text="okText"
        @ok="handleOk"
        @cancel="handleCancel">
        <a-collapse v-model:activeKey="activeTabKey">
            <a-collapse-panel
                v-for="(result, index) in stressResults"
                :key="index"
                :header="`${$t('stress.task.requestTitle')} ${index + 1}: ${result.task_name}`">
                <a-collapse :defaultActiveKey="['report']">
                    <a-collapse-panel
                        key="details"
                        :header="$t('stress.task.report_page.report_detail_params')">
                        <a-tabs default-active-key="1">
                            <a-tab-pane
                                :key="1"
                                :tab="$t('stress.task.report.basicParam')">
                                <a-row :gutter="12">
                                    <a-col :span="24">
                                        <a-table
                                            :show-header="false"
                                            :columns="descriptionColumns"
                                            :dataSource="getDescriptionData(result)"
                                            :pagination="false"
                                            bordered />
                                    </a-col>
                                </a-row>
                            </a-tab-pane>

                            <a-tab-pane
                                :key="2"
                                tab="Param">
                                <a-row
                                    :gutter="12"
                                    class="mt-4">
                                    <a-col :span="24">
                                        <a-table
                                            :pagination="false"
                                            :columns="paramColumns"
                                            :dataSource="result.query"
                                            bordered />
                                    </a-col>
                                </a-row>
                            </a-tab-pane>

                            <!-- Headers Section -->
                            <a-tab-pane
                                :key="3"
                                tab="Header">
                                <a-row
                                    :gutter="12"
                                    class="mt-4">
                                    <a-col :span="24">
                                        <a-table
                                            :pagination="false"
                                            :columns="headerColumns"
                                            :dataSource="result.header"
                                            bordered />
                                    </a-col>
                                </a-row>
                            </a-tab-pane>

                            <a-tab-pane
                                :key="4"
                                tab="Body">
                                <!-- Body Section -->
                                <a-card
                                    title="Body"
                                    bordered>
                                    <vue-json-editor
                                        style="height: 300px"
                                        :mainMenuBar="false"
                                        mode="text"
                                        :content="result.body"
                                        :readOnly="true"
                                        :parser="LosslessJSONParser" />
                                </a-card>
                            </a-tab-pane>
                            <a-tab-pane
                                :key="5"
                                tab="Scripts">
                                <!-- Scripts Section -->
                                <a-row
                                    :gutter="12"
                                    class="mt-4">
                                    <a-col :span="24">
                                        <a-tabs default-active-key="11">
                                            <a-tab-pane
                                                :key="11"
                                                :tab="$t('stress.task.dynamicParamScript')">
                                                <Codemirror
                                                    v-model:value="result.dynamic_param_script"
                                                    :options="cmOptions"
                                                    border
                                                    height="300" />
                                            </a-tab-pane>
                                            <a-tab-pane
                                                :key="12"
                                                :tab="$t('stress.task.responseCheckScript')">
                                                <Codemirror
                                                    v-model:value="result.response_check_script"
                                                    :options="cmOptions"
                                                    border
                                                    height="300" />
                                            </a-tab-pane>
                                        </a-tabs>
                                    </a-col>
                                </a-row>
                            </a-tab-pane>
                        </a-tabs>
                    </a-collapse-panel>

                    <!-- Report Table Section -->
                    <a-collapse-panel
                        key="report"
                        :header="$t('stress.task.report_page.report_detail_table')">
                        <a-tabs default-active-key="101">
                            <a-tab-pane
                                :key="101"
                                :tab="$t('stress.task.time_cost_statistics')">
                                <a-button
                                    type="link"
                                    :href="result.metrics_url"
                                    target="_blank">
                                    <LinkOutlined />
                                    Grafana Metrics
                                </a-button>
                                <a-row
                                    :gutter="12"
                                    class="mt-4">
                                    <a-col :span="24">
                                        <a-table
                                            :pagination="false"
                                            :columns="getReportColumns(result.stress_type)"
                                            :dataSource="result.reports"
                                            rowKey="TaskId" />
                                    </a-col>
                                </a-row>
                            </a-tab-pane>
                            <!-- Histogram Section -->
                            <a-tab-pane
                                :key="102"
                                :tab="$t('stress.task.time_histogram')">
                                <a-row
                                    v-for="(report, reportIndex) in result.reports"
                                    :key="reportIndex"
                                    :gutter="15"
                                    class="mt-4">
                                    <a-col :span="24">
                                        <a-card
                                            :title="
                                                result.stress_type === 1
                                                    ? `${$t('stress.task.rps')} ${report.Num}`
                                                    : `${$t('stress.task.concurrencyNum')} ${report.Num}`
                                            ">
                                            <XChart
                                                :options="getHistogramOptions(report.Histogram)"
                                                :width="'950%'"
                                                :height="'350px'"></XChart>
                                        </a-card>
                                    </a-col>
                                </a-row>
                            </a-tab-pane>

                            <a-tab-pane
                                :key="103"
                                :tab="$t('stress.task.http_code_distribution')">
                                <a-row
                                    v-for="(report, reportIndex) in result.reports"
                                    :key="reportIndex"
                                    :gutter="15"
                                    class="mt-4">
                                    <a-col :span="24">
                                        <a-card
                                            :title="
                                                result.stress_type === 1
                                                    ? `${$t('stress.task.rps')} ${report.Num}`
                                                    : `${$t('stress.task.concurrencyNum')} ${report.Num}`
                                            ">
                                            <a-space
                                                size="middle"
                                                style="display: flex; justify-content: center; align-items: center">
                                                <XChart
                                                    :options="
                                                        getCodeDistributionOptions(
                                                            $t('stress.task.http_code_distribution'),
                                                            report.StatusCodeDist
                                                        )
                                                    "
                                                    :width="'300%'"
                                                    :height="'250px'" />
                                            </a-space>
                                        </a-card>
                                    </a-col>
                                </a-row>
                            </a-tab-pane>

                            <a-tab-pane
                                :key="104"
                                :tab="$t('stress.task.error_distribution')">
                                <a-row
                                    v-for="(report, reportIndex) in result.reports"
                                    :key="reportIndex"
                                    :gutter="15"
                                    class="mt-4">
                                    <a-col :span="24">
                                        <a-card
                                            :title="
                                                result.stress_type === 1
                                                    ? `${$t('stress.task.rps')} ${report.Num}`
                                                    : `${$t('stress.task.concurrencyNum')} ${report.Num}`
                                            ">
                                            <a-space
                                                v-if="
                                                    report.ErrorDist !== null &&
                                                    Object.keys(report.ErrorDist).length > 0
                                                "
                                                size="middle"
                                                style="display: flex; justify-content: center; align-items: center">
                                                <XChart
                                                    :options="
                                                        getCodeDistributionOptions(
                                                            $t('stress.task.error_distribution'),
                                                            report.ErrorDist
                                                        )
                                                    "
                                                    :width="'300%'"
                                                    :height="'250px'" />
                                            </a-space>
                                            <template v-else>
                                                <div>{{ $t('stress.task.noError') }}</div>
                                            </template>
                                        </a-card>
                                    </a-col>
                                </a-row>
                            </a-tab-pane>

                            <a-tab-pane
                                :key="105"
                                :tab="$t('stress.task.result_check_distribution')">
                                <a-row
                                    v-for="(report, reportIndex) in result.reports"
                                    :key="reportIndex"
                                    :gutter="15"
                                    class="mt-4">
                                    <a-col :span="24">
                                        <a-card
                                            :title="
                                                result.stress_type === 1
                                                    ? `${$t('stress.task.rps')} ${report.Num}`
                                                    : `${$t('stress.task.concurrencyNum')} ${report.Num}`
                                            ">
                                            <a-space
                                                v-if="
                                                    report.BodyCheckResultMap !== null &&
                                                    Object.keys(report.BodyCheckResultMap).length > 0
                                                "
                                                size="middle"
                                                style="display: flex; justify-content: center; align-items: center">
                                                <XChart
                                                    :options="
                                                        getCodeDistributionOptions(
                                                            $t('stress.task.result_check_distribution'),
                                                            report.BodyCheckResultMap
                                                        )
                                                    "
                                                    :width="'300%'"
                                                    :height="'250px'" />
                                            </a-space>
                                            <template v-else>
                                                <div>{{ $t('stress.task.noData') }}</div>
                                            </template>
                                        </a-card>
                                    </a-col>
                                </a-row>
                            </a-tab-pane>
                        </a-tabs>
                    </a-collapse-panel>
                </a-collapse>
            </a-collapse-panel>
        </a-collapse>
    </a-modal>
</template>
<script setup>
import { ref } from 'vue'
import { useModal } from '@/hooks'
import { useI18n } from 'vue-i18n'
import apis from '@/apis'
import { message } from 'ant-design-vue'
import Codemirror from 'codemirror-editor-vue3'
import 'codemirror/mode/go/go.js'

import { parse, stringify } from 'lossless-json'
import { LinkOutlined } from '@ant-design/icons-vue'

const LosslessJSONParser = { parse, stringify }

const { t } = useI18n()
const { modal, showModal, hideModal, showLoading, hideLoading } = useModal()
const activeTabKey = ref([])
const stressResults = ref([])

const cancelText = ref(t('button.cancel'))
const okText = ref(t('button.confirm'))

const descriptionColumns = [
    { title: 'Property', dataIndex: 'property', key: 'property' },
    { title: 'Value', dataIndex: 'value', key: 'value' },
]

function getDescriptionData(result) {
    return [
        { property: t('stress.task.report_page.task_id'), value: result.task_id },
        { property: t('stress.task.taskName'), value: result.task_name },
        { property: t('stress.task.url'), value: result.url },
        { property: 'METHOD', value: result.method },
        { property: concurrencyOrRpsLabelText(result.stress_type, result.stress_mode), value: result.num },
        {
            property: maxConcurrencyOrRpsLabelText(result.stress_type),
            value: result.stress_mode === 2 ? result.max_num : null,
        },
        {
            property: stepConcurrencyOrRpsLabelText(result.stress_type),
            value: result.stress_mode === 2 ? result.step_num : null,
        },
        { property: t('stress.task.requestTimeout'), value: result.timeout !== 0 ? result.timeout : null },
        {
            property: t('stress.task.disableCompression'),
            value: result.disable_compression === 1 ? t('stress.task.YES') : null,
        },
        {
            property: t('stress.task.disableKeepAlive'),
            value: result.disable_keep_alive === 1 ? t('stress.task.YES') : null,
        },
        {
            property: t('stress.task.disableRedirect'),
            value: result.disable_redirects === 1 ? t('stress.task.YES') : null,
        },
        {
            property: t('stress.task.enableHttp2'),
            value: result.h_2 === 1 ? t('stress.task.YES') : null,
        },
    ].filter((item) => item.value !== null)
}

const getHistogramOptions = (histogram) => {
    const histData = histogram
    return {
        xAxis: {
            type: 'category',
            name: t('stress.task.histogram.XAxisName'),
            data: histData.map((data) => data.Mark),
            axisLabel: {
                formatter: '{value} ms',
            },
        },
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'shadow',
            },
        },
        yAxis: {
            name: t('stress.task.histogram.YAxisName'),
            type: 'value',
        },
        series: [
            {
                data: histData.map((data) => data.Count),
                type: 'bar',
            },
        ],
    }
}

const getCodeDistributionOptions = (name, codeDist) => {
    const data = Object.entries(codeDist).map(([key, value]) => ({ name: key, value }))
    return {
        tooltip: {
            trigger: 'item',
            formatter: '{a} <br/>{b}: {c} ({d}%)',
        },
        series: [
            {
                name: name,
                type: 'pie',
                radius: '50%',
                data,
                emphasis: {
                    itemStyle: {
                        shadowBlur: 10,
                        shadowOffsetX: 0,
                        shadowColor: 'rgba(0, 0, 0, 0.5)',
                    },
                },
            },
        ],
    }
}

const paramColumns = ref([
    {
        title: 'Param Key',
        dataIndex: 'entry_key',
        key: 'entry_key',
    },
    {
        title: 'Param Value',
        dataIndex: 'entry_value',
        key: 'entry_value',
    },
])

const headerColumns = ref([
    {
        title: 'Header Key',
        dataIndex: 'entry_key',
        key: 'entry_key',
    },
    {
        title: 'Header Value',
        dataIndex: 'entry_value',
        key: 'entry_value',
    },
])
const cmOptions = {
    mode: 'text/x-go',
    tabSize: 4,
    readOnly: true,
}

function concurrencyOrRpsLabelText(stressType, stressMode) {
    if (stressMode === 2) {
        return stressType === 1 ? t('stress.task.startRps') : t('stress.task.startConcurrencyNum')
    }
    return stressType === 1 ? t('stress.task.rps') : t('stress.task.concurrencyNum')
}

function reportStepConcurrencyOrRpsLabelText(stressType) {
    return stressType === 1 ? t('stress.task.rps') : t('stress.task.concurrencyNum')
}

function maxConcurrencyOrRpsLabelText(stressType) {
    return stressType === 1 ? t('stress.task.maxRps') : t('stress.task.maxConcurrencyNum')
}

function stepConcurrencyOrRpsLabelText(stressType) {
    return stressType === 1 ? t('stress.task.stepRps') : t('stress.task.stepConcurrencyNum')
}

function getReportColumns(stressType) {
    return [
        { title: reportStepConcurrencyOrRpsLabelText(stressType), dataIndex: 'Num', key: 'Num' },
        { title: t('stress.task.record.Rps'), dataIndex: 'Rps', key: 'Rps' },
        { title: t('stress.task.record.totalNumRes'), dataIndex: 'NumRes', key: 'NumRes' },
        { title: t('stress.task.record.totalCostTime'), dataIndex: 'TotalCostTime', key: 'TotalCostTime' },
        { title: 'Max(ms)', dataIndex: 'Slowest', key: 'Slowest' },
        { title: 'Min(ms)', dataIndex: 'Fastest', key: 'Fastest' },
        { title: 'Avg(ms)', dataIndex: 'Average', key: 'Average' },
        { title: '90%(ms)', dataIndex: 'lat_90', key: 'lat_90' },
        { title: '95%(ms)', dataIndex: 'lat_95', key: 'lat_95' },
        { title: '99%(ms)', dataIndex: 'lat_99', key: 'lat_99' },
        { title: '99.9%(ms)', dataIndex: 'lat_999', key: 'lat_999' },
        { title: t('stress.task.record.errorNum'), dataIndex: 'ErrorCount', key: 'ErrorCount' },
        { title: t('stress.task.record.errorRate'), dataIndex: 'ErrorRate', key: 'ErrorRate' },
    ]
}

function showReport(record) {
    showModal({
        type: 'report',
        title: t('stress.plan.record.page.task_record_report'),
    })
    fetchStressResults(record.plan_id)
}

async function fetchStressResults(planId) {
    try {
        showLoading()
        const { success, data } = await apis.stress.getTaskList({
            plan_id: planId,
        })
        hideLoading()
        if (success) {
            stressResults.value = data.data
        } else {
            message.error(t('stress.report.message.error.fetchFailed'))
        }
    } catch (error) {
        hideLoading()
        message.error(t('stress.report.message.error.fetchFailed'))
    }
}

function handleOk() {
    hideModal()
}

function handleCancel() {
    hideModal()
}

function onAfterClose() {
    activeTabKey.value = []
    stressResults.value = []
}

defineExpose({
    showReport,
})
</script>

<style scoped lang="less">
.mt-4 {
    margin-top: 16px;
}
</style>
