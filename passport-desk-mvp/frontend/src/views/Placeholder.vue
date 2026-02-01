<template>
  <div class="placeholder-page">
    <n-empty :description="description">
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
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const description = computed(() => {
  const texts: Record<string, string> = {
    registrations: 'Модуль реєстрацій',
    reports: 'Модуль звітності',
    audit: 'Журнал аудиту'
  }
  const key = route.path.split('/').pop() || ''
  return texts[key] || 'Сторінка в розробці'
})
</script>

<style scoped>
.placeholder-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 400px;
  height: 100%;
}
</style>
