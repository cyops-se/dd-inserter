// Imports
import Vue from 'vue'
import Router from 'vue-router'
// import { trailingSlash } from '@/util/helpers'
import {
  layout,
  route,
} from '@/util/routes'

Vue.use(Router)

const router = new Router({
  mode: 'history',
  // base: '/admin/', // process.env.BASE_URL,
  base: '/ui', // process.env.BASE_URL,
  scrollBehavior: (to, from, savedPosition) => {
    if (to.hash) return { selector: to.hash }
    if (savedPosition) return savedPosition

    return { x: 0, y: 0 }
  },
  routes: [
    layout('Default', [
      route('Dashboard', null, '/'),

      // Pages
      route('Server Table', null, 'pages/servers'),
      route('Tag Browser', null, 'pages/browse/:serverid'),
      route('Group Table', null, 'pages/groups'),
      route('Tag Table', null, 'pages/tags'),
      route('System Settings', null, 'pages/systemsettings'),
      route('Listener Table', null, 'pages/listeners'),
      route('EmitterTable', null, 'pages/emitters'),
      route('Import Meta', null, 'pages/importmeta'),
      route('Monitoring', null, 'pages/monitoring'),

      // Tables
      route('Logs', null, 'tables/logs'),
      route('Users Table', null, 'tables/users'),
    ]),
    layout('Login', [

      // Pages
      route('Login', null, 'auth/login'),
    ]),
    layout('Logout', [

      // Pages
      route('Logout', null, 'auth/logout'),
    ]),
    layout('Register', [

      // Pages
      route('Register', null, 'auth/register'),
    ]),
  ],
})

router.beforeEach((to, from, next) => {
  return next() // to.path.endsWith('/') ? next() : next(trailingSlash(to.path))
})

export default router
