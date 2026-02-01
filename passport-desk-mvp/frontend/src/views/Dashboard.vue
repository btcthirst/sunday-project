<template>
  <n-layout class="layout" has-sider>
    <!-- Sidebar -->
    <n-layout-sider
      bordered
      :width="240"
      :collapsed-width="64"
      collapse-mode="width"
      :collapsed="collapsed"
      show-trigger
      @collapse="collapsed = true"
      @expand="collapsed = false"
    >
      <div class="logo">
        <n-icon size="32" color="#18a058">
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 3c1.66 0 3 1.34 3 3s-1.34 3-3 3-3-1.34-3-3 1.34-3 3-3zm0 14.2c-2.5 0-4.71-1.28-6-3.22.03-1.99 4-3.08 6-3.08 1.99 0 5.97 1.09 6 3.08-1.29 1.94-3.5 3.22-6 3.22z"/>
          </svg>
        </n-icon>
        <span v-if="!collapsed" class="logo-text">Паспортний стіл</span>
      </div>

      <n-menu
        :collapsed="collapsed"
        :collapsed-width="64"
        :collapsed-icon-size="22"
        :options="menuOptions"
        v-model:value="activeMenu"
      />
    </n-layout-sider>

    <!-- Main Content -->
    <n-layout>
      <!-- Header -->
      <n-layout-header class="header" bordered>
        <div class="header-left">
          <n-breadcrumb>
            <n-breadcrumb-item>Головна</n-breadcrumb-item>
            <n-breadcrumb-item>{{ currentPageTitle }}</n-breadcrumb-item>
          </n-breadcrumb>
        </div>
        <div class="header-right">
          <n-dropdown :options="userMenuOptions" trigger="click" @select="handleUserMenuSelect">
            <n-button quaternary>
              <template #icon>
                <n-icon>
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
                  </svg>
                </n-icon>
              </template>
              {{ operator?.full_name || 'Оператор' }}
            </n-button>
          </n-dropdown>
        </div>
      </n-layout-header>

      <!-- Content -->
      <n-layout-content class="content">
        <div class="content-inner">
          <!-- Dashboard Home -->
          <div v-if="activeMenu === 'dashboard'" class="dashboard-home">
            <n-h1>Ласкаво просимо!</n-h1>
            <p class="welcome-text">
              Ви увійшли як <strong>{{ operator?.full_name }}</strong>
            </p>

            <n-grid :cols="3" :x-gap="24" :y-gap="24" class="stats-grid">
              <n-gi>
                <n-card class="stat-card">
                  <n-statistic label="Громадян у базі" :value="stats.citizens">
                    <template #prefix>
                      <n-icon color="#18a058">
                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                          <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/>
                        </svg>
                      </n-icon>
                    </template>
                  </n-statistic>
                </n-card>
              </n-gi>
              <n-gi>
                <n-card class="stat-card">
                  <n-statistic label="Активних реєстрацій" :value="stats.registrations">
                    <template #prefix>
                      <n-icon color="#2080f0">
                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                          <path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7zm0 9.5c-1.38 0-2.5-1.12-2.5-2.5s1.12-2.5 2.5-2.5 2.5 1.12 2.5 2.5-1.12 2.5-2.5 2.5z"/>
                        </svg>
                      </n-icon>
                    </template>
                  </n-statistic>
                </n-card>
              </n-gi>
              <n-gi>
                <n-card class="stat-card">
                  <n-statistic label="Видано довідок" :value="stats.certificates">
                    <template #prefix>
                      <n-icon color="#f0a020">
                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                          <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/>
                        </svg>
                      </n-icon>
                    </template>
                  </n-statistic>
                </n-card>
              </n-gi>
            </n-grid>

            <n-card title="Швидкі дії" class="quick-actions">
              <n-space>
                <n-button type="primary" size="large" disabled>
                  <template #icon>
                    <n-icon>
                      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                        <path d="M15 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm-9-2V7H4v3H1v2h3v3h2v-3h3v-2H6zm9 4c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
                      </svg>
                    </n-icon>
                  </template>
                  Додати громадянина
                </n-button>
                <n-button size="large" disabled>
                  <template #icon>
                    <n-icon>
                      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                        <path d="M15.5 14h-.79l-.28-.27C15.41 12.59 16 11.11 16 9.5 16 5.91 13.09 3 9.5 3S3 5.91 3 9.5 5.91 16 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/>
                      </svg>
                    </n-icon>
                  </template>
                  Пошук
                </n-button>
                <n-button size="large" disabled>
                  <template #icon>
                    <n-icon>
                      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                        <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/>
                      </svg>
                    </n-icon>
                  </template>
                  Довідка
                </n-button>
              </n-space>
            </n-card>
          </div>

          <!-- Placeholder pages -->
          <div v-else class="placeholder-page">
            <n-empty :description="getPlaceholderText(activeMenu)">
              <template #icon>
                <n-icon size="64" color="#5c5c66">
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
                  </svg>
                </n-icon>
              </template>
              <template #extra>
                <n-text depth="3">Буде реалізовано в наступних етапах</n-text>
              </template>
            </n-empty>
          </div>
        </div>
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import { useRouter } from 'vue-router'
import { NIcon } from 'naive-ui'
import { GetCurrentOperator, Logout, UpdateActivity } from '../../wailsjs/go/main/App'

