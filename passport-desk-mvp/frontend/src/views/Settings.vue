<template>
  <div class="settings-page">
    <n-card title="Налаштування" :bordered="false">
      <n-tabs type="line" animated>
        <!-- Profile Tab -->
        <n-tab-pane name="profile" tab="Профіль">
          <n-form ref="profileFormRef" :model="profileForm" :rules="profileRules">
            <n-form-item label="Повне ім'я" path="fullName">
              <n-input v-model:value="profileForm.fullName" placeholder="Введіть повне ім'я" />
            </n-form-item>
            <n-space>
              <n-button type="primary" @click="handleUpdateProfile" :loading="profileLoading">
                Зберегти
              </n-button>
            </n-space>
          </n-form>
        </n-tab-pane>

        <!-- Security Tab -->
        <n-tab-pane name="security" tab="Безпека">
          <n-alert type="warning" title="Важливо" style="margin-bottom: 16px">
            Зміна пароля призведе до перешифрування бази даних. Це може зайняти деякий час.
          </n-alert>
          <n-form ref="passwordFormRef" :model="passwordForm" :rules="passwordRules">
            <n-form-item label="Старий пароль" path="oldPassword">
              <n-input
                v-model:value="passwordForm.oldPassword"
                type="password"
                placeholder="Введіть старий пароль"
                show-password-on="click"
              />
            </n-form-item>
            <n-form-item label="Новий пароль" path="newPassword">
              <n-input
                v-model:value="passwordForm.newPassword"
                type="password"
                placeholder="Введіть новий пароль"
                show-password-on="click"
              />
            </n-form-item>
            <n-form-item label="Підтвердження пароля" path="confirmPassword">
              <n-input
                v-model:value="passwordForm.confirmPassword"
                type="password"
                placeholder="Підтвердіть новий пароль"
                show-password-on="click"
              />
            </n-form-item>
            <n-space>
              <n-button type="primary" @click="handleUpdatePassword" :loading="passwordLoading">
                Змінити пароль
              </n-button>
            </n-space>
          </n-form>
        </n-tab-pane>

        <!-- System Tab -->
        <n-tab-pane name="system" tab="Система">
          <n-space vertical size="large">
            <div>
              <n-h3>Резервні копії</n-h3>
              <n-space vertical>
                <n-button type="primary" @click="handleManualBackup" :loading="backupLoading">
                  Створити резервну копію
                </n-button>
                <n-divider />
                <n-h4>Наявні резервні копії</n-h4>
                <n-spin :show="backupsLoading">
                  <n-list v-if="backups.length > 0" bordered>
                    <n-list-item v-for="backup in backups" :key="backup">
                      <n-thing>
                        <template #header>{{ formatBackupDate(backup) }}</template>
                        <template #description>{{ backup }}</template>
                      </n-thing>
                    </n-list-item>
                  </n-list>
                  <n-empty v-else description="Немає резервних копій" />
                </n-spin>
              </n-space>
            </div>
          </n-space>
        </n-tab-pane>

        <!-- Application Tab -->
        <n-tab-pane name="application" tab="Застосунок">
          <n-space vertical size="large">
            <div>
              <n-h3>Тема</n-h3>
              <n-space align="center">
                <span>Світла</span>
                <n-switch v-model:value="isDark" @update:value="handleThemeToggle">
                  <template #checked>
                    Темна
                  </template>
                  <template #unchecked>
                    Світла
                  </template>
                </n-switch>
                <span>Темна</span>
              </n-space>
            </div>
          </n-space>
        </n-tab-pane>
      </n-tabs>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useMessage, FormInst, FormRules } from 'naive-ui'
import {
  UpdateOperatorProfile,
  UpdateOperatorPassword,
  GetBackupsList,
  TriggerManualBackup,
  GetCurrentOperator
} from '../../wailsjs/go/main/App'
import { useTheme } from '../composables/useTheme'

const message = useMessage()
const { isDark, toggleTheme } = useTheme()

