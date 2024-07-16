<template>
    <a-form :model="stressForm">
        <div class="form-header">
            <a-form-item :label="$t('stress.planName')">
                <a-input
                    v-model:value="stressForm.plan_name"
                    style="width: 250px" />
            </a-form-item>

            <div class="button-group">
                <a-button
                    type="primary"
                    @click="startStress">
                    <RightSquareOutlined />
                    {{ t('stress.task.startStress') }}
                </a-button>

                <a-button
                    type="default"
                    class="reset-button"
                    @click="resetForm">
                    <ReloadOutlined />
                    {{ t('stress.task.resetParameters') }}
                </a-button>
            </div>
        </div>

        <a-form-item :label="$t('stress.time')">
            <a-input-number
                v-model:value="stressForm.stress_time"
                :placeholder="$t('stress.stressTimePlaceholder')"
                style="width: 250px" />
        </a-form-item>
        <a-form-item :label="$t('stress.stressType')">
            <a-select
                ref="select"
                v-model:value.number="stressForm.stress_type"
                style="width: 250px"
                @focus="focus">
                <a-select-option :value="1">{{ $t('stress.rpsMode') }}</a-select-option>
                <a-select-option :value="2">{{ $t('stress.currencyMode') }}</a-select-option>
            </a-select>
        </a-form-item>

        <a-form-item :label="$t('stress.stressMode')">
            <a-select
                ref="select"
                v-model:value.number="stressForm.stress_mode"
                style="width: 250px"
                @focus="focus">
                <a-select-option :value="1">{{ $t('stress.normalMode') }}</a-select-option>
                <a-select-option :value="2">{{ $t('stress.stepMode') }}</a-select-option>
            </a-select>
        </a-form-item>

        <a-form-item
            :label="$t('stress.stepIntervalTime')"
            v-if="stressForm.stress_mode === 2">
            <a-input-number
                v-model:value="stressForm.step_interval_time"
                :placeholder="$t('stress.stepIntervalTimePlaceholder')"
                style="width: 250px" />
        </a-form-item>

        <a-form-item>
            <a-collapse v-model:activeKey="activeTabKey">
                <a-collapse-panel
                    v-for="(item, index) in stressForm.tasks"
                    :key="index">
                    <template #header>
                        <div class="panel-header">
                            <span
                                >{{ $t('stress.task.requestTitle') }} {{ index + 1 }}{{ item.task_name ? ': ' : '' }}
                                <b style="font-weight: bold">{{ item.task_name }}</b></span
                            >
                            <div>
                                <a-button
                                    type="primary"
                                    @click="duplicateItem(index)"
                                    class="duplicate-button">
                                    <CopyOutlined />
                                </a-button>
                                <a-button
                                    type="primary"
                                    @click="removeItem(index)"
                                    class="delete-button"
                                    danger>
                                    <DeleteOutlined />
                                </a-button>
                            </div>
                        </div>
                    </template>
                    <a-form
                        :form="formRefs[index].value"
                        v-bind="formItemLayout"
                        :initialValues="item">
                        <a-tabs default-active-key="1">
                            <a-tab-pane
                                key="1"
                                :tab="$t('stress.task.tabBaseParam')">
                                <a-form-item
                                    :label="$t('stress.task.taskName')"
                                    name="taskName"
                                    rules="required">
                                    <a-input v-model:value="item.task_name" />
                                </a-form-item>
                                <a-form-item
                                    :label="$t('stress.task.url')"
                                    name="url"
                                    rules="required">
                                    <a-input v-model:value="item.url" />
                                </a-form-item>
                                <a-form-item
                                    :label="concurrencyOrRpsLabelText"
                                    name="num"
                                    rules="required">
                                    <a-input-number
                                        v-model:value="item.num"
                                        style="width: 150px" />
                                </a-form-item>
                                <a-form-item
                                    :label="maxConcurrencyOrRpsLabelText"
                                    name="max_num"
                                    rules="required"
                                    v-if="stressForm.stress_mode === 2">
                                    <a-input-number
                                        v-model:value="item.max_num"
                                        style="width: 150px" />
                                </a-form-item>
                                <a-form-item
                                    :label="stepConcurrencyOrRpsLabelText"
                                    name="step_num"
                                    rules="required"
                                    v-if="stressForm.stress_mode === 2">
                                    <a-input-number
                                        v-model:value="item.step_num"
                                        style="width: 150px" />
                                </a-form-item>
                                <a-form-item
                                    label="METHOD"
                                    name="method"
                                    rules="required">
                                    <a-select
                                        ref="select"
                                        v-model:value="item.method"
                                        style="width: 150px"
                                        @focus="focus">
                                        <a-select-option value="GET">GET</a-select-option>
                                        <a-select-option value="POST">POST</a-select-option>
                                        <a-select-option value="PUT">PUT</a-select-option>
                                        <a-select-option value="DELETE">DELETE</a-select-option>
                                    </a-select>
                                </a-form-item>
                            </a-tab-pane>

                            <a-tab-pane
                                key="2"
                                tab="Param">
                                <a-form-item
                                    label="Param"
                                    name="param">
                                    <x-form-table
                                        v-model="item.query"
                                        :row-tpl="{ entry_key: '', entry_value: '' }"
                                        bordered>
                                        <a-table-column
                                            data-index="queryKey"
                                            title="key">
                                            <template #default="{ record }">
                                                <a-input
                                                    v-model:value="record.entry_key"
                                                    :key="index" />
                                            </template>
                                        </a-table-column>
                                        <a-table-column
                                            data-index="queryValue"
                                            title="value">
                                            <template #default="{ record }">
                                                <a-input
                                                    v-model:value="record.entry_value"
                                                    :key="index" />
                                            </template>
                                        </a-table-column>
                                    </x-form-table>
                                </a-form-item>
                            </a-tab-pane>

                            <a-tab-pane
                                key="3"
                                tab="Header">
                                <a-form-item
                                    label="Header"
                                    name="header">
                                    <x-form-table
                                        v-model="item.header"
                                        :row-tpl="{ entry_key: '', entry_value: '' }"
                                        bordered>
                                        <a-table-column
                                            data-index="headerKey"
                                            title="key">
                                            <template #default="{ record }">
                                                <a-input v-model:value="record.entry_key" />
                                            </template>
                                        </a-table-column>
                                        <a-table-column
                                            data-index="headerValue"
                                            title="value">
                                            <template #default="{ record }">
                                                <a-input v-model:value="record.entry_value" />
                                            </template>
                                        </a-table-column>
                                    </x-form-table>
                                </a-form-item>
                            </a-tab-pane>
                            <a-tab-pane
                                key="4"
                                tab="Body">
                                <a-form-item
                                    label="Body"
                                    name="body">
                                    <vue-json-editor
                                        style="height: 300px"
                                        :mainMenuBar="false"
                                        mode="text"
                                        :content="item.bodyJson"
                                        :onChange="(content) => onBodyChange(item, content)"
                                        :parser="LosslessJSONParser" />
                                </a-form-item>
                            </a-tab-pane>
                            <a-tab-pane
                                key="5"
                                :tab="$t('stress.task.dynamicParamScript')">
                                <a-tooltip
                                    placement="top"
                                    color="white">
                                    <template #title>
                                        <div>
                                            <Codemirror
                                                v-model:value="defaultDynamicScript"
                                                :options="cmOptions"
                                                ref="defaultDynamicScriptRef"
                                                height="400"
                                                width="800" />
                                        </div>
                                    </template>
                                    <a-button
                                        shape="circle"
                                        icon="?" />
                                </a-tooltip>
                                <a-form-item
                                    :label="$t('stress.task.dynamicParamScript')"
                                    name="DynamicParamScript">
                                    <Codemirror
                                        v-model:value="item.dynamic_param_script"
                                        :options="cmOptions"
                                        border
                                        ref="cmRef"
                                        height="300"></Codemirror>
                                </a-form-item>
                            </a-tab-pane>
                            <a-tab-pane
                                key="6"
                                :tab="$t('stress.task.responseCheckScript')">
                                <a-tooltip
                                    placement="top"
                                    color="white">
                                    <template #title>
                                        <div>
                                            <Codemirror
                                                v-model:value="defaultResponseCheckScript"
                                                :options="cmOptions"
                                                height="400"
                                                ref="defaultResponseCheckScriptRef"
                                                width="800" />
                                        </div>
                                    </template>
                                    <a-button
                                        shape="circle"
                                        icon="?" />
                                </a-tooltip>
                                <a-form-item
                                    :label="$t('stress.task.responseCheckScript')"
                                    name="DynamicParamScript">
                                    <Codemirror
                                        v-model:value="item.response_check_script"
                                        :options="cmOptions"
                                        border
                                        ref="cmRsRef"
                                        height="300"></Codemirror>
                                </a-form-item>
                            </a-tab-pane>

                            <a-tab-pane
                                key="7"
                                :tab="$t('stress.task.otherOptions')">
                                <a-form-item
                                    :label="t('stress.task.maxConnections')"
                                    name="max_connections"
                                    rules="required">
                                    <a-input-number
                                        style="width: 150px"
                                        v-model:value="item.max_connections"
                                        min="0" />
                                </a-form-item>
                                <a-form-item
                                    :label="t('stress.task.requestTimeout')"
                                    name="timeout"
                                    rules="required">
                                    <a-input-number
                                        :placeholder="$t('stress.requestTimeoutPlaceholder')"
                                        style="width: 150px"
                                        v-model:value="item.timeout"
                                        min="0" />
                                </a-form-item>

                                <a-form-item :label="$t('stress.task.otherOptions')">
                                    <a-checkbox-group v-model:value="item.options">
                                        <a-checkbox
                                            value="disableCompression"
                                            name="disableCompression"
                                            >{{ t('stress.task.disableCompression') }}
                                        </a-checkbox>
                                        <a-checkbox
                                            value="disableKeepAlive"
                                            name="disableKeepAlive"
                                            >{{ t('stress.task.disableKeepAlive') }}
                                        </a-checkbox>
                                        <a-checkbox
                                            value="disableRedirect"
                                            name="disableRedirect"
                                            >{{ t('stress.task.disableRedirect') }}
                                        </a-checkbox>
                                        <a-checkbox
                                            value="enableHttp2"
                                            name="enableHttp2"
                                            >{{ t('stress.task.enableHttp2') }}
                                        </a-checkbox>
                                    </a-checkbox-group>
                                </a-form-item>
                            </a-tab-pane>
                        </a-tabs>
                    </a-form>
                </a-collapse-panel>
            </a-collapse>
            <a-button
                style="margin-top: 20px"
                class="add-request-button"
                @click="addOption">
                <PlusOutlined />
                {{ t('stress.task.addRequest') }}
            </a-button>
        </a-form-item>
    </a-form>
