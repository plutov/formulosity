<template>
  <AppLayout>
    <div v-if="loading">Loading...</div>
    <ErrCode v-else-if="errorMsg" :message="errorMsg" />
    <div v-else-if="survey" class="survey-responses">
      <h1 class="text-2xl font-bold mb-2">Survey Responses</h1>
      <h2 class="text-lg text-gray-600 mb-4">{{ survey.name }}</h2>
      <p class="text-gray-500 mb-6">{{ survey.stats.sessions_count_completed }} responses</p>
      
      <!-- Export button -->
      <div class="mb-6">
        <button 
          @click="exportResponses"
          :disabled="downloading"
          class="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
        >
          <Icon icon="heroicons:arrow-down-tray" class="w-4 h-4" />
          {{ downloading ? 'Exporting...' : 'Export as JSON' }}
        </button>
      </div>

      <div class="mt-8">
        <div v-if="sessions.length === 0" class="text-center py-8">
          <p>No responses yet.</p>
        </div>
        <div v-else class="overflow-x-auto">
          <table class="w-full text-sm text-left text-gray-500">
            <thead class="text-xs text-gray-700 uppercase bg-gray-50">
              <tr>
                <th 
                  v-for="col in columns" 
                  :key="col.key"
                  scope="col" 
                  class="p-4 cursor-pointer hover:bg-gray-100"
                  @click="sortBy(col.key)"
                >
                  <div class="flex items-center gap-1">
                    <Icon 
                      v-if="sortByField === col.key" 
                      :icon="sortOrder === 'asc' ? 'heroicons:chevron-up' : 'heroicons:chevron-down'" 
                      class="w-3 h-3" 
                    />
                    {{ col.label }}
                  </div>
                </th>
                <th scope="col" class="p-4">Webhook Status</th>
                <th scope="col" class="p-4">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr 
                v-for="session in sessions" 
                :key="session.uuid" 
                class="group border-b hover:bg-gray-50"
              >
                <td class="p-4">{{ session.uuid }}</td>
                <td class="p-4">
                  <span 
                    :class="[
                      'inline-flex items-center gap-1 px-2 py-1 text-xs font-medium rounded-full',
                      session.status === SurveySessionStatus.Completed 
                        ? 'bg-green-100 text-green-700' 
                        : 'bg-yellow-100 text-yellow-700'
                    ]"
                  >
                    <Icon 
                      :icon="session.status === SurveySessionStatus.Completed ? 'heroicons:check-circle' : 'heroicons:clock'" 
                      class="w-3 h-3" 
                    />
                    {{ session.status === SurveySessionStatus.Completed ? 'Completed' : 'In Progress' }}
                  </span>
                </td>
                <td class="p-4">{{ formatDate(session.created_at) }}</td>
                <td class="p-4">
                  {{ session.completed_at ? formatDate(session.completed_at) : '-' }}
                </td>
                <td class="p-4">{{ session.webhookData.statusCode || '-' }}</td>
                <td class="p-4">
                  <div class="flex items-center gap-2">
                    <button 
                      @click="viewSession = session"
                      class="inline-flex items-center gap-1 px-2 py-1 text-xs bg-blue-100 text-blue-700 rounded hover:bg-blue-200"
                    >
                      <Icon icon="heroicons:eye" class="w-3 h-3" />
                      View
                    </button>
                    <button 
                      @click="deleteSession(session)"
                      class="inline-flex items-center gap-1 px-2 py-1 text-xs bg-gray-100 text-gray-700 rounded hover:bg-gray-200"
                    >
                      <Icon icon="heroicons:trash" class="w-3 h-3" />
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div v-if="totalPages > 1" class="mt-6 flex justify-center">
          <div class="flex items-center gap-1">
            <button 
              v-for="page in totalPages" 
              :key="page"
              @click="changePage(page)"
              :class="[
                'px-3 py-1 text-sm rounded',
                page === currentPage 
                  ? 'bg-blue-600 text-white' 
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              ]"
            >
              {{ page }}
            </button>
          </div>
        </div>
      </div>

      <!-- Modal for viewing session details -->
      <div v-if="viewSession" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="viewSession = undefined">
        <div class="bg-white p-6 rounded-lg max-w-4xl max-h-96 overflow-y-auto">
          <div class="flex justify-between items-center mb-4">
            <h3 class="text-lg font-semibold text-gray-900">Response: {{ viewSession.uuid }}</h3>
            <button 
              @click="viewSession = undefined"
              class="text-gray-400 hover:text-gray-600"
            >
              <Icon icon="heroicons:x-mark" class="w-6 h-6" />
            </button>
          </div>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead class="text-xs text-gray-700 uppercase bg-gray-50">
                <tr>
                  <th class="p-3 text-left">Question ID</th>
                  <th class="p-3 text-left">Question</th>
                  <th class="p-3 text-left">Response</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="answer in viewSession.question_answers" :key="answer.question_uuid" class="border-b">
                  <td class="p-3 text-gray-900">{{ answer.question_id }}</td>
                  <td class="p-3 text-gray-900">{{ getQuestionLabel(answer.question_uuid) }}</td>
                  <td class="p-3 text-gray-900">
                    <div v-if="isFileAnswer(answer.question_uuid)">
                      <button 
                        @click="downloadFile(answer.answer.value as string)"
                        class="inline-flex items-center gap-1 px-2 py-1 bg-blue-100 text-blue-700 rounded text-xs hover:bg-blue-200"
                      >
                        <Icon icon="heroicons:arrow-down-tray" class="w-3 h-3" />
                        Download
                      </button>
                    </div>
                    <div v-else>{{ formatAnswer(answer.answer.value) }}</div>
                  </td>
                </tr>
              </tbody>
            </table>
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
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import moment from 'moment'
import { Icon } from '@iconify/vue'
import AppLayout from '@/components/app/AppLayout.vue'
import ErrCode from '@/components/ui/ErrCode.vue'
import { getSurveys, getSurveySessions, deleteSurveySession, download } from '@/lib/api'
import type { Survey, SurveySession } from '@/lib/types'
import { SurveySessionStatus, SurveySessionsLimit, SurveyQuestionType } from '@/lib/types'

