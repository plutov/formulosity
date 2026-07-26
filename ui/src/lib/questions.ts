import type { QuestionCondition, Survey, SurveyQuestion, SurveySession } from './types'

type AnswerMap = Record<string, unknown>

export function answersFor(survey: Survey, session: SurveySession): AnswerMap {
  const answers: AnswerMap = {}
  for (const answer of session.question_answers || []) {
    const question = survey.config.questions.questions.find((q) => q.uuid === answer.question_uuid)
    if (question && answer.answer) answers[question.id] = answer.answer.value
  }
  return answers
}

function equal(left: unknown, right: unknown): boolean {
  if (typeof left === 'number' && typeof right === 'number') return left === right
  if (Array.isArray(left) && Array.isArray(right)) return left.length === right.length && left.every((v, i) => equal(v, right[i]))
  return left === right
}

export function evaluateCondition(condition: QuestionCondition, answers: AnswerMap): boolean {
  if (condition.all) return condition.all.every((child) => evaluateCondition(child, answers))
  if (condition.any) return condition.any.some((child) => evaluateCondition(child, answers))
  if (condition.not) return !evaluateCondition(condition.not, answers)
  const value = condition.question ? answers[condition.question] : undefined
  const answered = condition.question !== undefined && answers[condition.question] !== undefined
  if (condition.is_answered !== undefined) return answered === condition.is_answered
  if (condition.is_not_answered !== undefined) return answered !== condition.is_not_answered
  if (condition.equals !== undefined) return equal(value, condition.equals)
  if (condition.not_equals !== undefined) return !equal(value, condition.not_equals)
  if (condition.contains !== undefined) return Array.isArray(value) ? value.some((item) => equal(item, condition.contains)) : false
  if (condition.greater_than !== undefined) return typeof value === 'number' && value > Number(condition.greater_than)
  if (condition.greater_than_or_equal !== undefined) return typeof value === 'number' && value >= Number(condition.greater_than_or_equal)
  if (condition.less_than !== undefined) return typeof value === 'number' && value < Number(condition.less_than)
  if (condition.less_than_or_equal !== undefined) return typeof value === 'number' && value <= Number(condition.less_than_or_equal)
  return false
}

export function isQuestionVisible(question: SurveyQuestion, answers: AnswerMap): boolean {
  return question.visible_if === undefined || evaluateCondition(question.visible_if, answers)
}

export function getVisibleQuestions(survey: Survey, session: SurveySession): SurveyQuestion[] {
  // Evaluate in declaration order and only expose answers belonging to visible
  // questions. This makes dependent cleanup transitive (A -> B -> C) even if
  // stale answers were returned by an older server.
  const storedAnswers = answersFor(survey, session)
  const visibleAnswers: AnswerMap = {}
  const visible: SurveyQuestion[] = []
  for (const question of survey.config.questions.questions) {
    if (!isQuestionVisible(question, visibleAnswers)) continue
    visible.push(question)
    if (question.id in storedAnswers) visibleAnswers[question.id] = storedAnswers[question.id]
  }
  return visible
}

// clearHiddenAnswers keeps local state safe and responsive before (or in case
// of) a subsequent authoritative API response.
export function clearHiddenAnswers(survey: Survey, session: SurveySession): SurveySession {
  const visibleUUIDs = new Set(getVisibleQuestions(survey, session).map((question) => question.uuid))
  return {
    ...session,
    question_answers: (session.question_answers || []).filter((answer) => visibleUUIDs.has(answer.question_uuid)),
  }
}

export function interpolateText(text: string, answers: AnswerMap): string {
  return text.replace(/{{\s*([A-Za-z0-9_-]+)\s*}}/g, (_, id: string) => {
    const value = answers[id]
    if (value === undefined || value === null) return ''
    return Array.isArray(value) ? value.join(', ') : String(value)
  })
}

export function determineInitialQuestion(survey: Survey, session: SurveySession): SurveyQuestion | undefined {
  const qList = getVisibleQuestions(survey, session)
  for (let i = 0; i < qList.length; i++) {
    const q = qList[i]
    const answerFound = (session.question_answers || []).find((a) => a.question_uuid === q.uuid)
    if (answerFound === undefined) { q.index = i; return q }
  }
  return undefined
}

function visibleIndex(survey: Survey, session: SurveySession, current: SurveyQuestion): { list: SurveyQuestion[]; index: number } {
  const list = getVisibleQuestions(survey, session)
  return { list, index: list.findIndex((q) => q.uuid === current.uuid) }
}

export function determineNextQuestion(survey: Survey, session: SurveySession, currentQuestion: SurveyQuestion): SurveyQuestion | undefined {
  const { list, index } = visibleIndex(survey, session, currentQuestion)
  if (index < 0 || index >= list.length - 1) return undefined
  const next = list[index + 1]
  next.index = index + 1
  const answer = (session.question_answers || []).find((a) => a.question_uuid === next.uuid)
  if (answer) next.answer = answer.answer
  return next
}

export function determinePrevQuestionWithAnswer(survey: Survey, session: SurveySession, current: SurveyQuestion): SurveyQuestion | undefined {
  const { list, index } = visibleIndex(survey, session, current)
  if (index <= 0) return undefined
  for (let i = index - 1; i >= 0; i--) {
    const prev = list[i]
    const answer = (session.question_answers || []).find((a) => a.question_uuid === prev.uuid)
    if (answer) { prev.index = i; prev.answer = answer.answer; return prev }
  }
  return undefined
}

export function determineNextQuestionWithAnswer(survey: Survey, session: SurveySession, current: SurveyQuestion): SurveyQuestion | undefined {
  const { list, index } = visibleIndex(survey, session, current)
  if (index < 0) return undefined
  for (let i = index + 1; i < list.length; i++) {
    const next = list[i]
    const answer = (session.question_answers || []).find((a) => a.question_uuid === next.uuid)
    if (answer) { next.index = i; next.answer = answer.answer; return next }
  }
  return undefined
}
