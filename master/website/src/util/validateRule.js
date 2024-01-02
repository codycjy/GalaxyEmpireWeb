const validateAccount = (rule, value, callback) => {
  if (value === '') {
    callback(new Error('账号不能为空'))
  } else {
    const chineseRegex = /[\u4e00-\u9fa5]/
    if (chineseRegex.test(value)) {
      callback(new Error('账号不能包含汉字'))
    } else {
      callback()
    }
  }
}

const pwdRe = /^.{8,16}$/

export const account = [
  {
    validator: validateAccount,
    trigger: 'change'
  }
]

export const registerPwd = [
  { required: true, message: '密码不能为空', trigger: 'change' },
  { pattern: pwdRe, message: '密码8-16位' }
]

export const loginPwd = [
  { required: true, message: '密码不能为空', trigger: 'change' }
]
