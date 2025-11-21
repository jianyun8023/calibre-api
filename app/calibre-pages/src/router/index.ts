import { createRouter, createWebHistory } from 'vue-router'

// Home 页面保持同步导入（首屏需要）
import Home from '../views/Home.vue'

// 其他页面使用懒加载（代码分割）
const Books = () => import('../views/Books.vue')
const Search = () => import('../views/Search.vue')
const Setting = () => import('../views/Setting.vue')
const Detail = () => import('../views/Detail.vue')
const BatchMeta = () => import('../views/BatchMeta.vue')
const ReadBook = () => import('../views/ReadBook.vue')
const Publisher = () => import('../views/Publisher.vue')
const Tasks = () => import('../views/Tasks.vue')


const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        { path: '/', name: 'home', component: Home },
        { path: '/books', component: Books },
        { path: '/search', component: Search },
        { path: '/setting', component: Setting },
        { path: '/metadata/manager', component: BatchMeta },
        { path: '/detail/:id', component: Detail, props: true },
        { path: '/read/:id', component: ReadBook, props: true },
        { path: '/publisher', component: Publisher },
        { path: '/tasks', component: Tasks },
    ]
})

export default router
