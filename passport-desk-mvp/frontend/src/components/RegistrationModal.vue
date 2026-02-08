<template>
  <n-modal
    v-bind:show="show"
    v-bind:mask-closable="false"
    preset="card"
    title="Реєстрація місця проживання"
    style="width: 600px"
    @close="handleClose"
  >
    <n-form
      ref="formRef"
      :model="formValue"
      :rules="rules"
      label-placement="left"
      label-width="160"
      require-mark-placement="right-hanging"
    >
      <n-form-item label="Тип реєстрації" path="registration_type">
        <n-radio-group v-model:value="formValue.registration_type">
          <n-radio-button value="permanent">Постійна</n-radio-button>
          <n-radio-button value="temporary">Тимчасова</n-radio-button>
        </n-radio-group>
      </n-form-item>

      <n-form-item label="Дата реєстрації" path="registration_date">
        <n-date-picker 
          v-model:formatted-value="formValue.registration_date"
          value-format="yyyy-MM-dd"
          type="date"
          style="width: 100%"
        />
      </n-form-item>

      <n-divider title-placement="left">Адреса</n-divider>

      <n-form-item label="Область" path="region">
        <n-input v-model:value="formValue.region" placeholder="Вінницька" />
      </n-form-item>

      <n-form-item label="Район" path="district">
        <n-input v-model:value="formValue.district" placeholder="Вінницький" />
      </n-form-item>

      <n-form-item label="Населений пункт" path="settlement">
        <n-input v-model:value="formValue.settlement" placeholder="Вінниця" />
      </n-form-item>

      <n-grid :cols="2" :x-gap="12">
        <n-gi>
          <n-form-item label="Вулиця" path="street">
            <n-input v-model:value="formValue.street" />
          </n-form-item>
        </n-gi>
        <n-gi>
          <n-form-item label="Будинок" path="house_number">
            <n-input v-model:value="formValue.house_number" />
          </n-form-item>
        </n-gi>
      </n-grid>

      <n-grid :cols="2" :x-gap="12">
        <n-gi>
          <n-form-item label="Квартира" path="apartment_number">
            <n-input v-model:value="formValue.apartment_number" />
          </n-form-item>
        </n-gi>
        <n-gi>
          <n-form-item label="Основа" path="basis_document">
            <n-input v-model:value="formValue.basis_document" placeholder="Заява, право власності" />
          </n-form-item>
        </n-gi>
      </n-grid>

    </n-form>

    <template #footer>
      <n-space justify="end">
        <n-button @click="handleClose">Скасувати</n-button>
        <n-button type="primary" @click="handleSave" :loading="loading">
          Зберегти
        </n-button>
      </n-space>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { FormInst, useMessage } from 'naive-ui'
import { CreateRegistration } from '../../wailsjs/go/main/App'
import { models } from '../../wailsjs/go/models'

const props = defineProps<{
  show: boolean
  citizenId: number
}>()

const emit = defineEmits(['update:show', 'saved'])

const message = useMessage()
const formRef = ref<FormInst | null>(null)
const loading = ref(false)

const formValue = ref(new models.RegistrationInput({
  registration_type: 'permanent',
  region: 'Вінницька', // Default
  registration_date: new Date().toISOString().split('T')[0]
}))

watch(() => props.show, (newVal) => {
  if (newVal) {
    // Reset form when opened
    formValue.value = new models.RegistrationInput({
      citizen_id: props.citizenId,
      registration_type: 'permanent',
      region: 'Вінницька',
      registration_date: new Date().toISOString().split('T')[0]
    })
  }
})

const rules = {
  registration_type: { required: true },
  registration_date: { required: true, message: 'Оберіть дату', trigger: ['blur', 'change'] },
  region: { required: true, message: 'Вкажіть область', trigger: 'blur' },
  settlement: { required: true, message: 'Вкажіть населений пункт', trigger: 'blur' },
  street: { required: true, message: 'Вкажіть вулицю', trigger: 'blur' },
  house_number: { required: true, message: 'Вкажіть номер будинку', trigger: 'blur' }
}

function handleClose() {
  emit('update:show', false)
}

async function handleSave() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  loading.value = true
  try {
    formValue.value.citizen_id = props.citizenId
    await CreateRegistration(formValue.value)
    message.success('Реєстрацію успішно створено')
    emit('saved')
    handleClose()
  } catch (err: any) {
    message.error('Помилка збереження: ' + (err.message || 'Unknown error'))
  } finally {
    loading.value = false
  }
}
</script>