</template>

<script setup>
import { useI18n } from 'vue-i18n'
import { computed, reactive, ref } from 'vue'
import apis from '@/apis'
import { DeleteOutlined, PlusOutlined, RightSquareOutlined, ReloadOutlined, CopyOutlined } from '@ant-design/icons-vue'
import { parse, stringify } from 'lossless-json'
import 'codemirror/mode/go/go.js'
import Codemirror from 'codemirror-editor-vue3'
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
const router = useRouter()

const LosslessJSONParser = { parse, stringify }

const { t } = useI18n() // 解构出t方法

const route = useRoute()

onMounted(async () => {
    const planId = route.query.plan_id
    if (planId) {
        const { data } = await apis.stress.planQuery({ plan_id: planId })
        // 填充表单
        Object.assign(stressForm, data)
    }
})

const defaultDynamicScript = `import "encoding/json"

type DynamicParam struct {
	Headers map[string]string
	Query   map[string]string
	Body    string
}

// function name must be GetParams
func GetParams() string {
	// write code, return json.Marshal([]DynamicParam{})
	var params []DynamicParam
	params = append(params, DynamicParam{
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Query: map[string]string{
			"a": "1",
			"b": "2",
		},
		Body: "{}", // json string
	})
	s, _ := json.Marshal(params)
	return string(s)
}
`

