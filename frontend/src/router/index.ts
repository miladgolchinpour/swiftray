import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/dashboard',
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('@/views/DashboardView.vue'),
    },
    {
      path: '/subscriptions',
      name: 'subscriptions',
      component: () => import('@/views/SubscriptionsView.vue'),
    },
    {
      path: '/nodes',
      name: 'nodes',
      component: () => import('@/views/NodesView.vue'),
    },
    {
      path: '/routing',
      name: 'routing',
      component: () => import('@/views/RoutingView.vue'),
    },
    {
      path: '/dns',
      name: 'dns',
      component: () => import('@/views/DnsView.vue'),
    },
    {
      path: '/logs',
      name: 'logs',
      component: () => import('@/views/LogsView.vue'),
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('@/views/SettingsView.vue'),
    },
  ],
})

export default router
