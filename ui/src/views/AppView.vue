<template>
  <AppLayout>
    <div v-if="loading">Loading...</div>
    <ErrCode v-else-if="errMsg" :message="errMsg" />
    <SurveysPage v-else :surveys="surveys" />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AppLayout from '@/components/app/AppLayout.vue'
import SurveysPage from '@/components/app/SurveysPage.vue'
import ErrCode from '@/components/ui/ErrCode.vue'
import { getSurveys } from '@/lib/api'
import type { Survey } from '@/lib/types'

const surveys = ref<Survey[]>([])
const errMsg = ref<string>('')
const loading = ref<boolean>(true)

onMounted(async () => {
  const surveysResp = await getSurveys()
  if (surveysResp.error) {
    errMsg.value = 'Failed to fetch surveys'
  } else {
    surveys.value = surveysResp.data.data
  }
  loading.value = false
})
</script>
