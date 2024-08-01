import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import Books from '../views/Books.vue'
import Search from '../views/Search.vue'
import Setting from '../views/Setting.vue'
import Detail from '../views/Detail.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', name: 'home', component: Home },
    { path: '/books', component: Books },
    { path: '/search', component: Search },
    { path: '/setting', component: Setting },
    { path: '/detail/:id', component: Detail, props: true }
  ]
})

export default router
