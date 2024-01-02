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
              placeholder="密码"
              @keyup.enter.native="login('loginForm')">
            </el-input>
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
            @click="login('loginForm')">登录
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
import { account, loginPwd } from '@/util/validateRule'
import aes from '@/util/aes'
import { userLogin } from '@/api/log'

export default {
  data () {
    return {
      activeName: 'pwdLogin',
      loginForm: {
        account: '',
        password: ''
      },
      checked: false,
      rules: {
        account,
        password: loginPwd
      }
    }
  },
  mounted () {
    this.getCookie()
  },
  methods: {
    login (formName) {
      this.$refs[formName].validate((valid) => {
        if (valid) {
          userLogin(this.loginForm).then(response => {
            localStorage.setItem('token', response.data.token)
            this.$router.push('/home')
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
