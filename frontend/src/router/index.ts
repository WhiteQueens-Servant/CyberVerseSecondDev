import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  scrollBehavior(_to, _from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    }

    return { left: 0, top: 0 }
  },
  routes: [
    {
      path: '/',
      redirect: '/scenic',
    },
    // Tourist-facing routes
    {
      path: '/scenic',
      name: 'scenic-landing',
      component: () => import('../pages/ScenicLandingPage.vue'),
    },
    {
      path: '/scenic/attraction/:id',
      name: 'scenic-attraction-detail',
      component: () => import('../pages/AttractionDetailPage.vue'),
    },
    {
      path: '/scenic/route/:id',
      name: 'scenic-route-detail',
      component: () => import('../pages/RouteDetailPage.vue'),
    },
    {
      path: '/session/:id',
      name: 'session',
      component: () => import('../pages/SessionPage.vue'),
    },
    {
      path: '/kanshan',
      name: 'kanshan-landing',
      component: () => import('../pages/KanshanLandingPage.vue'),
    },
    // Admin portal (sidebar layout)
    {
      path: '/admin',
      component: () => import('../layouts/AdminLayout.vue'),
      children: [
        {
          path: '',
          redirect: '/admin/dashboard',
        },
        {
          path: 'dashboard',
          name: 'admin-dashboard',
          component: () => import('../pages/DashboardPage.vue'),
        },
        {
          path: 'characters',
          name: 'admin-characters',
          component: () => import('../pages/CharacterListPage.vue'),
        },
        {
          path: 'characters/new',
          name: 'admin-character-create',
          component: () => import('../pages/CharacterEditPage.vue'),
        },
        {
          path: 'characters/:id/edit',
          name: 'admin-character-edit',
          component: () => import('../pages/CharacterEditPage.vue'),
        },
        {
          path: 'attractions',
          name: 'admin-attractions',
          component: () => import('../pages/AttractionManagePage.vue'),
        },
        {
          path: 'routes',
          name: 'admin-routes',
          component: () => import('../pages/RouteManagePage.vue'),
        },
        {
          path: 'sessions',
          name: 'admin-sessions',
          component: () => import('../pages/SessionLogPage.vue'),
        },
        {
          path: 'reports',
          name: 'admin-reports',
          component: () => import('../pages/ReportPage.vue'),
        },
        {
          path: 'settings',
          name: 'admin-settings',
          component: () => import('../pages/SettingsPage.vue'),
        },
        {
          path: 'launch/:id',
          name: 'admin-launch',
          component: () => import('../pages/LaunchConfigPage.vue'),
        },
      ],
    },
  ],
})

export default router
