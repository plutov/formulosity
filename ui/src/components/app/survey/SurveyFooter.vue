<template>
  <div class="completion-message text-center py-8">
    <h2 class="h2 text-gray-300">Thank you for completing the survey!</h2>
    <p v-if="config.outro" class="mt-4 text-gray-300" v-html="formatOutro(config.outro)"></p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Survey, SurveyConfig, SurveySession } from '@/lib/types'
import { answersFor, interpolateText } from '@/lib/questions'

interface Props {
  survey: Survey
  session?: SurveySession
}

const props = defineProps<Props>()

const config = computed(() => props.survey.config as SurveyConfig)

function formatOutro(outro: string): string {
  const answers = props.session ? answersFor(props.survey, props.session) : {}
  return interpolateText(outro, answers).replace(/(?:\r\n|\r|\n)/g, '<br>')
}
</script>