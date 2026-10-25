import { createRouter, createWebHistory } from 'vue-router'

const placeholder = (title: string) => ({
  component: () => import('@/views/PlaceholderView.vue'),
  props: { title },
})

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'today', component: () => import('@/views/today/TodayView.vue') },
    { path: '/trening', name: 'training', component: () => import('@/views/training/TrainingView.vue') },
    // Later stages (docs/design.md roadmap): nutrition and body (2), plan and proposals (4).
    { path: '/zywienie', name: 'nutrition', ...placeholder('Żywienie') },
    { path: '/cialo', name: 'body', ...placeholder('Ciało') },
    { path: '/plan', name: 'plan', ...placeholder('Plan') },
    { path: '/propozycje', name: 'proposals', ...placeholder('Propozycje') },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})
