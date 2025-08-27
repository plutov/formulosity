'use client'

import { BgPattern } from 'components/ui/BgPattern'
import { SurveyThemeCustom } from 'lib/types'
import { ReactNode } from 'react'
import 'styles/survey/default.css'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_ADDR || 'http://localhost:9900'

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
          href={API_BASE_URL + '/surveys/' + urlSlug + '/css'}
        />
      )}
      <BgPattern />
      {children}
    </main>
  )
}
