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
              <n-gi :span="2">
                <n-form-item label="Зразок паспорта" path="passport_type">
                  <n-radio-group v-model:value="formValue.passport_type" name="passport_type">
                    <n-space>
                      <n-radio value="old">Старий зразок (книжечка)</n-radio>
                      <n-radio value="new">Новий зразок (ID-картка)</n-radio>
                    </n-space>
                  </n-radio-group>
                </n-form-item>
              </n-gi>
              <n-gi v-if="formValue.passport_type === 'old'">
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
                <n-form-item :label="formValue.passport_type === 'new' ? 'Номер ID-картки' : 'Номер паспорта'" path="passport_number">
                  <n-input 
                    v-model:value="formValue.passport_number" 
                    :placeholder="formValue.passport_type === 'new' ? '123456789' : '123456'" 
                    :maxlength="formValue.passport_type === 'new' ? 9 : 6" 
                  />
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

          <n-tab-pane name="family" tab="Сім'я">
            <n-space vertical :size="20">
              <n-card bordered title="Додати члена сім'ї" size="small">
                <n-grid :x-gap="12" :cols="3">
                  <n-gi :span="1">
                    <n-form-item label="Пошук громадянина">
                      <n-select
                        v-model:value="selectedMemberID"
                        filterable
                        placeholder="Прізвище..."
                        :options="memberOptions"
                        :loading="searchingMembers"
                        clearable
                        remote
                        @search="handleSearchMembers"
                      />
                    </n-form-item>
                  </n-gi>
                  <n-gi :span="1">
                    <n-form-item label="Родинний зв'язок">
                      <n-select 
                        v-model:value="selectedRelation" 
                        :options="relationOptions" 
                        placeholder="Оберіть..."
                      />
                    </n-form-item>
                  </n-gi>
                  <n-gi :span="1">
                    <n-form-item label=" ">
                      <n-button type="primary" block @click="addRelation" :disabled="!selectedMemberID || !selectedRelation">
                        Додати
                      </n-button>
                    </n-form-item>
                  </n-gi>
                </n-grid>
              </n-card>

              <n-table :single-line="false" size="small">
                <thead>
                  <tr>
                    <th>ПІБ</th>
                    <th>Зв'язок</th>
                    <th style="width: 80px">Дії</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(rel, index) in familyRelations" :key="index">
                    <td>{{ rel.full_name }}</td>
                    <td>{{ rel.relation_type }}</td>
                    <td>
                      <n-button size="small" type="error" ghost @click="removeRelation(index)">
                        Видалити
                      </n-button>
                    </td>
                  </tr>
                  <tr v-if="familyRelations.length === 0">
                    <td colspan="3" style="text-align: center; color: #999; padding: 20px">
                      Членів сім'ї не додано
                    </td>
                  </tr>
                </tbody>
              </n-table>
            </n-space>
          </n-tab-pane>

          <n-tab-pane name="registration" tab="Реєстрація" :disabled="!isEdit">
            <n-alert v-if="!isEdit" type="info" style="margin-bottom: 16px;">
              Збережіть громадянина, щоб додати реєстрацію
            </n-alert>
            <div v-else>
               <n-space justify="space-between" align="center" style="margin-bottom: 16px">
                 <n-button @click="printCertificate" :loading="printing">
                   <template #icon>
                     <n-icon>
                       <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                         <path d="M19 8h-1V3H6v5H5c-1.66 0-3 1.34-3 3v6h3v4h14v-4h3v-6c0-1.66-1.34-3-3-3zM8 5h8v3H8V5zm8 12v2H8v-4h8v2zm2-2v-2H6v2H4v-4c0-.55.45-1 1-1h14c.55 0 1 .45 1 1v4h-2z"/>
                         <path d="M18 11.5c.28 0 .5-.22.5-.5s-.22-.5-.5-.5-.5.22-.5.5.22.5.5.5z"/>
                       </svg>
                     </n-icon>
                   </template>
                   Друк довідки (Зберегти)
                 </n-button>
               </n-space>
               <registration-history :citizen-id="citizenId" />
            </div>
          </n-tab-pane>
        </n-tabs>
      </n-form>
    </n-card>

    <certificate-options-modal
      v-model:show="showCertModal"
      title="Налаштування довідки про проживання"
      @confirm="handleCertConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage, NTable, NAlert, NCard } from 'naive-ui'
