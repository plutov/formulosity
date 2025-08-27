<template>
  <div v-if="isLoading" class="mx-auto text-center py-16">
    <div role="status">
      <Icon icon="svg-spinners:180-ring-with-bg" class="w-8 h-8 text-blue-600 mx-auto" />
    </div>
  </div>
  <div v-else-if="isNewSession">
    <SurveyIntro :survey="survey" />
  </div>
  <div v-else>
    <SurveyQuestions :survey="survey" :session="surveySession!" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import type { Survey, SurveySession } from '@/lib/types'
import { getSurveySession } from '@/lib/api'
import SurveyIntro from './SurveyIntro.vue'
import SurveyQuestions from './SurveyQuestions.vue'

interface Props {
  survey: Survey
}

const props = defineProps<Props>()

const surveySession = ref<SurveySession | undefined>(undefined)
const isNewSession = ref<boolean>(true)
const isLoading = ref<boolean>(true)

onMounted(async () => {
  if (typeof window !== 'undefined') {
    const lsValue = localStorage.getItem(
      `survey_session_id:${props.survey.url_slug}`
    )

    if (!lsValue) {
      isNewSession.value = true
      isLoading.value = false
      return
    }

    const sessionRes = await getSurveySession(props.survey.url_slug, lsValue)
    if (sessionRes.error || !sessionRes.data.data) {
      localStorage.removeItem(`survey_session_id:${props.survey.url_slug}`)
      isNewSession.value = true
      isLoading.value = false
      return
    }

    surveySession.value = sessionRes.data.data
    isNewSession.value = false
    isLoading.value = false
  }
})
</script>