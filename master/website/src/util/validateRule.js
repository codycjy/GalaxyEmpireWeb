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

const pwdRe = /^(?![a-zA-Z]+$)(?![A-Z0-9]+$)(?![A-Z\W_]+$)(?![a-z0-9]+$)(?![a-z\W_]+$)(?![0-9\W_]+$)[a-zA-Z0-9\W_]{8,16}$/

export const account = [
  {
    validator: validateAccount,
    trigger: 'change'
  }
]

export const registerPwd = [
  { required: true, message: '密码不能为空', trigger: 'change' },
  { pattern: pwdRe, message: '密码8-16位,且至少包含大小写字母,数字,特殊符号中的三个' }
]

export const loginPwd = [
  { required: true, message: '密码不能为空', trigger: 'change' }
]
