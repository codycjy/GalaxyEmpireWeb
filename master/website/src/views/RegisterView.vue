<template>
  <div class="wzf_container">
    <div class="login">
      <h1>注册</h1>
      <div class="selected">
        <el-form
          :model="registerForm"
          :rules="rules"
          label-width="0px"
          class="login-form"
          ref="registerForm"
          status-icon>
          <el-form-item prop="account">
            <el-input
              v-model="registerForm.account"
              placeholder="账号">
            </el-input>
          </el-form-item>
          <el-form-item prop="pwd">
            <el-input
              v-model="registerForm.pwd"
              show-password
              placeholder="密码">
            </el-input>
          </el-form-item>
          <el-form-item prop="checkPwd">
            <el-input
              v-model="registerForm.checkPwd"
              show-password
              placeholder="确认密码"
              @keyup.enter.native="register('registerForm')">
            </el-input>
          </el-form-item>
          <el-form-item prop="captcha">
            <el-row :gutter="20">
              <el-col :span="12">
                <el-input
                  v-model="registerForm.captcha"
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
        <div class="clearfix"></div>
        <div class="btns">
          <el-button
            type="primary"
            @click="register('registerForm')">注册
          </el-button>
        </div>
        <div class="register">
          <router-link to="/login">登录</router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { account, registerPwd, captcha } from '@/util/validateRule'
import { getCaptchaId, getCaptchaPhoto, userRegister } from '@/api/log'
export default {
  created () {
    this.getCaptcha()
  },
  data () {
    const validateCheckPwd = (rule, value, callback) => {
      if (value === '') {
        callback(new Error('请再次输入密码'))
      } else if (value !== this.registerForm.pwd) {
        callback(new Error('两次输入密码不一致!'))
      } else {
        callback()
      }
    }
    return {
      activeName: 'pwdLogin',
      registerForm: {
        account: '',
        pwd: '',
        checkPwd: '',
        captcha: '',
        captchaId: ''
      },
      captchaUrl: '',
      rules: {
        account,
        pwd: registerPwd,
        checkPwd: [
          { validator: validateCheckPwd, trigger: 'blur' }
        ],
        captcha
      }
    }
  },
  methods: {
    register (formName) {
      this.$refs[formName].validate((valid) => {
        if (valid) {
          userRegister(this.registerForm).then(response => {
            this.$message({
              showClose: true,
              message: '注册成功',
              type: 'success'
            })
          }).catch(err => {
            console.log(err)
          })
        } else {
          console.log('error submit!!')
          return false
        }
      })
    },
    async getCaptcha () {
      const captchaInfo = await getCaptchaId().catch(err => console.log(err))
      if (captchaInfo) {
        this.registerForm.captchaId = captchaInfo.captcha_id
        const captchaBinaryData = await getCaptchaPhoto(captchaInfo.captcha_id).catch(err => console.log(err))
        if (!captchaBinaryData) return
        this.captchaUrl = window.URL.createObjectURL(
          new Blob(
            [captchaBinaryData],
            { type: 'image/png' }))
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
    &-captcha {
      height: 40px;
    }
    .selected {
      width: 80%;
      margin: auto;
      padding: 30px 0;
      .infor {
        margin-top: 15px;
      }
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