const route = useRoute()

const survey = ref<Survey | null>(null)
const sessions = ref<SurveySession[]>([])
const errorMsg = ref<string>('')
const loading = ref<boolean>(true)
const downloading = ref<boolean>(false)
const viewSession = ref<SurveySession | undefined>(undefined)

// Pagination and sorting
const currentPage = ref<number>(1)
const sortByField = ref<string>('created_at')
const sortOrder = ref<string>('desc')

const columns = [
  { label: 'Session ID', key: 'uuid' },
  { label: 'Status', key: 'status' },
  { label: 'Started at', key: 'created_at' },
  { label: 'Completed at', key: 'completed_at' }
]

const totalPages = computed(() => {
  if (!survey.value) return 1
  return survey.value.pages_count || 1
})

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
  await fetchResponses(currentPage.value, sortByField.value, sortOrder.value)
  loading.value = false
})

async function fetchResponses(page: number, sortBy: string, order: string) {
  if (!survey.value) return
  
  errorMsg.value = ''
  const limit = SurveySessionsLimit
  const offset = (page - 1) * limit
  
  const sessionsResp = await getSurveySessions(
    survey.value.uuid,
    `limit=${limit}&offset=${offset}&sort_by=${sortBy}&order=${order}`
  )
  
  if (sessionsResp.error) {
    errorMsg.value = 'Unable to load survey sessions'
  } else {
    sessions.value = sessionsResp.data.data.sessions || []
    // Update pages count from the response
    if (sessionsResp.data.data.pages_count) {
      survey.value.pages_count = sessionsResp.data.data.pages_count
    }
  }
}

async function changePage(page: number) {
  currentPage.value = page
  await fetchResponses(page, sortByField.value, sortOrder.value)
}

async function sortBy(field: string) {
  let newOrder = 'asc'
  if (sortByField.value === field) {
    newOrder = sortOrder.value === 'asc' ? 'desc' : 'asc'
  }
  
  sortByField.value = field
  sortOrder.value = newOrder
  currentPage.value = 1
  
  await fetchResponses(currentPage.value, field, newOrder)
}

async function exportResponses() {
  if (!survey.value) return
  
  downloading.value = true
  errorMsg.value = ''
  
  const allSessionsResp = await getSurveySessions(
    survey.value.uuid,
    'limit=1000000&offset=0&sort_by=created_at&order=desc'
  )
  
  downloading.value = false
  
  if (allSessionsResp.error) {
    errorMsg.value = 'Unable to export survey sessions'
    return
  }
  
  const element = document.createElement('a')
  const file = new Blob(
    [JSON.stringify(allSessionsResp.data.data.sessions, null, 2)],
    { type: 'application/json' }
  )
  element.href = URL.createObjectURL(file)
  element.download = 'survey_responses.json'
  document.body.appendChild(element)
  element.click()
  document.body.removeChild(element)
  URL.revokeObjectURL(element.href)
}

async function deleteSession(session: SurveySession) {
  if (!survey.value) return
  
  errorMsg.value = ''
  
  const deleteResp = await deleteSurveySession(survey.value.uuid, session.uuid)
  
  if (deleteResp.error) {
    errorMsg.value = 'Unable to delete survey session'
  } else {
    await fetchResponses(1, sortByField.value, sortOrder.value)
    currentPage.value = 1
  }
}

function formatDate(dateString: string): string {
  return moment(dateString).format('MMM D, YYYY HH:mm')
}

function getQuestionLabel(questionUuid: string): string {
  if (!survey.value?.config?.questions?.questions) return 'Unknown question'
  const question = survey.value.config.questions.questions.find(q => q.uuid === questionUuid)
  return question ? question.label : 'Unknown question'
}

function isFileAnswer(questionUuid: string): boolean {
  if (!survey.value?.config?.questions?.questions) return false
  const question = survey.value.config.questions.questions.find(q => q.uuid === questionUuid)
  return question?.type === SurveyQuestionType.File
}

function formatAnswer(value: string | string[] | number | boolean): string {
  if (Array.isArray(value)) {
    return value.join(', ')
  }
  if (typeof value === 'boolean') {
    return value ? 'Yes' : 'No'
  }
  return String(value)
}

async function downloadFile(filePath: string) {
  if (!survey.value) return
  const fileName = filePath.substring(filePath.lastIndexOf('/') + 1)
  await download(survey.value.uuid, fileName)
}
</script>