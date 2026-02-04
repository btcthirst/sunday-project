<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">Звіти та Довідки</h1>
    </div>

    <n-card :bordered="false">
      <n-tabs type="line" animated v-model:value="activeTab">
        
        <!-- Tab 1: Family Composition -->
        <n-tab-pane name="family" tab="Склад сім'ї">
          <n-space vertical size="large">
            <n-alert title="Формування довідки про склад сім'ї" type="info">
              Виберіть головного громадянина, налаштуйте склад його сім'ї та виберіть членів для включення в довідку.
            </n-alert>

            <n-grid :cols="2" :x-gap="24">
              <n-gi>
                <n-form-item label="Пошук громадянина (ПІБ)">
                  <n-select
                    v-model:value="primaryCitizenID"
                    filterable
                    remote
                    :options="citizenOptions"
                    :loading="searchLoading"
                    placeholder="Введіть прізвище..."
                    @search="handleCitizenSearch"
                    @update:value="loadFamilyMembers"
                    clearable
                  />
                </n-form-item>
              </n-gi>
              <n-gi v-if="primaryCitizen">
                <n-card title="Дані громадянина" size="small">
                  <n-descriptions :column="1" bordered label-placement="left">
                    <n-descriptions-item label="ПІБ">{{ primaryCitizen.full_name }}</n-descriptions-item>
                    <n-descriptions-item label="Паспорт">{{ primaryCitizen.passport_masked }}</n-descriptions-item>
                    <n-descriptions-item label="ІПН">{{ primaryCitizen.tax_number_masked }}</n-descriptions-item>
                  </n-descriptions>
                </n-card>
              </n-gi>
            </n-grid>

            <div v-if="primaryCitizenID">
              <n-h3>Члени сім'ї</n-h3>
              <n-data-table
                :columns="familyColumns"
                :data="familyMembers"
                :loading="familyLoading"
              />
              <n-button @click="showAddMemberModal = true" style="margin-top: 12px">
                Додати члена сім'ї
              </n-button>
            </div>

            <n-divider />
            
            <div v-if="primaryCitizenID">
               <n-form-item v-if="familyMembers.length > 0" label="Виберіть кого включити в довідку">
                 <n-checkbox-group v-model:value="membersToPrint">
                    <n-space>
                      <!-- We include the primary citizen as well if they are in membersToPrint -->
                      <n-checkbox :value="primaryCitizenID" :label="primaryCitizen?.full_name + ' (Головний)' " />
                      <n-checkbox v-for="m in familyMembers" :key="m.id" :value="m.id" :label="m.full_name" />
                    </n-space>
                 </n-checkbox-group>
               </n-form-item>
               
               <n-p v-else>Громадянин проживає один (членів сім'ї не знайдено).</n-p>

               <n-divider title-placement="left">Деталі для довідки</n-divider>
               <n-grid :cols="2" :x-gap="12">
                 <n-gi>
                   <n-form-item label="Назва установи (видавець)">
                     <n-input v-model:value="certOptions.issuer_name" placeholder="напр. Житомирська міська рада" />
                   </n-form-item>
                 </n-gi>
                 <n-gi>
                   <n-form-item label="Місто">
                     <n-input v-model:value="certOptions.city" placeholder="напр. м. Житомир" />
                   </n-form-item>
                 </n-gi>
                 <n-gi>
                   <n-form-item label="Куди подається (установа)">
                     <n-input v-model:value="certOptions.target_institution" placeholder="напр. управління соцзахисту" />
                   </n-form-item>
                 </n-gi>
                 <n-gi>
                   <n-form-item label="Мета">
                     <n-input v-model:value="certOptions.purpose" placeholder="напр. призначення субсидії" />
                   </n-form-item>
                 </n-gi>
                 <n-gi>
                   <n-form-item label="Посада підписанта">
                     <n-input v-model:value="certOptions.signatory_title" placeholder="напр. Міський голова" />
                   </n-form-item>
                 </n-gi>
                 <n-gi>
                   <n-form-item label="ПІБ підписанта">
                     <n-input v-model:value="certOptions.signatory_name" placeholder="напр. В.О. Іванов" />
                   </n-form-item>
                 </n-gi>
               </n-grid>

               <n-button type="primary" :loading="generating" @click="generateFamilyCertificate" style="margin-top: 12px">
                 Сформувати PDF
               </n-button>
            </div>
          </n-space>
        </n-tab-pane>

        <!-- Tab 2: General List -->
        <n-tab-pane name="list" tab="Загальний список">
          <n-space vertical size="large">
            <n-grid :cols="4" :x-gap="12">
               <n-gi :span="2">
                 <n-input v-model:value="listSearchQuery" placeholder="Пошук..." @input="handleListSearch" />
               </n-gi>
               <n-gi>
                 <n-select v-model:value="statusFilter" :options="statusOptions" @update:value="loadGeneralList" />
               </n-gi>
               <n-gi>
                 <n-popover trigger="click" placement="bottom-end">
                    <template #trigger>
                      <n-button>Стовпчики</n-button>
                    </template>
                    <n-checkbox-group v-model:value="visibleColumns">
                      <n-space vertical>
                        <n-checkbox value="birth_date" label="Дата народження" />
                        <n-checkbox value="passport" label="Паспорт" />
                        <n-checkbox value="tax_number" label="ІПН" />
                        <n-checkbox value="phone" label="Телефон" />
                        <n-checkbox value="address" label="Адреса" />
                      </n-space>
                    </n-checkbox-group>
                 </n-popover>
               </n-gi>
               <n-gi>
                 <n-button :disabled="selectedRowKeys.length === 0" @click="handleExportSelected">
                   Експортувати обрані ({{ selectedRowKeys.length }})
                 </n-button>
               </n-gi>
            </n-grid>

            <n-data-table
              remote
              :columns="generalListColumns"
              :data="generalListData"
              :loading="listLoading"
              :pagination="pagination"
              :row-key="row => row.id"
              v-model:checked-row-keys="selectedRowKeys"
              @update:page="handlePageChange"
            />
          </n-space>
        </n-tab-pane>

        <!-- Tab 3: Export Registrations (Moved from Reports) -->
        <n-tab-pane name="registrations" tab="Реєстрації (Excel)">
          <n-space vertical size="large">
            <n-alert type="info" :bordered="false">
              Формування Excel-файлу зі списком всіх реєстрацій за обраний період.
            </n-alert>

            <n-form-item label="Період реєстрації">
              <n-date-picker
                v-model:formatted-value="dateRange"
                value-format="yyyy-MM-dd"
                type="daterange"
                clearable
                start-placeholder="З"
                end-placeholder="По"
                style="max-width: 400px"
              />
            </n-form-item>

            <n-button type="primary" size="large" @click="handleExportRegistrations" :loading="exportingRegs">
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
        </n-tab-pane>

      </n-tabs>
    </n-card>

    <!-- Modal for adding family member -->
    <n-modal v-model:show="showAddMemberModal" preset="dialog" title="Додати члена сім'ї">
       <n-form-item label="Громадянин">
         <n-select
           v-model:value="newMemberID"
           filterable
           remote
           :options="citizenOptions"
           :loading="searchLoading"
           placeholder="Пошук..."
           @search="handleCitizenSearch"
         />
       </n-form-item>
       <n-form-item label="Тип зв'язку">
         <n-input v-model:value="newMemberRelation" placeholder="напр. дружина, син..." />
       </n-form-item>
       <template #action>
         <n-button @click="showAddMemberModal = false">Скасувати</n-button>
         <n-button type="primary" :disabled="!newMemberID" @click="addFamilyMember">Додати</n-button>
       </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, h, reactive, computed, watch } from 'vue'
