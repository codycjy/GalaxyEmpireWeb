import request from '@/util/request'
import aes from '@/util/aes'

export function userLogin (loginData) {
  return request({
    method: 'POST',
    url: '/user/login',
    data: {
      account: loginData.account,
      pwd: aes.encrypt(loginData.password)
    }
  })
}

export function userRegister (registerData) {
  return request({
    method: 'POST',
    url: '/user/register',
    data: {
      account: registerData.account,
      pwd: aes.encrypt(registerData.pwd)
    }
  })
}