const router = useRouter()

const collapsed = ref(false)
const activeMenu = ref('dashboard')
const operator = ref<{ id: number; username: string; full_name: string } | null>(null)
const stats = ref({ citizens: 0, registrations: 0, certificates: 0 })

// Menu options
const menuOptions = [
  {
    label: 'Головна',
    key: 'dashboard',
    icon: () => h(NIcon, null, { default: () => h('svg', { xmlns: 'http://www.w3.org/2000/svg', viewBox: '0 0 24 24', fill: 'currentColor', innerHTML: '<path d="M3 13h8V3H3v10zm0 8h8v-6H3v6zm10 0h8V11h-8v10zm0-18v6h8V3h-8z"/>' }) })
  },
  {
    label: 'Громадяни',
    key: 'citizens',
    icon: () => h(NIcon, null, { default: () => h('svg', { xmlns: 'http://www.w3.org/2000/svg', viewBox: '0 0 24 24', fill: 'currentColor', innerHTML: '<path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/>' }) })
  },
  {
    label: 'Реєстрації',
    key: 'registrations',
    icon: () => h(NIcon, null, { default: () => h('svg', { xmlns: 'http://www.w3.org/2000/svg', viewBox: '0 0 24 24', fill: 'currentColor', innerHTML: '<path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7zm0 9.5c-1.38 0-2.5-1.12-2.5-2.5s1.12-2.5 2.5-2.5 2.5 1.12 2.5 2.5-1.12 2.5-2.5 2.5z"/>' }) })
  },
  {
    label: 'Звіти',
    key: 'reports',
    icon: () => h(NIcon, null, { default: () => h('svg', { xmlns: 'http://www.w3.org/2000/svg', viewBox: '0 0 24 24', fill: 'currentColor', innerHTML: '<path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/>' }) })
  },
  {
    label: 'Журнал',
    key: 'audit',
    icon: () => h(NIcon, null, { default: () => h('svg', { xmlns: 'http://www.w3.org/2000/svg', viewBox: '0 0 24 24', fill: 'currentColor', innerHTML: '<path d="M19 3h-4.18C14.4 1.84 13.3 1 12 1c-1.3 0-2.4.84-2.82 2H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-7 0c.55 0 1 .45 1 1s-.45 1-1 1-1-.45-1-1 .45-1 1-1zm2 14H7v-2h7v2zm3-4H7v-2h10v2zm0-4H7V7h10v2z"/>' }) })
  }
]

const userMenuOptions = [
  { label: 'Профіль', key: 'profile' },
  { label: 'Налаштування', key: 'settings' },
  { type: 'divider', key: 'd1' },
  { label: 'Вийти', key: 'logout' }
]

const currentPageTitle = computed(() => {
  const titles: Record<string, string> = {
    dashboard: 'Головна',
    citizens: 'Громадяни',
    registrations: 'Реєстрації',
    reports: 'Звіти',
    audit: 'Журнал операцій'
  }
  return titles[activeMenu.value] || ''
})

function getPlaceholderText(key: string): string {
  const texts: Record<string, string> = {
    citizens: 'Модуль обліку громадян',
    registrations: 'Модуль реєстрацій',
    reports: 'Модуль звітності',
    audit: 'Журнал аудиту'
  }
  return texts[key] || 'Сторінка в розробці'
}

async function handleUserMenuSelect(key: string) {
  if (key === 'logout') {
    await Logout()
    router.push('/login')
  }
}

// Track user activity
function trackActivity() {
  UpdateActivity()
}

onMounted(async () => {
  try {
    operator.value = await GetCurrentOperator()
  } catch (e) {
    console.error('Failed to get operator:', e)
    router.push('/login')
  }

  // Set up activity tracking
  document.addEventListener('click', trackActivity)
  document.addEventListener('keydown', trackActivity)
  document.addEventListener('mousemove', trackActivity)
})
</script>

<style scoped>
.layout {
  height: 100vh;
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.logo-text {
  font-size: 16px;
  font-weight: 600;
  color: #fff;
  white-space: nowrap;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 64px;
}

.header-left {
  display: flex;
  align-items: center;
}

.header-right {
  display: flex;
  align-items: center;
}

.content {
  background-color: #101014;
}

.content-inner {
  padding: 24px;
  min-height: calc(100vh - 64px);
}

.dashboard-home .welcome-text {
  color: rgba(255, 255, 255, 0.6);
  margin-bottom: 32px;
}

.stats-grid {
  margin-bottom: 24px;
}

.stat-card {
  text-align: center;
}

.quick-actions {
  margin-top: 24px;
}

.placeholder-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 400px;
}
</style>
