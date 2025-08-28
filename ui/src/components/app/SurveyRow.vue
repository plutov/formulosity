<template>
  <tr class="group border-b hover:bg-gray-50">
    <td class="p-4">
      <div class="space-y-1">
        <div class="font-medium text-white group-hover:text-gray-900">{{ survey.name }}</div>
        <div v-if="survey.config" class="text-sm text-gray-600">{{ survey.config.title }}</div>
        <div class="text-xs text-gray-500">{{ formatDate(survey.created_at) }}</div>
      </div>
    </td>
    <td class="p-4">
      <span 
        :class="[
          'inline-flex items-center gap-1 px-2 py-1 text-xs font-medium rounded-full cursor-pointer',
          getParseStatusColor(survey.parse_status)
        ]"
        @click="showErrorLog = !showErrorLog"
      >
        {{ survey.parse_status }}
        <Icon 
          v-if="survey.parse_status === SurveyParseStatus.Error" 
          :icon="showErrorLog ? 'heroicons:chevron-up' : 'heroicons:chevron-down'" 
          class="w-3 h-3" 
        />
      </span>
      <div v-if="showErrorLog && survey.parse_status === SurveyParseStatus.Error" class="mt-3">
        <div class="p-3 text-xs bg-gray-100 rounded-lg">
          <div class="font-medium mb-1">Error Details:</div>
          <code class="text-gray-700">{{ survey.error_log }}</code>
        </div>
      </div>
    </td>
    <td class="p-4">
      <button 
        v-if="isLaunched || canStartSurvey"
        :class="[
          'inline-flex items-center gap-1 px-3 py-1.5 text-sm font-medium rounded-lg transition-colors',
          isLaunched 
            ? 'bg-gray-100 text-gray-700 hover:bg-gray-200' 
            : 'bg-blue-100 text-blue-700 hover:bg-blue-200'
        ]"
        @click="updateSurveyStatus(survey.uuid, isLaunched ? 'stopped' : 'launched')"
      >
        <Icon :icon="isLaunched ? 'heroicons:pause' : 'heroicons:play'" class="w-4 h-4" />
        {{ isLaunched ? 'Stop' : 'Start' }}
      </button>
      <ErrCode v-if="errorMsg" :message="errorMsg" class="mt-2" />
    </td>
    <td class="p-4">
      <a 
        v-if="survey.delivery_status === SurveyDeliveryStatus.Launched"
        :href="survey.url"
        target="_blank"
        class="inline-flex items-center gap-1 text-blue-600 hover:text-blue-800 text-sm font-medium"
      >
        Public Link
        <Icon icon="heroicons:arrow-top-right-on-square" class="w-4 h-4" />
      </a>
    </td>
    <td class="p-4">
      <router-link 
        :to="`/app/surveys/${survey.uuid}/responses`"
        class="text-blue-600 hover:text-blue-800 font-medium"
      >
        {{ survey.stats.sessions_count_completed }}
      </router-link>
    </td>
    <td class="p-4 text-gray-900">{{ survey.stats.completion_rate }}%</td>
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
    [SurveyParseStatus.Success]: 'bg-green-100 text-green-700',
    [SurveyParseStatus.Error]: 'bg-orange-100 text-orange-700',
    [SurveyParseStatus.Deleted]: 'bg-gray-100 text-gray-700',
  }
  return colors[status] || 'bg-gray-100 text-gray-700'
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
