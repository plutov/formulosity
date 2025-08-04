import { useEffect, useState } from 'react'

import AppLayout from 'components/app/AppLayout'
import { SurveysPage } from 'components/app/SurveysPage'
import { ErrCode } from 'components/ui/ErrCode'
import { getSurveys } from 'lib/api'

export default function AppPage() {
  const [surveys, setSurveys] = useState([])
  const [errMsg, setErrMsg] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    document.title = 'Formulosity | App'

    async function fetchSurveys() {
      try {
        const surveysResp = await getSurveys()
        if (surveysResp.error) {
          setErrMsg('Failed to fetch surveys')
        } else {
          setSurveys(surveysResp.data.data)
        }
      } catch (error) {
        setErrMsg('Failed to fetch surveys')
        console.error('Error fetching surveys:', error)
      } finally {
        setLoading(false)
      }
    }

    fetchSurveys()
  }, [])

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
      <SurveysPage surveys={surveys} apiURL={apiURL} />
    </AppLayout>
  )
}
