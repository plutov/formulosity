'use client'

import { BgPattern } from 'components/ui/BgPattern'
import { ReactNode } from 'react'
import 'styles/survey/default.css'

type SurveyLayoutProps = {
  children?: ReactNode
  customThemeURL?: string
}

export default function SurveyLayout({
  children,
  customThemeURL,
}: SurveyLayoutProps) {
  return (
    <main className="flex flex-col h-screen">
      {customThemeURL && <link rel="stylesheet" href={customThemeURL} />}
      <BgPattern />
      {children}
    </main>
  )
}