import { 
  CreateCitizen, GetCitizen, UpdateCitizen, GenerateCitizenCertificate,
  SearchCitizens, GetFamilyMembers
} from '../../wailsjs/go/main/App'
import { services } from '../../wailsjs/go/models'
import RegistrationHistory from '../components/RegistrationHistory.vue'
import CertificateOptionsModal from '../components/CertificateOptionsModal.vue'

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
formValue.value.passport_type = 'old'

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
      if (formValue.value.passport_type === 'new') return true
      return /^[A-ZА-ЯІЇЄ]{2}$/.test(value || '') || new Error('2 літери')
    }
  },
  passport_number: { 
    required: true, 
    message: 'Введіть номер', 
    trigger: 'blur',
    validator: (_: any, value: string) => {
      const val = value || ''
      if (formValue.value.passport_type === 'new') {
        return /^\d{9}$/.test(val) || new Error('9 цифр')
      }
      return /^\d{6}$/.test(val) || new Error('6 цифр')
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

// --- Family Relations Logic ---
const familyRelations = ref<any[]>([])
const selectedMemberID = ref<number | null>(null)
const selectedRelation = ref<string>('')
const searchingMembers = ref(false)
const memberOptions = ref<any[]>([])

const relationOptions = [
  { label: 'Чоловік', value: 'Чоловік' },
  { label: 'Дружина', value: 'Дружина' },
  { label: 'Син', value: 'Син' },
  { label: 'Дочка', value: 'Дочка' },
  { label: 'Мати', value: 'Мати' },
  { label: 'Батько', value: 'Батько' },
  { label: 'Брат', value: 'Брат' },
  { label: 'Сестра', value: 'Сестра' },
  { label: 'Дідусь', value: 'Дідусь' },
  { label: 'Бабуся', value: 'Бабуся' },
  { label: 'Онук', value: 'Онук' },
  { label: 'Онука', value: 'Онука' }
]

async function handleSearchMembers(query: string) {
  if (!query) return
  searchingMembers.value = true
  try {
    const results = await SearchCitizens(query, 'name')
    memberOptions.value = results
      .filter(c => c.id !== citizenId.value) // Don't allow self as family member
      .map(c => ({
        label: `${c.full_name} (${c.birth_date})`,
        value: c.id,
        full_info: c
      }))
  } catch (e) {
    console.error(e)
  } finally {
    searchingMembers.value = false
  }
}

function addRelation() {
  const member = memberOptions.value.find(m => m.value === selectedMemberID.value)
  if (!member) return

  // Avoid duplicates
  if (familyRelations.value.find(r => r.member_id === selectedMemberID.value)) {
    message.warning('Цей громадянин вже доданий до списку')
    return
  }

  familyRelations.value.push({
    member_id: selectedMemberID.value,
    full_name: member.full_info.full_name,
    relation_type: selectedRelation.value
  })

  // Clear selection
  selectedMemberID.value = null
}

function removeRelation(index: number) {
  familyRelations.value.splice(index, 1)
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
        passport_type: citizen.passport_type || 'old',
        tax_number: citizen.tax_number,
        gender: citizen.gender,
        birth_place: citizen.birth_place,
        phone: citizen.phone,
        email: citizen.email,
        notes: citizen.notes
      })
      formValue.value = input
      
      // Load family members
      try {
        const members = await GetFamilyMembers(id)
        familyRelations.value = members.map(m => ({
          member_id: m.id,
          full_name: m.full_name,
          relation_type: m.relation_type
        }))
      } catch (e) {
        console.error('Failed to load family members', e)
      }
      
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
    // Attach relations to form value
    formValue.value.family_relations = familyRelations.value.map(r => ({
      member_id: r.member_id,
      relation_type: r.relation_type
    }))

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

const printing = ref(false)
const showCertModal = ref(false)

async function handleCertConfirm(opts: services.FamilyCertificateOptions) {
  printing.value = true
  try {
    const path = await GenerateCitizenCertificate(citizenId.value, opts)
    if (path === 'cancelled') {
        message.info('Збереження скасовано')
    } else {
        message.success('Довідку збережено: ' + path)
    }
  } catch (err: any) {
    const msg = err.message || err.toString()
    message.error('Неможливо сформувати довідку: ' + msg)
  } finally {
    printing.value = false
  }
}

async function printCertificate() {
  showCertModal.value = true
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
