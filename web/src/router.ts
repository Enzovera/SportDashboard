import { createRouter, createWebHistory } from 'vue-router'
import { applyMockParam } from '@/api/client'

export const router = createRouter({
  history: createWebHistory(),
  // /propozycje#<id> (links from Plan) scrolls to the card; other navigations start at the top.
  scrollBehavior: (to) => (to.hash ? { el: to.hash } : { top: 0 }),
  routes: [
    { path: '/', name: 'today', component: () => import('@/views/today/TodayView.vue') },
    { path: '/trening', name: 'training', component: () => import('@/views/training/TrainingView.vue') },
    { path: '/zywienie', name: 'nutrition', component: () => import('@/views/nutrition/NutritionView.vue') },
    { path: '/zywienie/suplementy', name: 'supplements', component: () => import('@/views/nutrition/SupplementsView.vue') },
    { path: '/cialo', name: 'body', component: () => import('@/views/body/BodyView.vue') },
    { path: '/plan', name: 'plan', component: () => import('@/views/plan/PlanView.vue') },
    { path: '/propozycje', name: 'proposals', component: () => import('@/views/proposals/ProposalsView.vue') },
    // Full-screen chat on the cover screen; on fold/desktop the chat is a panel next to the current view.
    { path: '/czat', name: 'chat', component: () => import('@/views/chat/ChatView.vue') },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

// ?mock=1|0 only sets the cookie and disappears from the address before the view loads data.
router.beforeEach((to) => {
  if (!('mock' in to.query)) return
  applyMockParam(to.query.mock)
  const { mock: _, ...query } = to.query
  return { path: to.path, query, hash: to.hash, replace: true }
})
