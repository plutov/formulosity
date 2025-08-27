<template>
  <div v-if="localSession.status === SurveySessionStatus.Completed">
    <SurveyFooter :survey="survey" />
  </div>
  <div v-else-if="currentQuestion">
    <div class="question-container">
      <h2 class="h2">{{ currentQuestion.label }}</h2>
      <p v-if="currentQuestion.description" class="text-gray-600">{{ currentQuestion.description }}</p>
      
      <!-- Single Choice -->
      <div v-if="currentQuestion.type === SurveyQuestionType.SingleChoice" class="space-y-2">
        <div v-for="option in currentQuestion.options" :key="option" class="flex items-center">
          <input 
            type="radio" 
            :id="option" 
            :name="currentQuestion.uuid"
            :value="option"
            v-model="answerValue"
            class="mr-2"
          />
          <label :for="option">{{ option }}</label>
        </div>
      </div>

      <!-- Multiple Choice -->
      <div v-else-if="currentQuestion.type === SurveyQuestionType.MultipleChoice" class="space-y-2">
        <div v-for="option in currentQuestion.options" :key="option" class="flex items-center">
          <input 
            type="checkbox" 
            :id="option" 
            :value="option"
            v-model="answerValue"
            class="mr-2"
          />
          <label :for="option">{{ option }}</label>
        </div>
      </div>

      <!-- Short Text -->
      <div v-else-if="currentQuestion.type === SurveyQuestionType.ShortText">
        <input 
          type="text"
          v-model="answerValue"
          class="w-full p-2 border border-gray-300 rounded"
        />
      </div>

      <!-- Long Text -->
      <div v-else-if="currentQuestion.type === SurveyQuestionType.LongText">
        <textarea 
          v-model="answerValue"
          class="w-full p-2 border border-gray-300 rounded h-32"
        ></textarea>
      </div>

      <!-- Email -->
      <div v-else-if="currentQuestion.type === SurveyQuestionType.Email">
        <input 
          type="email"
          v-model="answerValue"
          class="w-full p-2 border border-gray-300 rounded"
        />
      </div>

      <!-- Date -->
      <div v-else-if="currentQuestion.type === SurveyQuestionType.Date">
        <input 
          type="date"
          v-model="answerValue"
          class="w-full p-2 border border-gray-300 rounded"
        />
      </div>

      <!-- Yes/No -->
      <div v-else-if="currentQuestion.type === SurveyQuestionType.YesNo" class="space-y-2">
        <div class="flex items-center">
          <input 
            type="radio" 
            id="yes" 
            name="yesno"
            value="true"
            v-model="answerValue"
            class="mr-2"
          />
          <label for="yes">Yes</label>
        </div>
        <div class="flex items-center">
          <input 
            type="radio" 
            id="no" 
            name="yesno"
            value="false"
            v-model="answerValue"
            class="mr-2"
          />
          <label for="no">No</label>
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

      <div class="mt-4 flex gap-2">
        <button 
          v-if="prevQuestion"
          @click="goToPrev"
          class="px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600 flex items-center gap-1"
        >
          <Icon icon="heroicons:arrow-left" class="w-4 h-4" />
          Previous
        </button>
        <button 
          @click="submitAnswer"
          class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 flex items-center gap-1"
          :disabled="!answerValue"
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

const props = defineProps<Props>()

const initialQuestion = determineInitialQuestion(props.survey, props.session)
const currentQuestion = ref<SurveyQuestion | undefined>(initialQuestion)
const answerValue = ref<string | string[]>('')
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

// Initialize answer value from existing answer
watch(currentQuestion, (newQuestion) => {
  if (newQuestion?.answer?.value !== undefined) {
    const value = newQuestion.answer.value
    if (Array.isArray(value)) {
      answerValue.value = value.map(v => String(v))
    } else {
      answerValue.value = String(value)
    }
  } else {
    if (newQuestion?.type === SurveyQuestionType.MultipleChoice) {
      answerValue.value = []
    } else {
      answerValue.value = ''
    }
  }
}, { immediate: true })

async function submitAnswer() {
  if (!currentQuestion.value) return

  errMessage.value = ''
  
  const payload = {
    value: answerValue.value
  }

  const res = await submitQuestionAnswer(
    props.survey.url_slug,
    localSession.value.uuid,
    currentQuestion.value.uuid,
    payload
  )

  if (res.error) {
    errMessage.value = res.error
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
      answer: { value: answerValue.value }
    }
  } else {
    localSession.value.question_answers.push({
      question_id: currentQuestion.value.id,
      question_uuid: currentQuestion.value.uuid,
      answer: { value: answerValue.value }
    })
  }

  // Move to next question or complete
  const next = nextQuestion.value
  if (next) {
    currentQuestion.value = next
  } else {
    // Mark as completed
    localSession.value.status = SurveySessionStatus.Completed
  }
}

function goToPrev() {
  const prev = prevQuestion.value
  if (prev) {
    currentQuestion.value = prev
  }
}
</script>