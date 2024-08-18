import {createRouter, createWebHistory} from 'vue-router'
import Home from '../views/Home.vue'
import Books from '../views/Books.vue'
import Search from '../views/Search.vue'
import Setting from '../views/Setting.vue'
import Detail from '../views/Detail.vue'
import BatchMeta from '../views/BatchMeta.vue'
import ReadBook from '../views/ReadBook.vue'
import Publisher from '../views/Publisher.vue'


const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {path: '/', name: 'home', component: Home},
        {path: '/books', component: Books},
        {path: '/search', component: Search},
        {path: '/setting', component: Setting},
        {path: '/metadata/manager', component: BatchMeta},
        {path: '/detail/:id', component: Detail, props: true},
        {path: '/read/:id', component: ReadBook, props: true},
        {path: '/publisher', component: Publisher},
    ]
})

export default router
