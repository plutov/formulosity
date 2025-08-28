import { createRouter, createWebHistory } from 'vue-router'
import AppView from '../views/AppView.vue'
import SurveyView from '../views/SurveyView.vue'
import SurveyResponsesView from '../views/SurveyResponsesView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/app',
      name: 'app',
      component: AppView,
    },
    {
      path: '/survey/:urlSlug',
      name: 'survey',
      component: SurveyView,
    },
    {
      path: '/app/surveys/:surveyUuid/responses',
      name: 'surveyResponses',
      component: SurveyResponsesView,
    },
    {
      path: '/',
      redirect: '/app',
    },
  ],
})

export default router
