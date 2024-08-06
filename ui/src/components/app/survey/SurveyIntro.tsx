'use client'

import { useState } from 'react'
import { Survey, SurveyConfig, SurveySession } from 'lib/types'
import { Button } from 'flowbite-react'
import { ErrCode } from 'components/ui/ErrCode'
import SurveyQuestions from 'components/app/survey/SurveyQuestions'
import { createSurveySession } from 'lib/api'

type SurveyIntroProps = {
  survey: Survey
}

export default function SurveyIntro({ survey }: SurveyIntroProps) {
  const [errMessage, seterrMessage] = useState<string | undefined>(undefined)
  const [surveySession, setSurveySession] = useState<SurveySession | undefined>(
    undefined
  )
  const config = survey.config as SurveyConfig

  if (surveySession !== undefined) {
    return (
      <SurveyQuestions
        survey={survey}
        session={surveySession as SurveySession}
      />
    )
  }

  return (
    <div className="intro">
      <h1 className="h1">{config.title}</h1>
      <p
        className="intro-title"
        dangerouslySetInnerHTML={{
          __html: config.intro.replace(/(?:\r\n|\r|\n)/g, '<br>'),
        }}
      ></p>
      <div className="intro-start">
        <Button
          onClick={async () => {
            seterrMessage(undefined)
            const sessionRes = await createSurveySession(
              window.location.hostname,
              survey.url_slug
            )
            if (sessionRes.error) {
              seterrMessage(sessionRes.error)
              return
            }

            localStorage.setItem(
              `survey_session_id:${survey.url_slug}`,
              sessionRes.data.data.uuid
            )
            setSurveySession(sessionRes.data.data)
          }}
        >
          Start
        </Button>
      </div>
      {errMessage && (
        <div className="flex flex-col py-8 px-8">
          <ErrCode message={errMessage} />
        </div>
      )}
    </div>
  )
}