const defaultResponseCheckScript = `import (
	"encoding/json"
	"fmt"
	"strconv"
)

// function name must be Check
func Check(responseBody string) string {
	var data map[string]any
	err := json.Unmarshal([]byte(responseBody), &data)
	if err != nil {
		return "error parse"
	}
    // if the response body has code key, and the value is 200, return good
	if iCode, err := strconv.Atoi(fmt.Sprintf("%v", data["code"])); err == nil && iCode == 200 {
		return "good"
	}
    // else return bad
	return "bad"
}
`

const stressForm = reactive({
    plan_name: '',
    stress_time: null,
    stress_type: 1,
    stress_mode: 1,
    step_interval_time: 0,
    tasks: [],
})

const formRefs = ref([])
const activeTabKey = ref([])

const cmOptions = {
    mode: 'text/x-go',
    tabSize: 4,
}

const addOption = () => {
    activeTabKey.value.push(stressForm.tasks.length.toString())
    stressForm.tasks.push({ url: '', num: 0, method: 'GET', bodyJson: {}, body: '' })
    formRefs.value.push(ref({}))
}

const resetForm = () => {
    stressForm.plan_name = ''
    stressForm.stress_time = null
    stressForm.stress_type = 1
    stressForm.stress_mode = 1
    stressForm.step_interval_time = 0
    stressForm.tasks = []
    formRefs.value = []
    activeTabKey.value = []
    addOption()
}

