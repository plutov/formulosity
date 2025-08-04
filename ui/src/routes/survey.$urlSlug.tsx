import { useEffect, useState } from 'react'
import { useParams } from 'react-router'
import { getSurvey } from 'lib/api'
import { Survey } from 'lib/types'
import SurveyLayout from 'components/app/survey/SurveyLayout'
import SurveyNotFound from 'components/app/survey/SurveyNotFound'
import SurveyForm from 'components/app/survey/SurveyForm'

export default function SurveyPage() {
  const { urlSlug } = useParams<{ urlSlug: string }>()
  const [survey, setSurvey] = useState<Survey | null>(null)
  const [loading, setLoading] = useState(true)
  const [notFound, setNotFound] = useState(false)

  useEffect(() => {
    async function fetchSurvey() {
      if (!urlSlug) {
        setNotFound(true)
        setLoading(false)
        return
      }

      try {
        const surveyResp = await getSurvey(window.location.host, urlSlug)
        
        if (
          surveyResp.error ||
          !surveyResp.data.data ||
          !surveyResp.data.data.config
        ) {
          setNotFound(true)
          document.title = 'Survey not found'
        } else {
          const surveyData = surveyResp.data.data as Survey
          setSurvey(surveyData)
          document.title = surveyData.config.title
        }
      } catch (error) {
        console.error('Error fetching survey:', error)
        setNotFound(true)
        document.title = 'Survey not found'
      } finally {
        setLoading(false)
      }
    }

    fetchSurvey()
  }, [urlSlug])

  const apiURL = import.meta.env.VITE_API_URL || ''

  if (loading) {
    return (
      <SurveyLayout apiURL={apiURL}>
        <div className="flex justify-center items-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        </div>
      </SurveyLayout>
    )
  }

  if (notFound || !survey) {
    return (
      <SurveyLayout apiURL={apiURL}>
        <SurveyNotFound />
      </SurveyLayout>
    )
  }

  return (
    <SurveyLayout
      surveyTheme={survey.config.theme}
      urlSlug={survey.url_slug}
      apiURL={apiURL}
    >
      <SurveyForm survey={survey} apiURL={apiURL} />
    </SurveyLayout>
  )
}