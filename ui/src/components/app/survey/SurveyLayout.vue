<template>
  <main class="flex flex-col h-screen">
    <link 
      v-if="surveyTheme === SurveyThemeCustom" 
      rel="stylesheet" 
      :href="`${API_BASE_URL}/surveys/${urlSlug}/css`" 
    />
    <BgPattern />
    <slot />
  </main>
</template>

<script setup lang="ts">
import BgPattern from '@/components/ui/BgPattern.vue'
import { SurveyThemeCustom } from '@/lib/types'

interface Props {
  urlSlug?: string
  surveyTheme?: string
}

withDefaults(defineProps<Props>(), {
  urlSlug: '',
  surveyTheme: ''
})

const API_BASE_URL = import.meta.env.VITE_API_ADDR || 'http://localhost:9900'
</script>