module.exports = {
  output: 'standalone',
  reactStrictMode: false,
  headers: () => [
    {
      source: '/survey/:url_slug',
      headers: [
        {
          key: 'Cache-Control',
          value: 'no-store',
        },
      ],
    },
  ],
  async redirects() {
    return [
      {
        source: '/',
        destination: '/app',
        permanent: true,
      },
    ]
  },
}
