import request from '@/utils/request'

export const startStress = (params) => request.basic.post('/api/v1/stress/start', params)

export const stopStress = (params) => request.basic.post('/api/v1/stress/stop', params)

export const getPlanList = (params) => request.basic.get('/api/v1/stress/record_plan', params)

export const planQuery = (params) => request.basic.get('/api/v1/stress/plan_query', params)

export const getTaskList = (params) => request.basic.get('/api/v1/stress/record_task', params)

export const getNodeList = () => request.basic.get('/api/v1/nodes/list')

export const getNodeDetail = (params) => request.basic.get('/api/v1/nodes/detail', params)

export const updateNodeQuota = (params) => request.basic.post('/api/v1/nodes/update_quota', params)
