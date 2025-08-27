<template>
  <div class="completion-message text-center py-8">
    <h2 class="h2">Thank you for completing the survey!</h2>
    <p v-if="config.outro" class="mt-4" v-html="formatOutro(config.outro)"></p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Survey, SurveyConfig } from '@/lib/types'

interface Props {
  survey: Survey
}

const props = defineProps<Props>()

const config = computed(() => props.survey.config as SurveyConfig)

function formatOutro(outro: string): string {
  return outro.replace(/(?:\r\n|\r|\n)/g, '<br>')
}
</script>