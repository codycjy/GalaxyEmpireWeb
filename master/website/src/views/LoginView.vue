<template>
  <div class="wzf_container">
    <div class="login">
      <h1>登录</h1>
      <div class="selected">
        <el-form
          :model="loginForm"
          :rules="rules"
          label-width="0px"
          class="login-form"
          ref="loginForm">
          <el-form-item prop="account">
            <el-input
              v-model="loginForm.account"
              placeholder="账号">
            </el-input>
          </el-form-item>
          <el-form-item prop="password">
            <el-input
              v-model="loginForm.password"
              show-password
              placeholder="密码">
            </el-input>
          </el-form-item>
          <el-form-item prop="captcha">
            <el-row :gutter="20">
              <el-col :span="12">
                <el-input
                  v-model="loginForm.captcha"
                  placeholder="验证码">
                </el-input>
              </el-col>
              <el-col :span="12">
                <img
                  :src="captchaUrl"
                  alt="加载中"
                  class="login-captcha"
                  @click="getCaptcha"
                >
              </el-col>
            </el-row>
          </el-form-item>
        </el-form>
        <div class="tips">
          <div class="remeberMe">
            <el-checkbox v-model="checked">记住我</el-checkbox>
          </div>
          <div class="forgetPwd">
            <a href="/#">忘记密码？</a>
          </div>
        </div>
        <div class="clearfix"></div>
        <div class="btns">
          <el-button
            type="primary"
            @click="login">登录
          </el-button>
        </div>
        <div class="register">
          <router-link to="/register">注册账号</router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { account, loginPwd, captcha } from '@/util/validateRule'
import aes from '@/util/aes'
import { userLogin, getCaptchaId, getCaptchaPhoto } from '@/api/log'

export default {
  created () {
    this.getCaptcha()
  },
  data () {
    return {
      loginForm: {
        account: '',
        password: '',
        captcha: '',
        captchaId: ''
      },
      captchaUrl: '',
      checked: false,
      rules: {
        account,
        password: loginPwd,
        captcha
      }
    }
  },
  mounted () {
    this.getCookie()
  },
  methods: {
    login () {
      this.$refs.loginForm.validate((valid) => {
        if (valid) {
          console.log(this.loginForm)
          userLogin(this.loginForm).then(response => {
            localStorage.setItem('token', response.token)
            this.$router.push('/home')
            this.$message({
              showClose: true,
              message: '登录成功',
              type: 'success'
            })
          }).catch(err => {
            console.log(err)
          })
          // 是否写入cookie
          if (this.checked) {
            this.setCookie(this.loginForm.account, aes.encrypt(this.loginForm.password), 7)
          } else {
            this.setCookie('', '', -1)
          }
        } else {
          console.log('error submit!!')
          return false
        }
      })
    },
    setCookie (account, password, days) {
      const date = new Date()
      date.setTime(date.getTime() + 24 * 60 * 60 * 1000 * days)
      window.document.cookie = 'account=' + account + ';path=/;expires=' + date.toGMTString()
      window.document.cookie = 'pwd=' + password + ';path=/;expires=' + date.toGMTString()
    },
    getCookie () {
      if (document.cookie.length > 0) {
        const arr = document.cookie.split('; ')
        console.log('cookie' + arr)
        for (let i = 0; i < arr.length; i++) {
          const arr2 = arr[i].split('=')
          if (arr2[0] === 'account') {
            this.loginForm.account = arr2[1]
          } else if (arr2[0] === 'pwd') {
            console.log('cookie', arr2[1])
            this.loginForm.password = aes.decrypt(arr2[1])
            this.checked = true
          }
        }
      }
    },
    async getCaptcha () {
      const captchaInfo = await getCaptchaId()
      this.loginForm.captchaId = captchaInfo.captcha_id
      const captchaBinaryData = await getCaptchaPhoto(captchaInfo.captcha_id)
      this.captchaUrl = window.URL.createObjectURL(
        new Blob(
          [captchaBinaryData],
          { type: 'image/png' }))
    }
  }
}
</script>

<style scoped lang="less">
.wzf_container {
  width: 100%;
  height: 100%;
  display: grid;
  background-color: #f5f5f5;
  position: absolute;
  .login {
    min-height: 440px;
    min-width: 440px;
    justify-self: center;
    align-self: center;
    background-color: white;
    border: 3px solid white;
    border-radius: 10px;

    &-captcha {
      height: 40px;
    }
    h2 {
      font-weight: 300;
    }
    .selected {
      width: 80%;
      margin: auto;
      padding: 30px 0;
      .infor {
        margin-top: 15px;
      }
    }
    .remeberMe {
      float: right;
    }
    a {
      float: left;
      color: blue;
      text-decoration: none;
    }
    .btns {
      margin-top: 15px;
      el-button {
        background-color: blue;
        border: 1px solid #eee;
        border-radius: 10px;
      }
  }
    .register {
      margin-top: 15px;
      float: right;
      a {
        color: gray;
      }
      a:active {
        color: blue;
      }
    }
  }
}
.clearfix {
  clear: both;
}
</style>