const startStress = async () => {
    // 发起api请求， 组装请求
    const { success } = await apis.stress.startStress({ ...stressForm }).catch(() => {
        throw new Error()
    })
    if (success) {
        message.success(t('stress.start.success'))
        router.push({ name: 'record' })
    } else {
        message.error(t('stress.start.failed'))
    }
}

addOption()

const removeItem = (index) => {
    stressForm.tasks.splice(index, 1)
    formRefs.value.splice(index, 1)
    activeTabKey.value.splice(index, 1)
}

function duplicateItem(index) {
    const newItem = { ...stressForm.tasks[index] }
    stressForm.tasks.splice(index + 1, 0, newItem)
    formRefs.value.splice(index + 1, 0, ref({}))
    activeTabKey.value.push(String(index + 1))
}

const concurrencyOrRpsLabelText = computed(() => {
    if (stressForm.stress_mode === 2) {
        if (stressForm.stress_type === 1) {
            return t('stress.task.startRps')
        }
        if (stressForm.stress_type === 2) {
            return t('stress.task.startConcurrencyNum')
        }
    }
    if (stressForm.stress_type === 1) {
        return t('stress.task.rps')
    }
    return t('stress.task.concurrencyNum')
})

const maxConcurrencyOrRpsLabelText = computed(() => {
    if (stressForm.stress_type === 1) {
        return t('stress.task.maxRps')
    }
    return t('stress.task.maxConcurrencyNum')
})

const stepConcurrencyOrRpsLabelText = computed(() => {
    if (stressForm.stress_type === 1) {
        return t('stress.task.stepRps')
    }
    return t('stress.task.stepConcurrencyNum')
})

const onBodyChange = (item, content) => {
    item.body = content.text
}

const formItemLayout = {
    // TODO  这里如果浏览器变窄，会导致表单label展示不全
    wrapperCol: {
        xs: { span: 24 },
        sm: { span: 20 },
    },
}
</script>

<style lang="less" scoped>
.panel-header {
    display: flex;
    align-tasks: center;
}

.start-stress-button {
    background-color: #2dc26b;
}

.add-request-button {
    background-color: #f0f0f0;
}

.reset-button {
    margin-left: 5px;
}

.button-group {
    display: flex;
    gap: 8px;
}

.form-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    position: relative;
}

.delete-button {
    position: absolute;
    top: 8px;
    right: 8px;
}

.duplicate-button {
    background-color: #e67e23;
    position: absolute;
    top: 8px;
    right: 8px;
    margin-right: 58px;
}

.custom-tooltip {
    background-color: #fff;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
    border-radius: 4px;
    padding: 12px 16px;
}
</style>
