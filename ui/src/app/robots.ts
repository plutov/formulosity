import { MetadataRoute } from 'next'
import { siteConfig } from 'lib/siteConfig'

export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: '*',
      allow: '/',
      disallow: '/app/',
    },
    sitemap: `${siteConfig.url}/sitemap.xml`,
  }
}
