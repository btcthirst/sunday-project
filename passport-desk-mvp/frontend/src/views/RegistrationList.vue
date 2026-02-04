<template>
  <div class="page-container">
    <div class="page-header">
      <div class="title-section">
        <h1 class="page-title">Реєстрації</h1>
        <n-input
          v-model:value="searchQuery"
          type="text"
          placeholder="Пошук за ПІБ, телефоном або ІПН"
          class="search-input"
          clearable
          @update:value="handleSearchUpdate"
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

      <n-space align="center">
        <span class="filter-label">Стан:</span>
        <n-radio-group v-model:value="filterValue" @update:value="handleFilterChange">
          <n-radio-button value="all">Всі</n-radio-button>
          <n-radio-button value="active">Зареєстровані</n-radio-button>
          <n-radio-button value="inactive">Зняті</n-radio-button>
        </n-radio-group>
      </n-space>
    </div>

    <n-card class="table-card" :bordered="false">
      <n-data-table
        :columns="columns"
        :data="registrations"
        :loading="loading"
        :pagination="pagination"
        remote
        @update:page="handlePageChange"
      />
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { NTag, NSpace, useMessage, type DataTableColumns, NInput, NIcon, NCard, NRadioGroup, NRadioButton } from 'naive-ui'
import { ListRegistrations } from '../../wailsjs/go/main/App'
import { services } from '../../wailsjs/go/models'

const router = useRouter()
const message = useMessage()

const registrations = ref<services.RegistrationListItem[]>([])
const loading = ref(false)
const searchQuery = ref('')
const filterValue = ref('active')
const activeFilter = ref<boolean | null>(true)
let searchTimeout: number | null = null

function handleFilterChange(val: string) {
  if (val === 'all') activeFilter.value = null
  else if (val === 'active') activeFilter.value = true
  else activeFilter.value = false
  loadRegistrations()
}

const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  pageCount: 1,
  showSizePicker: false,
})

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

const columns: DataTableColumns<services.RegistrationListItem> = [
  {
    title: 'ПІБ',
    key: 'citizen_name',
    fixed: 'left',
    width: 250,
  },
  {
    title: 'Дата реєстрації',
    key: 'registration_date',
    width: 140,
    render(row) {
      return formatDate(row.registration_date)
    }
  },
  {
    title: 'Тип',
    key: 'registration_type',
    width: 120,
    render(row) {
      const type = row.registration_type === 'permanent' ? 'Постійна' : 'Тимчасова'
      const color = row.registration_type === 'permanent' ? 'success' : 'warning'
      return h(NTag, { type: color, size: 'small' }, { default: () => type })
    }
  },
  {
    title: 'Адреса',
    key: 'address',
    render(row) {
      return `${row.settlement}, ${row.street} ${row.house_number}${row.apartment_number ? ', кв. ' + row.apartment_number : ''}`
    }
  },
  {
    title: 'Статус',
    key: 'is_active',
    width: 120,
    render(row) {
      return h(
        NTag,
        { type: row.is_active ? 'info' : 'error', size: 'small', bordered: false },
        { default: () => (row.is_active ? 'Активна' : 'Знято') }
      )
    }
  },
]

async function loadRegistrations() {
  loading.value = true
  try {
    const result = await ListRegistrations(
      searchQuery.value,
      activeFilter.value,
      pagination.page,
      pagination.pageSize
    )
    registrations.value = result.items
    pagination.itemCount = result.total
    pagination.pageCount = result.total_pages
  } catch (err: any) {
    message.error('Помилка завантаження реєстрацій: ' + err.message)
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) {
  pagination.page = page
  loadRegistrations()
}

function handleSearchUpdate(val: string) {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = window.setTimeout(() => {
    pagination.page = 1
    loadRegistrations()
  }, 500)
}

onMounted(() => {
  loadRegistrations()
})
</script>

<style scoped>
.page-container {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.title-section {
  display: flex;
  align-items: center;
  gap: 24px;
}

.page-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #fff;
}

.search-input {
  width: 350px;
}

.table-card {
  background-color: #18181c;
}

.filter-label {
  color: rgba(255, 255, 255, 0.6);
  font-size: 14px;
}
</style>