import { useMessage, NButton, NSpace, NTag } from 'naive-ui'
import { 
  SearchCitizens, ListCitizens,
  GetFamilyMembers, AddFamilyMember, RemoveFamilyMember,
  ImportCitizens, ExportCitizensToExcel, GenerateFamilyStatusCertificate,
  GetCitizen, ExportCustomCitizensToExcel, ExportCitizens
} from '../../wailsjs/go/main/App'

const message = useMessage()
const activeTab = ref('family')

// --- Family Tab ---
const primaryCitizenID = ref<number | null>(null)
const primaryCitizen = ref<any>(null)
const familyMembers = ref<any[]>([])
const familyLoading = ref(false)
const searchLoading = ref(false)
const citizenOptions = ref<any[]>([])
const showAddMemberModal = ref(false)
const newMemberID = ref<number | null>(null)
const newMemberRelation = ref('')
const membersToPrint = ref<number[]>([])
const generating = ref(false)

const certOptions = reactive({
  issuer_name: '',
  city: '',
  target_institution: '',
  purpose: '',
  signatory_title: '',
  signatory_name: ''
})

const familyColumns = [
  { title: 'ПІБ', key: 'full_name' },
  { title: 'Зв\'язок', key: 'relation_type' },
  { title: 'Паспорт', key: 'passport_masked' },
  {
    title: 'Дії',
    key: 'actions',
    render(row: any) {
      return h(NButton, {
        size: 'small',
        type: 'error',
        ghost: true,
        onClick: () => removeMember(row.id)
      }, { default: () => 'Видалити' })
    }
  }
]

async function handleCitizenSearch(query: string) {
  if (query.length < 2) return
  searchLoading.value = true
  try {
    const res = await SearchCitizens(query, 'name')
    citizenOptions.value = res.map((c: any) => ({
      label: `${c.full_name} (${c.birth_date})`,
      value: c.id,
      data: c
    }))
  } finally {
    searchLoading.value = false
  }
}

async function loadFamilyMembers(id: number) {
  if (!id) {
    primaryCitizen.value = null
    familyMembers.value = []
    return
  }
  familyLoading.value = true
  try {
    primaryCitizen.value = await GetCitizen(id)
    familyMembers.value = await GetFamilyMembers(id)
    membersToPrint.value = familyMembers.value.map(m => m.id)
    if (id && !membersToPrint.value.includes(id)) {
      membersToPrint.value.push(id)
    }
  } finally {
    familyLoading.value = false
  }
}

