<template>
  <tr class="bg-gray-800 border-b border-gray-700">
    <td class="px-6 py-4 bg-gray-800">
      <div>
        <div class="text-base font-semibold">{{ survey.name }}</div>
        <div v-if="survey.config" class="font-normal">{{ survey.config.title }}</div>
        <div class="font-normal">
          Created on: {{ formatDate(survey.created_at) }}
        </div>
      </div>
    </td>
    <td class="px-6 py-4">
      <span 
        :class="[
          'cursor-pointer text-xs font-medium px-2.5 py-0.5 rounded',
          getParseStatusColor(survey.parse_status)
        ]"
        @click="showErrorLog = !showErrorLog"
      >
        {{ survey.parse_status }}
        <Icon 
          v-if="survey.parse_status === SurveyParseStatus.Error" 
          :icon="showErrorLog ? 'heroicons:chevron-up' : 'heroicons:chevron-down'" 
          class="inline w-4 h-4 ml-1" 
        />
      </span>
      <div v-if="showErrorLog && survey.parse_status === SurveyParseStatus.Error" class="mt-2">
        <div class="p-4 text-sm text-gray-800 rounded-lg bg-gray-50 dark:bg-gray-800 dark:text-gray-300">
          <p>
            <span class="font-medium">Error log:</span>
            <br />
            <code>{{ survey.error_log }}</code>
          </p>
        </div>
      </div>
    </td>
    <td class="px-6 py-4">
      <button 
        v-if="isLaunched || canStartSurvey"
        :class="[
          'h-8 bg-red-600 hover:bg-red-700 px-2 py-0.5 rounded text-sm flex items-center gap-1 whitespace-nowrap text-white',
          { 'bg-green-600 hover:bg-green-700': !isLaunched }
        ]"
        @click="updateSurveyStatus(survey.uuid, isLaunched ? 'stopped' : 'launched')"
      >
        <Icon :icon="isLaunched ? 'heroicons:pause' : 'heroicons:play'" class="w-4 h-4" />
        <span>{{ isLaunched ? 'Stop' : 'Start' }}</span>
      </button>
      <ErrCode v-if="errorMsg" :message="errorMsg" class="w-full mt-2" />
    </td>
    <td class="px-6 py-4">
      <a 
        v-if="survey.delivery_status === SurveyDeliveryStatus.Launched"
        :href="survey.url"
        target="_blank"
        class="text-red-400 hover:text-red-300 flex items-center gap-1"
      >
        Public Link
        <Icon icon="heroicons:arrow-top-right-on-square" class="w-4 h-4" />
      </a>
    </td>
    <td class="px-6 py-4">
      <router-link 
        :to="`/app/surveys/${survey.uuid}/responses`"
        class="text-red-400 hover:text-red-300"
      >
        {{ survey.stats.sessions_count_completed }}
      </router-link>
    </td>
    <td class="px-6 py-4">{{ survey.stats.completion_rate }} %</td>
  </tr>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import moment from 'moment'
import { Icon } from '@iconify/vue'
import type { Survey } from '@/lib/types'
import { SurveyDeliveryStatus, SurveyParseStatus } from '@/lib/types'
import { updateSurvey } from '@/lib/api'
import ErrCode from '@/components/ui/ErrCode.vue'

interface Props {
  survey: Survey
}

const props = defineProps<Props>()

const errorMsg = ref<string>('')
const showErrorLog = ref<boolean>(false)

const isLaunched = computed(() => props.survey.delivery_status === SurveyDeliveryStatus.Launched)
const canStartSurvey = computed(() => props.survey.parse_status === SurveyParseStatus.Success && !isLaunched.value)

function formatDate(dateString: string): string {
  return moment(dateString).format('MMM D, YYYY')
}

function getParseStatusColor(status: string): string {
  const colors: Record<string, string> = {
    [SurveyParseStatus.Success]: 'bg-green-100 text-green-800',
    [SurveyParseStatus.Error]: 'bg-red-100 text-red-800',
    [SurveyParseStatus.Deleted]: 'bg-yellow-100 text-yellow-800',
  }
  return colors[status] || 'bg-gray-100 text-gray-800'
}

async function updateSurveyStatus(surveyUUID: string, status: string) {
  const res = await updateSurvey(surveyUUID, {
    delivery_status: status,
  })

  if (res.error) {
    errorMsg.value = res.error
  } else {
    window.location.href = '/app'
  }
}
</script>