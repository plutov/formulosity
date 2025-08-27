'use client'

import 'styles/global.css'
import { ReactNode } from 'react'

type LayoutProps = { children?: ReactNode }

export default function RootLayout({ children }: LayoutProps) {
  return (
    <html lang="en" className="dark" suppressHydrationWarning>
      <head />
      <body className="bg-slate-1 text-slate-12">{children}</body>
    </html>
  )
}
