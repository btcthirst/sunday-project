<template>
  <div>
    <n-space justify="end" style="margin-bottom: 16px;">
      <n-button type="primary" @click="showModal = true">
        <template #icon>
          <n-icon>
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
              <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/>
            </svg>
          </n-icon>
        </template>
        Нова реєстрація
      </n-button>
    </n-space>

    <n-data-table
      :columns="columns"
      :data="history"
      :loading="loading"
      :pagination="false"
      size="small"
    />

    <registration-modal
      v-model:show="showModal"
      :citizen-id="citizenId"
      @saved="loadHistory"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, h, watch } from 'vue'
import { DataTableColumns, NTag, NButton, useMessage, useDialog, NSpace, NIcon, NDataTable } from 'naive-ui'
import { GetRegistrationHistory, DeregisterCitizen } from '../../wailsjs/go/main/App'
import { database } from '../../wailsjs/go/models'
import RegistrationModal from './RegistrationModal.vue'

const props = defineProps<{
  citizenId: number
}>()

const message = useMessage()
const dialog = useDialog()
const loading = ref(false)
const history = ref<database.RegistrationOutput[]>([])
const showModal = ref(false)

const columns: DataTableColumns<database.RegistrationOutput> = [
  {
    title: 'Дата',
    key: 'registration_date',
    width: 120
  },
  {
    title: 'Тип',
    key: 'registration_type',
    width: 120,
    render(row) {
      const isPerm = row.registration_type === 'permanent'
      return h(NTag, { type: isPerm ? 'info' : 'warning', bordered: false }, {
        default: () => isPerm ? 'Постійна' : 'Тимчасова'
      })
    }
  },
  {
    title: 'Адреса',
    key: 'address',
    render(row) {
      const parts = [
        row.region,
        row.district,
        row.settlement,
        row.street,
        row.house_number,
        row.apartment_number ? `кв. ${row.apartment_number}` : null
      ].filter(Boolean)
      return parts.join(', ')
    }
  },
  {
    title: 'Статус',
    key: 'is_active',
    width: 100,
    render(row) {
      return h(NTag, { type: row.is_active ? 'success' : 'default', bordered: false }, {
        default: () => row.is_active ? 'Активна' : 'Архів'
      })
    }
  },
  {
    title: 'Вибуття',
    key: 'deregistration_date',
    width: 120
  },
  {
    title: 'Дії',
    key: 'actions',
    width: 100,
    render(row) {
      if (!row.is_active) return null
      return h(NButton, {
        size: 'tiny',
        type: 'error',
        secondary: true,
        onClick: () => confirmDeregister(row)
      }, { default: () => 'Зняти' })
    }
  }
]

async function loadHistory() {
  if (!props.citizenId) return
  
  loading.value = true
  try {
    const res = await GetRegistrationHistory(props.citizenId)
    history.value = res || []
  } catch (err) {
    console.error(err)
    message.error('Не вдалося завантажити історію')
  } finally {
    loading.value = false
  }
}

function confirmDeregister(row: database.RegistrationOutput) {
  dialog.warning({
    title: 'Зняття з реєстрації',
    content: 'Ви впевнені, що хочете зняти громадянина з реєстрації?',
    positiveText: 'Так, зняти',
    negativeText: 'Скасувати',
    onPositiveClick: async () => {
      try {
        const today = new Date().toISOString().split('T')[0]
        await DeregisterCitizen(row.id, today)
        message.success('Знято з реєстрації')
        loadHistory()
      } catch (err: any) {
        message.error('Помилка: ' + err.message)
      }
    }
  })
}

watch(() => props.citizenId, (newId) => {
  if (newId) {
    loadHistory()
  }
}, { immediate: true })
</script>
