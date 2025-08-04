import { useEffect, useState } from 'react'
import { useParams } from 'react-router'

import AppLayout from 'components/app/AppLayout'
import { ErrCode } from 'components/ui/ErrCode'
import { getSurveys, getSurveySessions } from 'lib/api'
import { Survey, SurveySessionsLimit } from 'lib/types'
import { SurveyResponsesPage } from 'components/app/SurveyResponsesPage'

export default function ResponsesPage() {
  const { surveyUuid } = useParams<{ surveyUuid: string }>()
  const [currentSurvey, setCurrentSurvey] = useState<Survey | undefined>()
  const [errMsg, setErrMsg] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    document.title = 'Survey Responses | Formulosity'

    async function fetchSurveyData() {
      if (!surveyUuid) {
        setErrMsg('Survey UUID is required')
        setLoading(false)
        return
      }

      try {
        const surveysResp = await getSurveys()
        if (surveysResp.error) {
          setErrMsg('Unable to fetch surveys')
          setLoading(false)
          return
        }

        const surveys = surveysResp.data.data
        const survey = surveys.find(
          (survey: Survey) => survey.uuid === surveyUuid
        )

        if (!survey) {
          setErrMsg('Survey not found')
          setLoading(false)
          return
        }

        const apiURL = import.meta.env.VITE_API_URL || ''
        const surveySessionsResp = await getSurveySessions(
          survey.uuid,
          `limit=${SurveySessionsLimit}&offset=0&sort_by=created_at&order=desc`,
          apiURL
        )

        if (surveySessionsResp.error) {
          setErrMsg('Unable to fetch survey sessions')
        } else {
          const updatedSurvey = surveySessionsResp.data.data.survey
          updatedSurvey.sessions = surveySessionsResp.data.data.sessions
          updatedSurvey.pages_count = surveySessionsResp.data.data.pages_count
          setCurrentSurvey(updatedSurvey)
        }
      } catch (error) {
        console.error('Error fetching survey data:', error)
        setErrMsg('Failed to fetch survey data')
      } finally {
        setLoading(false)
      }
    }

    fetchSurveyData()
  }, [surveyUuid])

  if (loading) {
    return (
      <AppLayout>
        <div className="flex justify-center items-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        </div>
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

  const apiURL = import.meta.env.VITE_API_URL || ''

  return (
    <AppLayout>
      <SurveyResponsesPage currentSurvey={currentSurvey} apiURL={apiURL} />
    </AppLayout>
  )
}