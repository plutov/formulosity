import { Metadata } from 'next'

import AppLayout from 'components/app/AppLayout'
import { SurveysPage } from 'components/app/SurveysPage'
import { ErrCode } from 'components/ui/ErrCode'
import { getSurveys } from 'lib/api'

export const metadata: Metadata = {
  title: 'Formulosity',
}

export default async function AppPage() {
  let errMsg = ''

  let surveys = []
  const surveysResp = await getSurveys()
  if (surveysResp.error) {
    errMsg = 'Failed to fetch surveys'
  } else {
    surveys = surveysResp.data.data
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
      <SurveysPage surveys={surveys} />
    </AppLayout>
  )
}
