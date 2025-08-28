<template>
  <div v-if="localSession.status === SurveySessionStatus.Completed">
    <SurveyFooter :survey="survey" />
  </div>
  <div v-else-if="!currentQuestion">
    <div class="completion-message text-center py-8">
      <h2 class="h2 text-gray-300">No more questions found in the survey.</h2>
      <div class="mt-6">
        <button 
          @click="tryAgain"
          class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          Try again
        </button>
      </div>
    </div>
  </div>
  <div v-else-if="currentQuestion">
    <div class="question-container">
      <h2 class="h2 text-gray-300">{{ currentQuestion.label }}</h2>
      <p v-if="currentQuestion.description" class="text-gray-300 mb-6">{{ currentQuestion.description }}</p>
      
      <!-- Single Choice -->
      <div v-if="currentQuestion.type === SurveyQuestionType.SingleChoice" class="space-y-3">
        <div v-for="option in currentQuestion.options" :key="option" class="flex items-center">
          <input 
            type="radio" 
            :id="option" 
            :name="currentQuestion.uuid"
            :value="option"
            v-model="answerValue"
            class="mr-3 text-blue-600 focus:ring-blue-500"
          />
          <label :for="option" class="text-gray-300">{{ option }}</label>
        </div>
      </div>

      <!-- Multiple Choice -->
      <div v-else-if="currentQuestion.type === SurveyQuestionType.MultipleChoice" class="space-y-3">
        <div v-for="option in currentQuestion.options" :key="option" class="flex items-center">
          <input 
            type="checkbox" 
            :id="option" 
            :value="option"
            v-model="answerValue"
            class="mr-3 text-blue-600 focus:ring-blue-500 rounded"
          />
          <label :for="option" class="text-gray-300">{{ option }}</label>
        </div>
      </div>

      <!-- Short Text -->
      <div v-else-if="currentQuestion.type === SurveyQuestionType.ShortText">
        <input 
          type="text"
          v-model="answerValue"
          class="w-full p-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-gray-900"
        />
      </div>

      <!-- Long Text -->
      <div v-else-if="currentQuestion.type === SurveyQuestionType.LongText">
        <textarea 
          v-model="answerValue"
          class="w-full p-3 border border-gray-300 rounded-lg h-32 focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-gray-900"
        ></textarea>
      </div>

      <!-- Email -->
      <div v-else-if="currentQuestion.type === SurveyQuestionType.Email">
        <input 
          type="email"
          v-model="answerValue"
          class="w-full p-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-gray-900"
        />
      </div>

      <!-- Date -->
      <div v-else-if="currentQuestion.type === SurveyQuestionType.Date">
        <input 
          type="date"
          v-model="answerValue"
          class="w-full p-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-gray-900"
        />
      </div>

      <!-- Yes/No -->
      <div v-else-if="currentQuestion.type === SurveyQuestionType.YesNo" class="space-y-3">
        <div class="flex items-center">
          <input 
            type="radio" 
            id="yes" 
            name="yesno"
            value="true"
            v-model="answerValue"
            class="mr-3 text-blue-600 focus:ring-blue-500"
          />
          <label for="yes" class="text-gray-300">Yes</label>
        </div>
        <div class="flex items-center">
          <input 
            type="radio" 
            id="no" 
            name="yesno"
            value="false"
            v-model="answerValue"
            class="mr-3 text-blue-600 focus:ring-blue-500"
          />
          <label for="no" class="text-gray-300">No</label>
        </div>
      </div>

      <!-- Rating -->
      <div v-else-if="currentQuestion.type === SurveyQuestionType.Rating" class="flex items-center space-x-2">
        <span v-for="i in (currentQuestion.max || 5)" :key="i" class="cursor-pointer text-2xl select-none">
          <Icon 
            icon="heroicons:star-solid"
            :class="[i <= Number(answerValue) ? 'text-yellow-400' : 'text-gray-300']"
            @click="answerValue = i.toString()"
          />
        </span>
      </div>

      <!-- Ranking -->
      <div v-else-if="currentQuestion.type === SurveyQuestionType.Ranking" class="space-y-4">
        <p class="text-gray-300 text-sm">Drag and drop items below to rank them.</p>
        <div class="space-y-2">
          <div 
            v-for="(item, index) in sortableItems" 
            :key="item.id"
            draggable="true"
            @dragstart="onDragStart(index)"
            @dragover.prevent
            @drop="onDrop(index)"
            @dragenter.prevent
            class="flex items-center p-3 bg-gray-100 border border-gray-300 rounded-lg cursor-move hover:bg-gray-200 transition-colors"
            :class="{ 'opacity-50': draggedIndex === index }"
          >
            <Icon icon="heroicons:bars-3" class="w-5 h-5 text-gray-500 mr-3" />
            <span class="text-gray-900">{{ item.name }}</span>
          </div>
        </div>
      </div>

      <div class="mt-6 flex gap-3">
        <button 
          v-if="prevQuestion"
          @click="goToPrev"
          class="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 flex items-center gap-2"
        >
          <Icon icon="heroicons:arrow-left" class="w-4 h-4" />
          Previous
        </button>
        <button 
          @click="submitAnswer"
          class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
          :disabled="!canSubmit"
        >
          {{ nextQuestion ? 'Next' : 'Complete' }}
          <Icon icon="heroicons:arrow-right" class="w-4 h-4" />
        </button>
      </div>

      <ErrCode v-if="errMessage" :message="errMessage" class="mt-4" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Icon } from '@iconify/vue'