function handleThemeToggle() {
  toggleTheme()
  message.success(`Тему змінено на ${isDark.value ? 'темну' : 'світлу'}`)
}

// Profile form
const profileFormRef = ref<FormInst | null>(null)
const profileForm = ref({
  fullName: ''
})
const profileLoading = ref(false)

const profileRules: FormRules = {
  fullName: [
    { required: true, message: "Введіть повне ім'я", trigger: 'blur' },
    { min: 2, message: "Ім'я повинно містити мінімум 2 символи", trigger: 'blur' }
  ]
}

// Password form
const passwordFormRef = ref<FormInst | null>(null)
const passwordForm = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})
const passwordLoading = ref(false)

const passwordRules: FormRules = {
  oldPassword: [
    { required: true, message: 'Введіть старий пароль', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: 'Введіть новий пароль', trigger: 'blur' },
    { min: 6, message: 'Пароль повинен містити мінімум 6 символів', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: 'Підтвердіть новий пароль', trigger: 'blur' },
    {
      validator: (rule, value) => {
        return value === passwordForm.value.newPassword
      },
      message: 'Паролі не співпадають',
      trigger: 'blur'
    }
  ]
}

// Backups
const backups = ref<string[]>([])
const backupsLoading = ref(false)
const backupLoading = ref(false)

async function loadProfile() {
  try {
    const operator = await GetCurrentOperator()
    profileForm.value.fullName = operator.full_name || ''
  } catch (error) {
    message.error('Не вдалося завантажити профіль')
    console.error(error)
  }
}

async function loadBackups() {
  backupsLoading.value = true
  try {
    backups.value = await GetBackupsList()
  } catch (error) {
    message.error('Не вдалося завантажити список резервних копій')
    console.error(error)
  } finally {
    backupsLoading.value = false
  }
}

async function handleUpdateProfile() {
  if (!profileFormRef.value) return
  
  try {
    await profileFormRef.value.validate()
    profileLoading.value = true
    
    await UpdateOperatorProfile(profileForm.value.fullName)
    message.success('Профіль оновлено')
  } catch (error: any) {
    if (error?.message) {
      message.error(`Помилка: ${error.message}`)
    } else {
      message.error('Не вдалося оновити профіль')
    }
    console.error(error)
  } finally {
    profileLoading.value = false
  }
}

async function handleUpdatePassword() {
  if (!passwordFormRef.value) return
  
  try {
    await passwordFormRef.value.validate()
    passwordLoading.value = true
    
    await UpdateOperatorPassword(
      passwordForm.value.oldPassword,
      passwordForm.value.newPassword
    )
    
    message.success('Пароль успішно змінено')
    
    // Clear form
    passwordForm.value = {
      oldPassword: '',
      newPassword: '',
      confirmPassword: ''
    }
  } catch (error: any) {
    if (error?.message) {
      message.error(`Помилка: ${error.message}`)
    } else {
      message.error('Не вдалося змінити пароль')
    }
    console.error(error)
  } finally {
    passwordLoading.value = false
  }
}

async function handleManualBackup() {
  backupLoading.value = true
  try {
    await TriggerManualBackup()
    message.success('Резервну копію створено')
    await loadBackups()
  } catch (error) {
    message.error('Не вдалося створити резервну копію')
    console.error(error)
  } finally {
    backupLoading.value = false
  }
}

function formatBackupDate(timestamp: string): string {
  // Format: 20060102_150405 -> 02.01.2006 15:04:05
  if (timestamp.length !== 15) return timestamp
  
  const year = timestamp.substring(0, 4)
  const month = timestamp.substring(4, 6)
  const day = timestamp.substring(6, 8)
  const hour = timestamp.substring(9, 11)
  const minute = timestamp.substring(11, 13)
  const second = timestamp.substring(13, 15)
  
  return `${day}.${month}.${year} ${hour}:${minute}:${second}`
}

onMounted(() => {
  loadProfile()
  loadBackups()
})
</script>

<style scoped>
.settings-page {
  max-width: 800px;
}
</style>
