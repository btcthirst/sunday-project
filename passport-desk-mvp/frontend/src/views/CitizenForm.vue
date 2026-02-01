<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-left">
        <n-button quaternary circle @click="router.back()">
          <template #icon>
            <n-icon>
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"/>
              </svg>
            </n-icon>
          </template>
        </n-button>
        <h1 class="page-title">{{ isEdit ? 'Редагування громадянина' : 'Новий громадянин' }}</h1>
      </div>
      <n-space>
        <n-button @click="router.back()">Скасувати</n-button>
        <n-button type="primary" :loading="saving" @click="handleSave">
          Зберегти
        </n-button>
      </n-space>
    </div>

    <n-card :bordered="false" class="form-card">
      <n-form
        ref="formRef"
        :model="formValue"
        :rules="rules"
        label-placement="top"
        size="medium"
      >
        <n-tabs type="line" animated v-model:value="activeTab">
          <n-tab-pane name="personal" tab="Особисті дані">
            <n-grid :x-gap="24" :y-gap="24" :cols="2">
              <n-gi>
                <n-form-item label="Прізвище" path="last_name">
                  <n-input v-model:value="formValue.last_name" placeholder="Іванов" />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="Ім'я" path="first_name">
                  <n-input v-model:value="formValue.first_name" placeholder="Іван" />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="По батькові" path="middle_name">
                  <n-input v-model:value="formValue.middle_name" placeholder="Іванович" />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="Дата народження" path="birth_date">
                  <n-date-picker 
                    v-model:formatted-value="formValue.birth_date"
                    value-format="yyyy-MM-dd"
                    type="date"
                    style="width: 100%"
                  />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="Стать" path="gender">
                  <n-select v-model:value="formValue.gender" :options="genderOptions" />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="Місце народження" path="birth_place">
                  <n-input v-model:value="formValue.birth_place" placeholder="м. Київ" />
                </n-form-item>
              </n-gi>
            </n-grid>
          </n-tab-pane>

          <n-tab-pane name="documents" tab="Документи">
            <n-grid :x-gap="24" :y-gap="24" :cols="2">
              <n-gi>
                <n-form-item label="Серія паспорта" path="passport_series">
                  <n-input 
                    v-model:value="formValue.passport_series" 
                    placeholder="АА" 
                    @input="v => formValue.passport_series = v.toUpperCase()"
                    maxlength="2"
                  />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="Номер паспорта" path="passport_number">
                  <n-input v-model:value="formValue.passport_number" placeholder="123456" maxlength="6" />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="ІПН (РНОКПП)" path="tax_number">
                  <n-input v-model:value="formValue.tax_number" placeholder="1234567890" maxlength="10" />
                </n-form-item>
              </n-gi>
              <n-gi>
                <!-- Placeholder for future document scans -->
              </n-gi>
            </n-grid>
          </n-tab-pane>

          <n-tab-pane name="contacts" tab="Контакти">
            <n-grid :x-gap="24" :y-gap="24" :cols="1">
              <n-gi>
                <n-form-item label="Телефон" path="phone">
                  <n-input v-model:value="formValue.phone" placeholder="+380..." />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="Email" path="email">
                  <n-input v-model:value="formValue.email" placeholder="email@example.com" />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="Примітки" path="notes">
                  <n-input
                    v-model:value="formValue.notes"
                    type="textarea"
                    placeholder="Додаткова інформація"
                  />
                </n-form-item>
              </n-gi>
            </n-grid>
          </n-tab-pane>

          <n-tab-pane name="registration" tab="Реєстрація" :disabled="!isEdit">
            <n-alert v-if="!isEdit" type="info" style="margin-bottom: 16px;">
              Збережіть громадянина, щоб додати реєстрацію
            </n-alert>
            <div v-else>
               <registration-history :citizen-id="citizenId" />
            </div>
          </n-tab-pane>
        </n-tabs>
      </n-form>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { CreateCitizen, GetCitizen, UpdateCitizen } from '../../wailsjs/go/main/App'
