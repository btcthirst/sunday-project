<template>
  <n-modal
    v-model:show="show"
    preset="dialog"
    :title="title || 'Налаштування довідки'"
    positive-text="Сформувати PDF"
    negative-text="Скасувати"
    @positive-click="handleConfirm"
    @negative-click="show = false"
  >
    <n-form :model="options" label-placement="top">
      <n-grid :cols="2" :x-gap="12">
        <n-gi :span="2">
          <n-form-item label="Назва установи (видавець)">
            <n-input v-model:value="options.issuer_name" placeholder="напр. Житомирська міська рада" />
          </n-form-item>
        </n-gi>
        <n-gi>
          <n-form-item label="Місто">
            <n-input v-model:value="options.city" placeholder="напр. м. Житомир" />
          </n-form-item>
        </n-gi>
        <n-gi>
          <n-form-item label="Куди подається (установа)">
            <n-input v-model:value="options.target_institution" placeholder="напр. управління соцзахисту" />
          </n-form-item>
        </n-gi>
        <n-gi>
          <n-form-item label="Мета">
            <n-input v-model:value="options.purpose" placeholder="напр. за місцем вимоги" />
          </n-form-item>
        </n-gi>
        <n-gi>
          <n-form-item label="Посада підписанта">
            <n-input v-model:value="options.signatory_title" placeholder="напр. Міський голова" />
          </n-form-item>
        </n-gi>
        <n-gi>
          <n-form-item label="ПІБ підписанта">
            <n-input v-model:value="options.signatory_name" placeholder="напр. Іванов Іван Іванович" />
          </n-form-item>
        </n-gi>
      </n-grid>
    </n-form>
  </n-modal>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { services } from '../../wailsjs/go/models'

const props = defineProps<{
  show: boolean
  title?: string
  initialOptions?: Partial<services.FamilyCertificateOptions>
}>()

const emit = defineEmits(['update:show', 'confirm'])

const show = ref(props.show)
const options = reactive({
  issuer_name: props.initialOptions?.issuer_name || '',
  city: props.initialOptions?.city || 'м. Житомир',
  target_institution: props.initialOptions?.target_institution || '',
  purpose: props.initialOptions?.purpose || 'за місцем вимоги',
  signatory_title: props.initialOptions?.signatory_title || 'Міський голова',
  signatory_name: props.initialOptions?.signatory_name || ''
})

watch(() => props.show, (val) => {
  show.value = val
})

watch(show, (val) => {
  emit('update:show', val)
})

function handleConfirm() {
  emit('confirm', { ...options })
  show.value = false
}
</script>
