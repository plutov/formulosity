import React from 'react'
import ReactDOM from 'react-dom/client'
import { createBrowserRouter, RouterProvider } from 'react-router'
import 'styles/global.css'

// Import route components
import AppPage from './routes/app'
import SurveyPage from './routes/survey.$urlSlug'
import SurveyResponsesPage from './routes/app.surveys.$surveyUuid.responses'
import NotFoundPage from './routes/not-found'
import RootLayout from './routes/layout'

const router = createBrowserRouter([
  {
    path: '/',
    element: <RootLayout />,
    errorElement: <NotFoundPage />,
    children: [
      {
        path: 'app',
        element: <AppPage />,
      },
      {
        path: 'app/surveys/:surveyUuid/responses',
        element: <SurveyResponsesPage />,
      },
      {
        path: 'survey/:urlSlug',
        element: <SurveyPage />,
      },
      {
        path: '*',
        element: <NotFoundPage />,
      },
    ],
  },
])

const root = ReactDOM.createRoot(document.getElementById('root')!)
root.render(
  <React.StrictMode>
    <RouterProvider router={router} />
  </React.StrictMode>
)