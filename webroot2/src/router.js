import { createRouter, createWebHistory } from 'vue-router';

import MileStone from './components/MileStone.vue'

export default createRouter({
  history : createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'milestone',
      component: MileStone
    },
    {
      path: '/milestone',
      name: 'milestone',
      component: MileStone
    },
  ]
})
