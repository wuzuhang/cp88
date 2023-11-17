import { createRouter, createWebHashHistory } from 'vue-router'

const router = createRouter({
  history: createWebHashHistory(), // hash模式：createWebHashHistory，history模式：createWebHistory
  routes: [
    {
      path: '/',
      redirect: '/other'
    },
    // {
    //   path: '/commodity',
    //   name: 'commodity',
    //   component: () => import(/* webpackChunkName: "commodity" */ '../views/commodity.vue')
    // },
    // {
    //   path: '/franchise',
    //   name: 'franchise',
    //   component: () => import(/* webpackChunkName: "franchise" */ '../views/franchise.vue')
    // },
    {
      path: '/other',
      name: 'other',
      component: () => import(/* webpackChunkName: "other" */ '../views/other.vue')
    },
    // {
    //   path: '/login',
    //   name: 'login',
    //   component: () => import(/* webpackChunkName: "login" */ '../views/Login.vue')
    // },
    
  ]
})

export default router