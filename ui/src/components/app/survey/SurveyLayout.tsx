'use client'

import { BgPattern } from 'components/ui/BgPattern'
import { getConsoleApiHost } from 'lib/api'
import { SurveyThemeCustom } from 'lib/types'
import { ReactNode } from 'react'
import 'styles/survey/default.css'

type SurveyLayoutProps = {
  children?: ReactNode
  urlSlug?: string
  surveyTheme?: string
}

export default function SurveyLayout({
  children,
  urlSlug,
  surveyTheme,
}: SurveyLayoutProps) {
  return (
    <main className="flex flex-col h-screen">
      {surveyTheme == SurveyThemeCustom && (
        <link
          rel="stylesheet"
          href={getConsoleApiHost() + '/surveys/' + urlSlug + '/css'}
        />
      )}
      <BgPattern />
      {children}
    </main>
  )
}
