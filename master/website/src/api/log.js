import request from '@/util/request'
import aes from '@/util/aes'

export function userLogin (loginData) {
  return request({
    method: 'POST',
    url: '/login',
    data: {
      username: loginData.account,
      password: aes.encrypt(loginData.password)
    },
    headers: {
      captchaId: loginData.captchaId,
      userInput: loginData.captcha
    }
  })
}

export function userRegister (registerData) {
  return request({
    method: 'POST',
    url: '/register',
    data: {
      username: registerData.account,
      password: aes.encrypt(registerData.pwd)
    },
    headers: {
      captchaId: registerData.captchaId,
      userInput: registerData.captcha
    }
  })
}

export function getCaptchaId () {
  return request({
    method: 'GET',
    url: '/captcha'
  })
}

export function getCaptchaPhoto (captchaId) {
  return request({
    method: 'GET',
    url: `/captcha/${captchaId}`,
    responseType: 'blob'
  })
}
