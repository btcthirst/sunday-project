<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <n-icon size="64" color="#18a058">
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 3c1.66 0 3 1.34 3 3s-1.34 3-3 3-3-1.34-3-3 1.34-3 3-3zm0 14.2c-2.5 0-4.71-1.28-6-3.22.03-1.99 4-3.08 6-3.08 1.99 0 5.97 1.09 6 3.08-1.29 1.94-3.5 3.22-6 3.22z"/>
          </svg>
        </n-icon>
        <h1>Паспортний стіл</h1>
        <p v-if="isFirstRun">Створіть обліковий запис оператора</p>
        <p v-else>Вхід в систему</p>
      </div>

      <n-form ref="formRef" :model="formData" :rules="rules" @submit.prevent="handleSubmit">
        <!-- First run: setup form -->
        <template v-if="isFirstRun">
          <n-form-item label="Ім'я користувача" path="username">
            <n-input 
              v-model:value="formData.username" 
              placeholder="Введіть логін"
              :disabled="loading"
            />
          </n-form-item>
          
          <n-form-item label="Повне ім'я" path="fullName">
            <n-input 
              v-model:value="formData.fullName" 
              placeholder="Прізвище Ім'я По-батькові"
              :disabled="loading"
            />
          </n-form-item>

          <n-form-item label="Пароль" path="password">
            <n-input
              v-model:value="formData.password"
              type="password"
              show-password-on="click"
              placeholder="Введіть пароль"
              :disabled="loading"
            />
          </n-form-item>

          <n-form-item label="Підтвердження пароля" path="confirmPassword">
            <n-input
              v-model:value="formData.confirmPassword"
              type="password"
              show-password-on="click"
              placeholder="Повторіть пароль"
              :disabled="loading"
            />
          </n-form-item>
        </template>

        <!-- Regular login form -->
        <template v-else>
          <n-form-item label="Ім'я користувача" path="username">
            <n-input 
              v-model:value="formData.username" 
              placeholder="Введіть логін"
              :disabled="loading"
            />
          </n-form-item>

          <n-form-item label="Пароль" path="password">
            <n-input
              v-model:value="formData.password"
              type="password"
              show-password-on="click"
              placeholder="Введіть пароль"
              :disabled="loading"
              @keyup.enter="handleSubmit"
            />
          </n-form-item>
        </template>

        <n-button
          type="primary"
          block
          strong
          :loading="loading"
          @click="handleSubmit"
        >
          {{ isFirstRun ? 'Створити та увійти' : 'Увійти' }}
        </n-button>
      </n-form>

      <div v-if="error" class="error-message">
        <n-alert type="error" :title="error" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { IsFirstRun, Login, SetupInitialOperator } from '../../wailsjs/go/main/App'

const router = useRouter()
const message = useMessage()

const isFirstRun = ref(true)
const loading = ref(false)
const error = ref('')

const formRef = ref()
const formData = ref({
  username: '',
  password: '',
  confirmPassword: '',
  fullName: ''
})

const rules = {
  username: {
    required: true,
    message: 'Введіть ім\'я користувача',
    trigger: 'blur'
  },
  password: {
    required: true,
    message: 'Введіть пароль',
    trigger: 'blur'
  },
  fullName: {
    required: true,
    message: 'Введіть повне ім\'я',
    trigger: 'blur'
  },
  confirmPassword: {
    required: true,
    validator: (_rule: any, value: string) => {
      if (!value) {
        return new Error('Підтвердіть пароль')
      }
      if (value !== formData.value.password) {
        return new Error('Паролі не співпадають')
      }
      return true
    },
    trigger: 'blur'
  }
}

onMounted(async () => {
  try {
    isFirstRun.value = await IsFirstRun()
  } catch (e) {
    console.error('Failed to check first run:', e)
    isFirstRun.value = true
  }
})

async function handleSubmit() {
  error.value = ''
  loading.value = true

  try {
    await formRef.value?.validate()
  } catch {
    loading.value = false
    return
  }

  try {
    if (isFirstRun.value) {
      await SetupInitialOperator(
        formData.value.username,
        formData.value.password,
        formData.value.fullName
      )
      message.success('Обліковий запис створено!')
    } else {
      await Login(formData.value.username, formData.value.password)
    }
    
    router.push('/dashboard')
  } catch (e: any) {
    error.value = e.toString()
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
  padding: 20px;
}

.login-card {
  background: rgba(24, 24, 28, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  padding: 48px 40px;
  width: 100%;
  max-width: 420px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-header h1 {
  color: #fff;
  font-size: 1.75rem;
  font-weight: 600;
  margin: 16px 0 8px;
}

.login-header p {
  color: rgba(255, 255, 255, 0.6);
  margin: 0;
}

.error-message {
  margin-top: 16px;
}
</style>
