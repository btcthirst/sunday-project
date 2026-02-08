<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">Журнал аудиту</h1>
      <n-space>
        <n-button @click="loadLogs">Оновити</n-button>
      </n-space>
    </div>

    <n-card class="table-card" :bordered="false">
      <n-data-table
        :columns="columns"
        :data="logs"
        :loading="loading"
        :pagination="pagination"
      />
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { NTag, useMessage, type DataTableColumns } from 'naive-ui'
import { GetAuditLogs } from '../../wailsjs/go/main/App'
import { models } from '../../wailsjs/go/models'

const message = useMessage()
const logs = ref<models.AuditLogOutput[]>([])
const loading = ref(false)

const pagination = {
  pageSize: 20
}

const columns: DataTableColumns<models.AuditLogOutput> = [
  {
    title: 'Час',
    key: 'timestamp',
    width: 200,
    render(row) {
      return formatDate(row.timestamp)
    }
  },
  {
    title: 'Оператор',
    key: 'operator_name',
    width: 180
  },
  {
    title: 'Дія',
    key: 'action_type',
    width: 120,
    render(row) {
      const type = row.action_type
      let color: 'success' | 'info' | 'warning' | 'error' | 'default' = 'default'
      
      switch(type) {
        case 'CREATE': color = 'success'; break;
        case 'UPDATE': color = 'warning'; break;
        case 'DELETE': color = 'error'; break;
        case 'READ': color = 'info'; break;
      }
      
      return h(NTag, { type: color, bordered: false, size: 'small' }, { default: () => type })
    }
  },
  {
    title: 'Таблиця',
    key: 'table_name',
    width: 150
  },
  {
    title: 'Опис',
    key: 'description'
  }
]

function formatDate(dateStr: string) {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleString('uk-UA')
}

async function loadLogs() {
  loading.value = true
  try {
    const result = await GetAuditLogs(500) // Show last 500 logs
    logs.value = result || []
  } catch (e: any) {
    message.error('Помилка завантаження логів: ' + e.toString())
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadLogs()
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

.page-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}

.table-card {
  flex: 1;
}
</style>
