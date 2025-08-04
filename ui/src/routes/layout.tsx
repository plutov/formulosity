import { Outlet } from 'react-router'
import { useEffect } from 'react'

export default function RootLayout() {
  useEffect(() => {
    // Set document title and meta tags
    document.title = 'Formulosity'
    
    // Set meta description
    const metaDescription = document.querySelector('meta[name="description"]')
    if (metaDescription) {
      metaDescription.setAttribute('content', 'Create and manage surveys with Formulosity')
    } else {
      const meta = document.createElement('meta')
      meta.name = 'description'
      meta.content = 'Create and manage surveys with Formulosity'
      document.head.appendChild(meta)
    }
  }, [])

  return <Outlet />
}