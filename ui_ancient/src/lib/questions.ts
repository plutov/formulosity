import { Survey, SurveyQuestion, SurveySession } from './types'

export function determineInitialQuestion(
  survey: Survey,
  session: SurveySession
): SurveyQuestion | undefined {
  const qList = survey.config.questions.questions

  for (let i = 0; i < qList.length; i++) {
    const q = qList[i]

    const answers = session.question_answers || []
    const answerFound = answers.find((a) => a.question_uuid === q.uuid)

    if (answerFound === undefined) {
      q.index = i
      return q
    }
  }

  return undefined
}

export function determineNextQuestion(
  survey: Survey,
  session: SurveySession,
  currentQuestion: SurveyQuestion
): SurveyQuestion | undefined {
  const qList = survey.config.questions.questions

  for (let i = 0; i < qList.length; i++) {
    const q = qList[i]

    const isCurrent = q.uuid === currentQuestion.uuid
    const isLast = i === qList.length - 1
    if (isCurrent) {
      if (isLast) {
        return undefined
      }

      const next = qList[i + 1]
      const answers = session.question_answers || []
      const answerFound = answers.find((a) => a.question_uuid === next.uuid)

      next.index = i + 1
      if (answerFound !== undefined) {
        next.answer = answerFound.answer
      }
      return next
    }
  }

  return undefined
}

export function determinePrevQuestionWithAnswer(
  survey: Survey,
  session: SurveySession,
  current: SurveyQuestion
): SurveyQuestion | undefined {
  const qList = survey.config.questions.questions

  let prev: SurveyQuestion | undefined
  for (let i = 0; i < qList.length; i++) {
    const q = qList[i]
    const isCurrent = q.uuid === current.uuid
    if (isCurrent) {
      if (prev === undefined) {
        return undefined
      }

      const answers = session.question_answers || []
      const answerFound = answers.find((a) => a.question_uuid === prev?.uuid)

      if (answerFound !== undefined) {
        prev.index = i - 1
        prev.answer = answerFound.answer
        return prev
      }
    }

    prev = q
  }

  return undefined
}

export function determineNextQuestionWithAnswer(
  survey: Survey,
  session: SurveySession,
  current: SurveyQuestion
): SurveyQuestion | undefined {
  const qList = survey.config.questions.questions

  for (let i = 0; i < qList.length; i++) {
    const q = qList[i]
    const isCurrent = q.uuid === current.uuid
    const isLast = i === qList.length - 1
    if (isCurrent) {
      if (isLast) {
        return undefined
      }

      const next = qList[i + 1]
      const answers = session.question_answers || []
      const answerFound = answers.find((a) => a.question_uuid === next.uuid)

      if (answerFound !== undefined) {
        next.index = i + 1
        next.answer = answerFound.answer
        return next
      }
    }
  }

  return undefined
}