import { services } from '../../wailsjs/go/models'
import { ArrowBack } from '@vicons/ionicons5'
import RegistrationHistory from '../components/RegistrationHistory.vue'

const route = useRoute()
const router = useRouter()
const message = useMessage()

const formRef = ref()
const saving = ref(false)
const activeTab = ref('personal')
const isEdit = computed(() => route.params.id !== undefined)
const citizenId = computed(() => {
  const id = route.params.id
  return id ? parseInt(id as string) : 0
})

const formValue = ref(new services.CitizenInput())

// Initialize defaults
formValue.value.gender = 'M'

const genderOptions = [
  { label: 'Чоловік', value: 'M' },
  { label: 'Жінка', value: 'F' }
]

const rules = {
  last_name: { required: true, message: 'Введіть прізвище', trigger: 'blur' },
  first_name: { required: true, message: 'Введіть ім\'я', trigger: 'blur' },
  birth_date: { required: true, message: 'Оберіть дату народження', trigger: ['blur', 'change'] },
  passport_series: { 
    required: true, 
    message: 'Введіть серію', 
    trigger: 'blur',
    validator: (_: any, value: string) => {
      return /^[A-ZА-ЯІЇЄ]{2}$/.test(value) || new Error('2 літери')
    }
  },
  passport_number: { 
    required: true, 
    message: 'Введіть номер', 
    trigger: 'blur',
    validator: (_: any, value: string) => {
      return /^\d{6}$/.test(value) || new Error('6 цифр')
    }
  },
  tax_number: {
    required: true,
    message: 'Введіть ІПН',
    trigger: 'blur',
    validator: (_: any, value: string) => {
      if (!value) return true
      return /^\d{10}$/.test(value) || new Error('10 цифр')
    }
  },
  phone: {
    trigger: 'blur',
    validator: (_: any, value: string) => {
      if (!value) return true
      return /^[\d\+]{10,13}$/.test(value) || new Error('Невірний формат телефону')
    }
  },
  email: {
    trigger: 'blur',
    validator: (_: any, value: string) => {
      if (!value) return true
      return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value) || new Error('Невірний email')
    }
  }
}

onMounted(async () => {
  if (isEdit.value) {
    const id = parseInt(route.params.id as string)
    try {
      const citizen = await GetCitizen(id)
      
      // Fix date format from backend (RFC3339) to YYYY-MM-DD
      let birthDate = citizen.birth_date
      if (birthDate && birthDate.length > 10) {
        birthDate = birthDate.substring(0, 10)
      }

      // Map output to input format
      const input = new services.CitizenInput({
        last_name: citizen.last_name,
        first_name: citizen.first_name,
        middle_name: citizen.middle_name,
        birth_date: birthDate,
        passport_series: citizen.passport_series,
        passport_number: citizen.passport_number,
        tax_number: citizen.tax_number,
        gender: citizen.gender,
        birth_place: citizen.birth_place,
        phone: citizen.phone,
        email: citizen.email,
        notes: citizen.notes
      })
      formValue.value = input
      
    } catch (e: any) {
      console.error(e)
      message.error('Помилка завантаження: ' + e.toString())
      router.push({ name: 'Citizens' })
    }
  }
})

async function handleSave() {
  try {
    await formRef.value?.validate()
  } catch {
    message.error('Перевірте правильність заповнення полів')
    return
  }

  saving.value = true
  try {
    if (isEdit.value) {
      const id = parseInt(route.params.id as string)
      await UpdateCitizen(id, formValue.value)
      message.success('Зміни збережено')
    } else {
      await CreateCitizen(formValue.value)
      message.success('Громадянина створено')
    }
    router.back()
  } catch (e: any) {
    message.error('Помилка збереження: ' + e.toString())
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.page-container {
  max-width: 1000px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.page-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}
</style>
