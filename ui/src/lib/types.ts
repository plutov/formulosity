export enum SurveyParseStatus {
  Success = 'success',
  Error = 'error',
  Deleted = 'deleted',
}

export enum SurveyDeliveryStatus {
  Launched = 'launched',
  Stopped = 'stopped',
}

export type Survey = {
  uuid: string
  created_at: string
  name: string
  parse_status: string
  delivery_status: string
  error_log: string
  url: string
  url_slug: string
  config: SurveyConfig
  stats: SurveyStats
  sessions: Array<SurveySession>
  pages_count: number
}

export const SurveySessionsLimit = 50

export type SurveyStats = {
  sessions_count_in_progress: number
  sessions_count_completed: number
  completion_rate: number
}

export const SurveyThemeCustom = 'custom'

export type QuestionCondition = {
  question?: string
  equals?: unknown
  not_equals?: unknown
  contains?: unknown
  greater_than?: unknown
  greater_than_or_equal?: unknown
  less_than?: unknown
  less_than_or_equal?: unknown
  is_answered?: boolean
  is_not_answered?: boolean
  all?: QuestionCondition[]
  any?: QuestionCondition[]
  not?: QuestionCondition
}

export type SurveyConfig = {
  title: string
  intro: string
  outro: string
  theme: string
  questions: SurveyQuestions
}

export type SurveyQuestions = {
  questions: SurveyQuestion[]
}

export type SurveyQuestion = {
  id: string
  uuid: string
  type: SurveyQuestionType
  label: string
  description: string
  options: string[]
  min?: number
  max?: number
  index: number
  answer: SurveyQuestionAnswerData
  visible_if?: QuestionCondition
}

export enum SurveyQuestionType {
  SingleChoice = 'single-choice',
  MultipleChoice = 'multiple-choice',
  ShortText = 'short-text',
  LongText = 'long-text',
  Date = 'date',
  Rating = 'rating',
  Ranking = 'ranking',
  YesNo = 'yes-no',
  Email = 'email',
  File = 'file',
}

export enum SurveySessionStatus {
  Completed = 'completed',
  InProgress = 'in_progress',
}

export type SurveySession = {
  uuid: string
  status: SurveySessionStatus
  created_at: string
  completed_at: string
  question_answers: SurveyQuestionAnswer[]
  webhookData: WebhookData
}

export type WebhookData = {
  response: string
  statusCode: number
}

export type SurveyQuestionAnswer = {
  question_id: string
  question_uuid: string
  answer: SurveyQuestionAnswerData
}

export type SurveyQuestionAnswerData = {
  value: string[] | string | number | boolean
}