<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">Звіти</h1>
    </div>

    <n-card title="Експорт списку зареєстрованих громадян" :bordered="false">
      <n-space vertical size="large">
        <n-alert type="info" title="Інформація" :bordered="false">
          Цей звіт формує Excel-файл (.xlsx) зі списком всіх реєстрацій за обраний період.
          Файл включає: ПІБ, дату народження, адресу, дату реєстрації та тип.
        </n-alert>

        <n-form-item label="Період реєстрації">
          <n-date-picker
            v-model:formatted-value="dateRange"
            value-format="yyyy-MM-dd"
            type="daterange"
            clearable
            start-placeholder="З"
            end-placeholder="По"
            class="date-picker-width"
          />
        </n-form-item>

        <n-button type="primary" size="large" @click="handleExport" :loading="loading">
          <template #icon>
            <n-icon>
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/>
              </svg>
            </n-icon>
          </template>
          Завантажити Excel
        </n-button>
      </n-space>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useMessage } from 'naive-ui'
import { ExportCitizens } from '../../wailsjs/go/main/App'

const message = useMessage()
const loading = ref(false)
const dateRange = ref<[string, string] | null>(null)

async function handleExport() {
  loading.value = true
  try {
    const from = dateRange.value ? dateRange.value[0] : ''
    const to = dateRange.value ? dateRange.value[1] : ''
    
    // Call backend
    const savedPath = await ExportCitizens(from, to)
    
    if (savedPath === 'cancelled') {
      message.info('Експорт скасовано')
    } else {
      message.success('Звіт збережено: ' + savedPath)
    }
  } catch (err: any) {
    message.error('Помилка формування звіту: ' + err.message)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.page-container {
  padding: 0;
}

.page-header {
  margin-bottom: 24px;
}

.page-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}

.date-picker-width {
  width: 100%;
  max-width: 400px;
}
</style>
