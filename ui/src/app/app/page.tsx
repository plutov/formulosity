'use client'

import { useEffect, useState } from 'react'

import AppLayout from 'components/app/AppLayout'
import { SurveysPage } from 'components/app/SurveysPage'
import { ErrCode } from 'components/ui/ErrCode'
import { getSurveys } from 'lib/api'
import { Survey } from 'lib/types'

export default function AppPage() {
  const [surveys, setSurveys] = useState<Survey[]>([])
  const [errMsg, setErrMsg] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchSurveys = async () => {
      const surveysResp = await getSurveys()
      if (surveysResp.error) {
        setErrMsg('Failed to fetch surveys')
      } else {
        setSurveys(surveysResp.data.data)
      }
      setLoading(false)
    }
    fetchSurveys()
  }, [])

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
      <SurveysPage surveys={surveys} />
    </AppLayout>
  )
}
