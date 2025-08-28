<template>
  <SurveyLayout 
    :surveyTheme="survey?.config?.theme" 
    :urlSlug="survey?.url_slug"
  >
    <div v-if="loading">Loading...</div>
    <SurveyNotFound v-else-if="notFound || !survey" />
    <SurveyForm v-else :survey="survey" />
  </SurveyLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getSurvey } from '@/lib/api'
import type { Survey } from '@/lib/types'
import SurveyLayout from '@/components/app/survey/SurveyLayout.vue'
import SurveyNotFound from '@/components/app/survey/SurveyNotFound.vue'
import SurveyForm from '@/components/app/survey/SurveyForm.vue'

const route = useRoute()

const survey = ref<Survey | null>(null)
const loading = ref<boolean>(true)
const notFound = ref<boolean>(false)

onMounted(async () => {
  const urlSlug = route.params.urlSlug as string
  if (!urlSlug) return

  const surveyResp = await getSurvey(urlSlug)

  if (
    surveyResp.error ||
    !surveyResp.data.data ||
    !surveyResp.data.data.config
  ) {
    notFound.value = true
  } else {
    survey.value = surveyResp.data.data as Survey
  }
  loading.value = false
})
</script>