import {
  SurveyQuestionType,
  SurveySessionStatus,
} from '@/lib/types'
import type {
  Survey,
  SurveySession,
  SurveyQuestion,
} from '@/lib/types'
import { submitQuestionAnswer } from '@/lib/api'
import ErrCode from '@/components/ui/ErrCode.vue'
import SurveyFooter from './SurveyFooter.vue'
import {
  determineNextQuestion,
  determinePrevQuestionWithAnswer,
  determineInitialQuestion,
} from '@/lib/questions'

interface Props {
  survey: Survey
  session: SurveySession
}

interface SortableItem {
  id: number
  name: string
}

const props = defineProps<Props>()

const initialQuestion = determineInitialQuestion(props.survey, props.session)
const currentQuestion = ref<SurveyQuestion | undefined>(initialQuestion)
const answerValue = ref<string | string[]>('')
const sortableItems = ref<SortableItem[]>([])
const draggedIndex = ref<number | null>(null)
const errMessage = ref<string>('')
const localSession = ref<SurveySession>({ ...props.session })

const nextQuestion = computed(() => {
  if (!currentQuestion.value) return undefined
  return determineNextQuestion(props.survey, localSession.value, currentQuestion.value)
})

const prevQuestion = computed(() => {
  if (!currentQuestion.value) return undefined
  return determinePrevQuestionWithAnswer(props.survey, localSession.value, currentQuestion.value)
})

const canSubmit = computed(() => {
  if (!currentQuestion.value) return false
  
  if (currentQuestion.value.type === SurveyQuestionType.Ranking) {
    return sortableItems.value.length > 0
  }
  
  return !!answerValue.value
})

// Drag and drop functions
function onDragStart(index: number) {
  draggedIndex.value = index
}

function onDrop(dropIndex: number) {
  if (draggedIndex.value === null || draggedIndex.value === dropIndex) return
  
  const draggedItem = sortableItems.value[draggedIndex.value]
  sortableItems.value.splice(draggedIndex.value, 1)
  sortableItems.value.splice(dropIndex, 0, draggedItem)
  
  draggedIndex.value = null
}

// Initialize answer value from existing answer
watch(currentQuestion, (newQuestion) => {
  if (!newQuestion) return

  if (newQuestion.type === SurveyQuestionType.Ranking) {
    // Initialize sortable items for ranking questions
    if (newQuestion.answer?.value) {
      const rankingArray = newQuestion.answer.value as string[]
      sortableItems.value = rankingArray.map((item, index) => ({
        id: index,
        name: item
      }))
    } else {
      sortableItems.value = newQuestion.options.map((option, index) => ({
        id: index,
        name: option
      }))
    }
    answerValue.value = []
  } else if (newQuestion.answer?.value !== undefined) {
    const value = newQuestion.answer.value
    if (Array.isArray(value)) {
      answerValue.value = value.map(v => String(v))
    } else {
      answerValue.value = String(value)
    }
  } else {
    if (newQuestion.type === SurveyQuestionType.MultipleChoice) {
      answerValue.value = []
    } else {
      answerValue.value = ''
    }
  }
}, { immediate: true })

async function submitAnswer() {
  if (!currentQuestion.value) return

  errMessage.value = ''
  
  // Convert answer value based on question type
  let processedValue = answerValue.value
  if (currentQuestion.value.type === SurveyQuestionType.Rating) {
    processedValue = Number(answerValue.value)
  } else if (currentQuestion.value.type === SurveyQuestionType.YesNo) {
    processedValue = answerValue.value === 'true'
  } else if (currentQuestion.value.type === SurveyQuestionType.Ranking) {
    processedValue = sortableItems.value.map(item => item.name)
  }
  
  const payload = {
    value: processedValue
  }

  const res = await submitQuestionAnswer(
    props.survey.url_slug,
    localSession.value.uuid,
    currentQuestion.value.uuid,
    payload
  )

  if (res.error) {
    // Check if we have detailed error information
    if (res.data && res.data.error_details) {
      errMessage.value = res.data.error_details
    } else {
      errMessage.value = res.error
    }
    return
  }

  // Update local session with new answer
  const existingAnswerIndex = localSession.value.question_answers.findIndex(
    (a) => a.question_uuid === currentQuestion.value!.uuid
  )

  if (existingAnswerIndex >= 0) {
    localSession.value.question_answers[existingAnswerIndex] = {
      question_id: currentQuestion.value.id,
      question_uuid: currentQuestion.value.uuid,
      answer: { value: processedValue }
    }
  } else {
    localSession.value.question_answers.push({
      question_id: currentQuestion.value.id,
      question_uuid: currentQuestion.value.uuid,
      answer: { value: processedValue }
    })
  }

  // Move to next question or complete
  const next = nextQuestion.value
  if (next) {
    currentQuestion.value = next
  } else {
    // Mark as completed - this should trigger the thank you page
    console.log('Survey completed, setting status to completed')
    localSession.value.status = SurveySessionStatus.Completed
    localStorage.removeItem(`survey_session_id:${props.survey.url_slug}`)
  }
}

function goToPrev() {
  const prev = prevQuestion.value
  if (prev) {
    currentQuestion.value = prev
  }
}

function tryAgain() {
  localStorage.removeItem(`survey_session_id:${props.survey.url_slug}`)
  window.location.reload()
}
</script>

<style scoped>
/* Drag and drop styling handled via Tailwind classes in template */
</style>