'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'

import AppLayout from 'components/app/AppLayout'
import { ErrCode } from 'components/ui/ErrCode'
import { getSurveys, getSurveySessions } from 'lib/api'
import { Survey, SurveySessionsLimit } from 'lib/types'
import { SurveyResponsesPage } from 'components/app/SurveyResponsesPage'

export default function ResponsesPage() {
  const params = useParams()
  const [currentSurvey, setCurrentSurvey] = useState<Survey | undefined>(
    undefined
  )
  const [errMsg, setErrMsg] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchData = async () => {
      if (!params.survey_uuid) return

      const surveysResp = await getSurveys()
      if (surveysResp.error) {
        setErrMsg('Unable to fetch surveys')
        setLoading(false)
        return
      }

      const surveys = surveysResp.data.data
      const survey = surveys.find(
        (survey: Survey) => survey.uuid === params.survey_uuid
      )

      if (!survey) {
        setErrMsg('Survey not found')
        setLoading(false)
        return
      }

      const surveySessionsResp = await getSurveySessions(
        survey.uuid,
        `limit=${SurveySessionsLimit}&offset=0&sort_by=created_at&order=desc`
      )

      if (surveySessionsResp.error) {
        setErrMsg('Unable to fetch survey sessions')
      } else {
        const updatedSurvey = surveySessionsResp.data.data.survey
        updatedSurvey.sessions = surveySessionsResp.data.data.sessions
        updatedSurvey.pages_count = surveySessionsResp.data.data.pages_count
        setCurrentSurvey(updatedSurvey)
      }

      setLoading(false)
    }

    fetchData()
  }, [params.survey_uuid])

  if (loading) {
    return (
      <AppLayout>
        <div>Loading...</div>
      </AppLayout>
    )
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
      <SurveyResponsesPage currentSurvey={currentSurvey!} />
    </AppLayout>
  )
}
