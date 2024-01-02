import request from '@/util/request'

export function getAccountsByUserId (id) {
  return request({
    method: 'GET',
    url: `/account/user/${id}`
  })
}
