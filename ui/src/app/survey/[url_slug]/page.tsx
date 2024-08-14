import { getSurvey } from 'lib/api'
import { Survey } from 'lib/types'
import SurveyLayout from 'components/app/survey/SurveyLayout'
import SurveyNotFound from 'components/app/survey/SurveyNotFound'
import SurveyForm from 'components/app/survey/SurveyForm'
import { headers } from 'next/headers'

export async function generateMetadata({
  params,
}: {
  params: { url_slug: string }
}) {
  const headersList = headers()
  const surveyResp = await getSurvey(
    headersList.get('host') as string,
    params.url_slug
  )
  if (
    surveyResp.error ||
    !surveyResp.data.data ||
    !surveyResp.data.data.config
  ) {
    return {
      title: 'Survey not found',
    }
  }

  const survey = surveyResp.data.data as Survey

  return {
    title: survey.config.title,
  }
}

export default async function SurveyPage({
  params,
}: {
  params: { url_slug: string }
}) {
  const headersList = headers()
  const surveyResp = await getSurvey(
    headersList.get('host') as string,
    params.url_slug
  )
  const apiURL = process.env.CONSOLE_API_ADDR || ''
  if (
    surveyResp.error ||
    !surveyResp.data.data ||
    !surveyResp.data.data.config
  ) {
    return (
      <SurveyLayout apiURL={apiURL}>
        <SurveyNotFound />
      </SurveyLayout>
    )
  }

  const survey = surveyResp.data.data as Survey

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