async function addFamilyMember() {
  if (!primaryCitizenID.value || !newMemberID.value) return
  try {
    await AddFamilyMember(primaryCitizenID.value, newMemberID.value, newMemberRelation.value)
    message.success('Додано')
    showAddMemberModal.value = false
    loadFamilyMembers(primaryCitizenID.value)
  } catch (e: any) {
    message.error(e.toString())
  }
}

async function removeMember(memberID: number) {
  if (!primaryCitizenID.value) return
  try {
    await RemoveFamilyMember(primaryCitizenID.value, memberID)
    loadFamilyMembers(primaryCitizenID.value)
  } catch (e: any) {
    message.error(e.toString())
  }
}

watch(primaryCitizenID, (newID) => {
  if (newID) {
    // When primary citizen changes, ensure they are in the printed members
    if (!membersToPrint.value.includes(newID)) {
      membersToPrint.value.push(newID)
    }
  } else {
    membersToPrint.value = []
  }
})

async function generateFamilyCertificate() {
  if (membersToPrint.value.length === 0) {
    message.warning('Виберіть хоча б одного громадянина')
    return
  }
  generating.value = true
  try {
    const res = await GenerateFamilyStatusCertificate({
      citizen_ids: membersToPrint.value,
      ...certOptions
    })
    if (res !== 'cancelled') message.success('Довідку збережено: ' + res)
  } catch (e: any) {
    message.error(e.toString())
  } finally {
    generating.value = false
  }
}

// --- General List Tab ---
const listSearchQuery = ref('')
const statusFilter = ref('active')
const statusOptions = [
  { label: 'Активні', value: 'active' },
  { label: 'Видалені', value: 'deleted' }
]
const visibleColumns = ref(['birth_date', 'passport', 'address'])
const generalListData = ref<any[]>([])
const listLoading = ref(false)
const pagination = reactive({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  pageCount: 0
})
const selectedRowKeys = ref<number[]>([])

// --- Registration Export (from Reports.vue) ---
const dateRange = ref<[string, string] | null>(null)
const exportingRegs = ref(false)

async function handleExportRegistrations() {
  exportingRegs.value = true
  try {
    const from = dateRange.value ? dateRange.value[0] : ''
    const to = dateRange.value ? dateRange.value[1] : ''
    const savedPath = await ExportCitizens(from, to)
    if (savedPath === 'cancelled') {
      message.info('Експорт скасовано')
    } else {
      message.success('Звіт збережено: ' + savedPath)
    }
  } catch (err: any) {
    message.error('Помилка формування звіту: ' + err.message)
  } finally {
    exportingRegs.value = false
  }
}

const generalListColumns = computed(() => {
  const cols: any[] = [
    { type: 'selection' },
    { title: 'ПІБ', key: 'full_name', sortable: true }
  ]
  if (visibleColumns.value.includes('birth_date')) cols.push({ title: 'Дата народж.', key: 'birth_date' })
  if (visibleColumns.value.includes('passport')) cols.push({ title: 'Паспорт', key: 'passport_masked' })
  if (visibleColumns.value.includes('tax_number')) cols.push({ title: 'ІПН', key: 'tax_number_masked' })
  if (visibleColumns.value.includes('phone')) cols.push({ title: 'Телефон', key: 'phone' })
  if (visibleColumns.value.includes('address')) cols.push({ title: 'Адреса', key: 'active_address' })
  return cols
})

async function loadGeneralList() {
  listLoading.value = true
  try {
    if (listSearchQuery.value) {
      const results = await SearchCitizens(listSearchQuery.value, 'name')
      generalListData.value = results.filter((c: any) => 
        statusFilter.value === 'deleted' ? c.deleted : !c.deleted
      )
      pagination.itemCount = generalListData.value.length
      pagination.pageCount = 1
    } else {
      const res = await ListCitizens(pagination.page, pagination.pageSize, statusFilter.value === 'deleted')
      generalListData.value = res.items || []
      pagination.itemCount = res.total
      pagination.pageCount = res.total_pages
    }
  } finally {
    listLoading.value = false
  }
}

function handlePageChange(page: number) {
  pagination.page = page
  loadGeneralList()
}

function handleListSearch() {
  pagination.page = 1
  loadGeneralList()
}


// --- Import/Export (Custom Export remains)
async function handleExportSelected() {
  if (selectedRowKeys.value.length === 0) return
  try {
    const res = await ExportCustomCitizensToExcel(selectedRowKeys.value, visibleColumns.value)
    if (res !== 'cancelled') message.success('Експортовано: ' + res)
  } catch (e: any) {
    message.error(e.toString())
  }
}

onMounted(() => {
  loadGeneralList()
})
</script>

<style scoped>
.page-container {
  padding: 24px;
}
.page-header {
  margin-bottom: 24px;
}
.page-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #fff;
}
.manual-content {
  color: rgba(255, 255, 255, 0.82);
}
</style>
