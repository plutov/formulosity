<template>
  <div v-if="surveySession !== undefined">
    <SurveyQuestions :survey="survey" :session="surveySession" />
  </div>
  <div v-else class="intro">
    <h1 class="h1">{{ config.title }}</h1>
    <p class="intro-title" v-html="formatIntro(config.intro)"></p>
    <div class="intro-start">
      <button 
        class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        @click="startSurvey"
      >
        Start
      </button>
    </div>
    <div v-if="errMessage" class="flex flex-col py-8 px-8">
      <ErrCode :message="errMessage" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Survey, SurveyConfig, SurveySession } from '@/lib/types'
import { createSurveySession } from '@/lib/api'
import ErrCode from '@/components/ui/ErrCode.vue'
import SurveyQuestions from './SurveyQuestions.vue'

interface Props {
  survey: Survey
}

const props = defineProps<Props>()

const errMessage = ref<string | undefined>(undefined)
const surveySession = ref<SurveySession | undefined>(undefined)

const config = computed(() => props.survey.config as SurveyConfig)

function formatIntro(intro: string): string {
  return intro.replace(/(?:\r\n|\r|\n)/g, '<br>')
}

async function startSurvey() {
  errMessage.value = undefined
  const sessionRes = await createSurveySession(props.survey.url_slug)
  if (sessionRes.error) {
    errMessage.value = sessionRes.error
    return
  }

  localStorage.setItem(
    `survey_session_id:${props.survey.url_slug}`,
    sessionRes.data.data.uuid
  )
  surveySession.value = sessionRes.data.data
}
</script>