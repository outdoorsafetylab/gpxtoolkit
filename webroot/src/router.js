import { createRouter, createWebHistory } from 'vue-router';

import MileStone from './components/MileStone.vue'
import HelloWorld from './components/HelloWorld.vue'

export default createRouter({
  history : createWebHistory(),
  routes: [
    {
      path: '/',
      component: MileStone
    },
    {
      path: '/milestone',
      component: MileStone
    },
    {
      path: '/hello',
      component: HelloWorld
    },
  ]
})
