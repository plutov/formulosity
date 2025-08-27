'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import { getSurvey } from 'lib/api'
import { Survey } from 'lib/types'
import SurveyLayout from 'components/app/survey/SurveyLayout'
import SurveyNotFound from 'components/app/survey/SurveyNotFound'
import SurveyForm from 'components/app/survey/SurveyForm'

export default function SurveyPage() {
  const params = useParams()
  const [survey, setSurvey] = useState<Survey | null>(null)
  const [loading, setLoading] = useState(true)
  const [notFound, setNotFound] = useState(false)

  useEffect(() => {
    const fetchSurvey = async () => {
      if (!params.url_slug) return

      const surveyResp = await getSurvey(params.url_slug as string)

      if (
        surveyResp.error ||
        !surveyResp.data.data ||
        !surveyResp.data.data.config
      ) {
        setNotFound(true)
      } else {
        setSurvey(surveyResp.data.data as Survey)
      }
      setLoading(false)
    }

    fetchSurvey()
  }, [params.url_slug])

  if (loading) {
    return (
      <SurveyLayout>
        <div>Loading...</div>
      </SurveyLayout>
    )
  }

  if (notFound || !survey) {
    return (
      <SurveyLayout>
        <SurveyNotFound />
      </SurveyLayout>
    )
  }

  return (
    <SurveyLayout surveyTheme={survey.config.theme} urlSlug={survey.url_slug}>
      <SurveyForm survey={survey} />
    </SurveyLayout>
  )
}
