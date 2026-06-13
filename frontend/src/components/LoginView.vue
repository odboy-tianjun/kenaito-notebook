<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <h1 class="login-title">Kenaito Notebook</h1>
        <p class="login-subtitle">{{ isLogin ? '欢迎回来，请登录' : '创建新账号' }}</p>
      </div>

      <el-form
        ref="form"
        :model="form"
        :rules="rules"
        @submit.native.prevent="handleSubmit"
      >
        <el-form-item prop="email">
          <el-input
            v-model="form.email"
            prefix-icon="el-icon-message"
            placeholder="请输入邮箱"
            size="medium"
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            prefix-icon="el-icon-lock"
            type="password"
            placeholder="请输入密码"
            size="medium"
            show-password
          />
        </el-form-item>

        <el-form-item v-if="!isLogin" prop="confirmPassword">
          <el-input
            v-model="form.confirmPassword"
            prefix-icon="el-icon-lock"
            type="password"
            placeholder="请再次输入密码"
            size="medium"
            show-password
          />
        </el-form-item>

        <el-form-item v-if="!isLogin" prop="code">
          <div style="display: flex; gap: 10px;">
            <el-input
              v-model="form.code"
              prefix-icon="el-icon-postcard"
              placeholder="请输入验证码"
              size="medium"
              style="flex: 1;"
            />
            <el-button
              size="medium"
              :disabled="codeCooldown > 0 || !form.email"
              :loading="sendingCode"
              @click="handleSendCode"
            >
              {{ codeCooldown > 0 ? codeCooldown + 's' : '发送验证码' }}
            </el-button>
          </div>
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            native-type="submit"
            :loading="loading"
            style="width: 100%"
            size="medium"
          >
            {{ loading ? '处理中...' : (isLogin ? '登 录' : '注 册') }}
          </el-button>
        </el-form-item>
      </el-form>

      <div class="login-footer">
        <span>{{ isLogin ? '还没有账号？' : '已有账号？' }}</span>
        <el-button type="text" @click="toggleMode">
          {{ isLogin ? '立即注册' : '去登录' }}
        </el-button>
      </div>
    </div>
  </div>
</template>

<script>
import { SendVerifyCode } from '../../wailsjs/go/main/App'

export default {
  name: 'LoginView',
  data() {
    const validateConfirmPassword = (rule, value, callback) => {
      if (value !== this.form.password) {
        callback(new Error('两次输入的密码不一致'))
      } else {
        callback()
      }
    }
    return {
      isLogin: true,
      loading: false,
      sendingCode: false,
      codeCooldown: 0,
      cooldownTimer: null,
      form: {
        email: '',
        password: '',
        confirmPassword: '',
        code: ''
      },
      rules: {
        email: [
          { required: true, message: '请输入邮箱', trigger: 'blur' },
          { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }
        ],
        password: [
          { required: true, message: '请输入密码', trigger: 'blur' },
          { min: 6, message: '密码至少 6 个字符', trigger: 'blur' }
        ],
        confirmPassword: [
          { required: true, message: '请再次输入密码', trigger: 'blur' },
          { validator: validateConfirmPassword, trigger: 'blur' }
        ],
        code: [
          { required: true, message: '请输入验证码', trigger: 'blur' },
          { len: 6, message: '验证码为 6 位数字', trigger: 'blur' }
        ]
      }
    }
  },
  beforeDestroy() {
    if (this.cooldownTimer) {
      clearInterval(this.cooldownTimer)
    }
  },
  methods: {
    toggleMode() {
      this.isLogin = !this.isLogin
      if (this.cooldownTimer) {
        clearInterval(this.cooldownTimer)
        this.codeCooldown = 0
      }
      this.form.code = ''
      this.$nextTick(() => {
        if (this.$refs.form) {
          this.$refs.form.clearValidate()
        }
      })
    },
    async handleSendCode() {
      try {
        await this.$refs.form.validateField('email')
      } catch {
        return
      }

      this.sendingCode = true
      try {
        const purpose = 'register'
        await SendVerifyCode(this.form.email, purpose)
        this.$message.success('验证码已发送到您的邮箱')
        this.startCooldown()
      } catch (err) {
        this.$message.error(err.toString() || '发送验证码失败')
      } finally {
        this.sendingCode = false
      }
    },
    startCooldown() {
      this.codeCooldown = 60
      this.cooldownTimer = setInterval(() => {
        this.codeCooldown--
        if (this.codeCooldown <= 0) {
          clearInterval(this.cooldownTimer)
          this.cooldownTimer = null
        }
      }, 1000)
    },
    async handleSubmit() {
      if (!this.$refs.form) return

      try {
        await this.$refs.form.validate()
      } catch {
        return
      }

      this.loading = true
      try {
        if (this.isLogin) {
          await window.go.main.App.Login(this.form.email, this.form.password)
          this.$message.success('登录成功')
        } else {
          await window.go.main.App.Register(this.form.email, this.form.password, this.form.code)
          this.$message.success('注册成功')
        }
        this.$emit('login-success')
      } catch (err) {
        this.$message.error(err.toString() || '操作失败')
      } finally {
        this.loading = false
      }
    }
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: #f0f2f5;
}

.login-card {
  width: 400px;
  padding: 40px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-title {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 8px;
}

.login-subtitle {
  font-size: 14px;
  color: #909399;
}

.login-footer {
  text-align: center;
  margin-top: 16px;
  font-size: 14px;
  color: #909399;
}

.el-form-item.is-required >>> .el-form-item__label::before {
  content: none;
}
</style>
