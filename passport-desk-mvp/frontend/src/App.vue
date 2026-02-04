<template>
  <n-config-provider :theme="darkTheme" :locale="ukUA" :date-locale="dateUkUA">
    <n-message-provider>
      <n-dialog-provider>
        <n-notification-provider>
          <router-view />
        </n-notification-provider>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { darkTheme, ukUA, dateUkUA } from 'naive-ui'
import { EventsOn } from '../wailsjs/runtime'

const router = useRouter()

onMounted(() => {
  EventsOn('session-locked', () => {
    router.push({ name: 'Login' })
  })
})
</script>

<style>
html, body {
  margin: 0;
  padding: 0;
  height: 100%;
  background-color: #18181c;
}

#app {
  height: 100%;
}
</style>
