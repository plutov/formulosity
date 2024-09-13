import 'styles/global.css'
import { ReactNode } from 'react'
import { Metadata } from 'next'
import { siteConfig } from 'lib/siteConfig'

export const metadata: Metadata = {
  title: {
    default: siteConfig.name,
    template: `%s | ${siteConfig.name}`,
  },
  description: siteConfig.description,
  keywords: [],
  alternates: {
    canonical: '/',
  },
  openGraph: {
    type: 'website',
    locale: 'en_US',
    title: siteConfig.name,
    description: siteConfig.description,
    siteName: siteConfig.name,
  },
  icons: {
    icon: '/favicon.ico',
  },
  manifest: '/manifest.webmanifest',
}

type LayoutProps = { children?: ReactNode }

export default async function RootLayout({ children }: LayoutProps) {
  return (
    <html lang="en" className="dark" suppressHydrationWarning>
      <head />
      <body className="bg-slate-1 text-slate-12">{children}</body>
    </html>
  )
}
