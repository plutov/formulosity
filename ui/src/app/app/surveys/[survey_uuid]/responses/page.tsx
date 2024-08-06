import { Metadata } from 'next'

import AppLayout from 'components/app/AppLayout'
import { ErrCode } from 'components/ui/ErrCode'
import { getSurveys, getSurveySessions } from 'lib/api'
import { Survey, SurveySessionsLimit } from 'lib/types'
import { SurveyResponsesPage } from 'components/app/SurveyResponsesPage'

export const metadata: Metadata = {
  title: 'Survey Responses',
}

export default async function ResponsesPage({
  params,
}: {
  params: { survey_uuid: string }
}) {
  let errMsg = ''

  let currentSurvey = undefined
  const surveysResp = await getSurveys()
  if (surveysResp.error) {
    errMsg = 'Unable to fetch surveys'
  } else {
    const surveys = surveysResp.data.data
    const survey = surveys.find(
      (survey: Survey) => survey.uuid === params.survey_uuid
    )
    if (!survey) {
      errMsg = 'Survey not found'
    } else {
      currentSurvey = survey

      const surveySessionsResp = await getSurveySessions(
        currentSurvey.uuid,
        `limit=${SurveySessionsLimit}&offset=0&sort_by=created_at&order=desc`
      )
      if (surveySessionsResp.error) {
        errMsg = 'Unable to fetch survey sessions'
      } else {
        currentSurvey = surveySessionsResp.data.data.survey
        currentSurvey.sessions = surveySessionsResp.data.data.sessions
        currentSurvey.pages_count = surveySessionsResp.data.data.pages_count
      }
    }
  }

  if (errMsg) {
    return (
      <AppLayout>
        <ErrCode message={errMsg} />
      </AppLayout>
    )
  }

  return (
    <AppLayout>
      <SurveyResponsesPage currentSurvey={currentSurvey} />
    </AppLayout>
  )
}
