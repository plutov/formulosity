<template>
  <AppLayout>
    <div v-if="loading">Loading...</div>
    <ErrCode v-else-if="errorMsg" :message="errorMsg" />
    <div v-else-if="survey" class="survey-responses">
      <h1 class="h1">Survey Responses</h1>
      <h2 class="h2">{{ survey.name }}</h2>
      <p>{{ survey.stats.sessions_count_completed }} responses</p>
      
      <div class="mt-8">
        <div v-if="sessions.length === 0" class="text-center py-8">
          <p>No responses yet.</p>
        </div>
        <div v-else class="overflow-x-auto">
          <table class="w-full text-sm text-left text-gray-500 dark:text-gray-400">
            <thead class="text-xs text-gray-700 uppercase bg-gray-50 dark:bg-gray-700 dark:text-gray-400">
              <tr>
                <th scope="col" class="px-6 py-3">Status</th>
                <th scope="col" class="px-6 py-3">Created At</th>
                <th scope="col" class="px-6 py-3">Completed At</th>
                <th scope="col" class="px-6 py-3">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="session in sessions" :key="session.uuid" class="bg-gray-800 border-b border-gray-700">
                <td class="px-6 py-4">
                  <span v-if="session.status === SurveySessionStatus.Completed" class="text-green-400 flex items-center gap-1">
                    <Icon icon="heroicons:check-circle" class="w-4 h-4" />
                    Completed
                  </span>
                  <span v-else class="text-yellow-400 flex items-center gap-1">
                    <Icon icon="heroicons:clock" class="w-4 h-4" />
                    In Progress
                  </span>
                </td>
                <td class="px-6 py-4">{{ formatDate(session.created_at) }}</td>
                <td class="px-6 py-4">{{ session.completed_at ? formatDate(session.completed_at) : '-' }}</td>
                <td class="px-6 py-4">
                  <button 
                    @click="viewSession = session"
                    class="text-blue-400 hover:text-blue-300 flex items-center gap-1"
                  >
                    <Icon icon="heroicons:eye" class="w-4 h-4" />
                    View Details
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Modal for viewing session details -->
      <div v-if="viewSession" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="viewSession = undefined">
        <div class="bg-white p-6 rounded-lg max-w-2xl max-h-96 overflow-y-auto">
          <div class="flex justify-between items-center mb-4">
            <h3 class="text-lg font-semibold">Session Details</h3>
            <button 
              @click="viewSession = undefined"
              class="text-gray-400 hover:text-gray-600"
            >
              <Icon icon="heroicons:x-mark" class="w-6 h-6" />
            </button>
          </div>
          <div class="space-y-4">
            <div v-for="answer in viewSession.question_answers" :key="answer.question_uuid">
              <div class="border-b pb-2">
                <p class="font-medium">Question: {{ getQuestionLabel(answer.question_uuid) }}</p>
                <p class="text-gray-600">Answer: {{ formatAnswer(answer.answer.value) }}</p>
              </div>
            </div>
          </div>
          <button 
            @click="viewSession = undefined"
            class="mt-4 px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600 flex items-center gap-1"
          >
            <Icon icon="heroicons:x-mark" class="w-4 h-4" />
            Close
          </button>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import moment from 'moment'
import { Icon } from '@iconify/vue'
import AppLayout from '@/components/app/AppLayout.vue'
import ErrCode from '@/components/ui/ErrCode.vue'
import { getSurveys, getSurveySessions } from '@/lib/api'
import type { Survey, SurveySession } from '@/lib/types'
import { SurveySessionStatus } from '@/lib/types'

const route = useRoute()

const survey = ref<Survey | null>(null)
const sessions = ref<SurveySession[]>([])
const errorMsg = ref<string>('')
const loading = ref<boolean>(true)
const viewSession = ref<SurveySession | undefined>(undefined)

onMounted(async () => {
  const surveyUuid = route.params.surveyUuid as string
  if (!surveyUuid) return

  // First get the survey details
  const surveysResp = await getSurveys()
  if (surveysResp.error) {
    errorMsg.value = 'Failed to fetch surveys'
    loading.value = false
    return
  }

  const foundSurvey = surveysResp.data.data.find((s: Survey) => s.uuid === surveyUuid)
  if (!foundSurvey) {
    errorMsg.value = 'Survey not found'
    loading.value = false
    return
  }

  survey.value = foundSurvey

  // Then get the sessions
  const sessionsResp = await getSurveySessions(surveyUuid, '')
  if (sessionsResp.error) {
    errorMsg.value = 'Failed to fetch survey sessions'
  } else {
    sessions.value = sessionsResp.data.data || []
  }
  loading.value = false
})

function formatDate(dateString: string): string {
  return moment(dateString).format('MMM D, YYYY HH:mm')
}

function getQuestionLabel(questionUuid: string): string {
  if (!survey.value) return 'Unknown question'
  const question = survey.value.config.questions.questions.find(q => q.uuid === questionUuid)
  return question ? question.label : 'Unknown question'
}

function formatAnswer(value: string | string[] | number | boolean): string {
  if (Array.isArray(value)) {
    return value.join(', ')
  }
  return String(value)
}
</script>