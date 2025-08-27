<template>
  <div>
    <div class="flex flex-col w-full tabs">
      <div class="flex flex-col w-full items-center">
        <div class="w-full">
          <div class="overflow-x-auto">
            <table class="w-full text-sm text-left text-gray-500 dark:text-gray-400">
              <thead class="text-xs text-gray-700 uppercase bg-gray-50 dark:bg-gray-700 dark:text-gray-400">
                <tr>
                  <th scope="col" class="px-6 py-3">Name/Title</th>
                  <th scope="col" class="px-6 py-3">Build</th>
                  <th scope="col" class="px-6 py-3">Delivery</th>
                  <th scope="col" class="px-6 py-3">Share</th>
                  <th scope="col" class="px-6 py-3">Responses</th>
                  <th scope="col" class="px-6 py-3">Completion</th>
                </tr>
              </thead>
              <tbody>
                <SurveyRow 
                  v-for="survey in surveys" 
                  :key="survey.uuid" 
                  :survey="survey" 
                />
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Survey } from '@/lib/types'
import SurveyRow from './SurveyRow.vue'

interface Props {
  surveys: Survey[]
}

defineProps<Props>()
</script>