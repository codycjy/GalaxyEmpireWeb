import Vue from 'vue'
import VueRouter from 'vue-router'

Vue.use(VueRouter)

const routes = [
  {
    path: '/',
    redirect: '/login'
  },
  {
    path: '/login',
    name: 'login',
    component: () => import(/* webpackChunkName: "login" */ '../views/LoginView.vue')
  },
  {
    path: '/register',
    name: 'register',
    component: () => import(/* webpackChunkName: "register" */ '../views/RegisterView.vue')
  },
  {
    path: '/home',
    name: 'home',
    component: () => import(/* webpackChunkName: "home" */ '../views/HomeView.vue'),
    children: [
      {
        path: '/',
        redirect: '/account/list'
      },
      {
        path: '/account/list',
        component: () => import(/* webpackChunkName: "account-list" */ '../views/account/AccountList.vue')
      }
    ]
  },
  {
    path: '*',
    component: () => import(/* webpackChunkName: "notFound" */ '../views/error/NotFound.vue')
  }
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
})

export default router
