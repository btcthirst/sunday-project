<template>
  <div class="page-container">
    <div class="page-header">
      <div class="title-section">
        <h1 class="page-title">Громадяни</h1>
        <n-input
          v-model:value="searchQuery"
          type="text"
          placeholder="Пошук за ПІБ, паспортом або ІПН"
          class="search-input"
          clearable
          @update:value="handleSearchUpdate"
          @keydown.enter="handleSearch"
        >
          <template #prefix>
            <n-icon>
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                <path d="M15.5 14h-.79l-.28-.27C15.41 12.59 16 11.11 16 9.5 16 5.91 13.09 3 9.5 3S3 5.91 3 9.5 5.91 16 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/>
              </svg>
            </n-icon>
          </template>
        </n-input>
      </div>
      
      <n-button type="primary" @click="router.push({ name: 'NewCitizen' })">
        <template #icon>
          <n-icon>
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
              <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/>
            </svg>
          </n-icon>
        </template>
        Додати громадянина
      </n-button>
    </div>

    <n-card class="table-card" :bordered="false">
      <n-data-table
        :columns="columns"
        :data="citizens"
        :loading="loading"
        :pagination="pagination"
        remote
        @update:page="handlePageChange"
      />
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NTag, NSpace, useMessage, useDialog, type DataTableColumns } from 'naive-ui'
import { ListCitizens, SearchCitizens, DeleteCitizen, RestoreCitizen } from '../../wailsjs/go/main/App'
import { services } from '../../wailsjs/go/models'

const router = useRouter()
const message = useMessage()
const dialog = useDialog()

const citizens = ref<services.CitizenOutput[]>([])
const loading = ref(false)
const searchQuery = ref('')
let searchTimeout: number | null = null

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  try {
    const date = new Date(dateStr)
    if (isNaN(date.getTime())) return dateStr
    return date.toLocaleDateString('uk-UA')
  } catch {
    return dateStr
  }
}

const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  pageCount: 1,
  showSizePicker: false,
})

const columns: DataTableColumns<services.CitizenOutput> = [
  {
    title: 'ПІБ',
    key: 'full_name',
    sorter: 'default',
  },
  {
    title: 'Дата народження',
    key: 'birth_date',
    width: 150,
    render(row) {
      return formatDate(row.birth_date)
    }
  },
  {
    title: 'Адреса реєстрації',
    key: 'active_address',
    width: 250,
    render(row: services.CitizenOutput) {
      return (row as any).active_address || '-'
    }
  },
  {
    title: 'Паспорт',
    key: 'passport_masked',
    width: 180,
    render(row: services.CitizenOutput) {
      return h(NTag, { type: 'info', bordered: false }, { default: () => row.passport_masked })
    }
  },
  {
    title: 'Стать',
    key: 'gender_display',
    width: 100
  },
  {
    title: 'Дії',
    key: 'actions',
    width: 150,
    render(row: services.CitizenOutput) {
      if (row.deleted) {
        return h(NButton, {
          size: 'small',
          type: 'warning',
          secondary: true,
          onClick: () => handleRestore(row)
        }, { default: () => 'Відновити' })
      }
      
      return h(NSpace, null, {
        default: () => [
          h(NButton, {
            size: 'small',
            secondary: true,
            onClick: () => router.push({ name: 'EditCitizen', params: { id: row.id } })
          }, { default: () => 'Ред.' }),
          h(NButton, {
            size: 'small',
            type: 'error',
            ghost: true,
            onClick: () => handleDelete(row)
          }, { default: () => 'Вид.' })
        ]
      })
    }
  }
]

async function loadData(page = 1) {
  loading.value = true
  try {
    if (searchQuery.value) {
      // If searching, we use search API (no pagination for now based on service implementation)
      const results = await SearchCitizens(searchQuery.value, 'name')
      // Also try searching by passport/tax number if name search yields few results or just combine?
      // For MVP simple search implementation:
      if (results.length === 0 && /\d/.test(searchQuery.value)) {
         const passportResults = await SearchCitizens(searchQuery.value, 'passport')
         const taxResults = await SearchCitizens(searchQuery.value, 'tax_number')
         citizens.value = [...passportResults, ...taxResults]
      } else {
         citizens.value = results
      }
      pagination.itemCount = citizens.value.length
      pagination.pageCount = 1
    } else {
      const result = await ListCitizens(page, pagination.pageSize)
      if (result) {
        citizens.value = result.items || []
        pagination.page = result.page
        pagination.itemCount = result.total
        pagination.pageCount = result.total_pages
      }
    }
  } catch (e: any) {
    message.error('Помилка завантаження даних: ' + e.toString())
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) {
  pagination.page = page
  loadData(page)
}

function handleSearch() {
  pagination.page = 1
  loadData(1)
}

function handleSearchUpdate(value: string) {
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }
  
  if (value === '') {
    handleSearch()
    return
  }

  if (value.length >= 3) {
    searchTimeout = window.setTimeout(() => {
      handleSearch()
    }, 500)
  }
}

function handleDelete(row: services.CitizenOutput) {
  dialog.warning({
    title: 'Видалення',
    content: `Ви впевнені, що хочете видалити громадянина ${row.full_name}?`,
    positiveText: 'Видалити',
    negativeText: 'Скасувати',
    onPositiveClick: async () => {
      try {
        await DeleteCitizen(row.id)
        message.success('Успішно видалено')
        loadData(pagination.page)
      } catch (e: any) {
        message.error('Помилка видалення: ' + e.toString())
      }
    }
  })
}

async function handleRestore(row: services.CitizenOutput) {
  try {
    await RestoreCitizen(row.id)
    message.success('Успішно відновлено')
    loadData(pagination.page)
  } catch (e: any) {
    message.error('Помилка відновлення: ' + e.toString())
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.page-container {
  padding: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.title-section {
  display: flex;
  align-items: center;
  gap: 24px;
  flex: 1;
}

.page-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  white-space: nowrap;
}

.search-input {
  max-width: 400px;
}

.table-card {
  flex: 1;
  display: flex;
  flex-direction: column;
}

:deep(.n-data-table) {
  flex: 1;
}
</style>
