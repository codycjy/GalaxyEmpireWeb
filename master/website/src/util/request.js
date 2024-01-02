import axios from 'axios'
import { Message } from 'element-ui'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 5000
})

// request 拦截器
// 可以自请求发送前对请求做一些处理
// 比如统一加token，对请求参数统一加密
request.interceptors.request.use(config => {
  if (typeof config.headers['Content-Type'] === 'undefined') config.headers['Content-Type'] = 'application/json;charset=utf-8'
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = token
  } // 设置请求头
  return config
}, error => {
  return Promise.reject(error)
})

// response 拦截器
// 可以在接口响应后统一处理结果
request.interceptors.response.use(
  response => {
    console.log('response', response)
    let res = response.data
    // 如果是返回的文件
    if (response.config.responseType === 'blob') {
      return res
    }
    // 兼容服务端返回的字符串数据
    if (typeof res === 'string') {
      res = res ? JSON.parse(res) : res
    }
    return res
  },
  error => {
    console.log('err' + error) // for debug
    if (error.message === 'Network Error' && !error.response) {
      Message.error(error.message)
    } else {
      Message({
        showClose: true,
        message: error.response.data.message,
        type: 'error'
      })
    }
    return Promise.reject(error)
  }
)

export default request
