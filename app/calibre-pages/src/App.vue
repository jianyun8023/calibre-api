<template>
  <el-container>
    <el-header>
      <Header/>

    </el-header>
    <el-container>
      <el-row>
        <el-col :span="4" :sm="24" :xs="24" :lg="2">
          <el-aside class="sidebar">
            <Sidebar/>
          </el-aside>
        </el-col>
        <el-col :span="20" :sm="24" :xs="24" :lg="22">
          <el-container>
            <el-main>
              <router-view></router-view>
            </el-main>
          </el-container>
        </el-col>
      </el-row>
    </el-container>
    <el-footer>
      <Footer/>
    </el-footer>
  </el-container>
</template>

<script>
import Header from './components/Header.vue'
import Sidebar from './components/Sidebar.vue'
import Footer from './components/Footer.vue'

export default {
  name: 'App',
  components: {
    Header,
    Sidebar,
    Footer
  },
  data() {
    return {
      isSidebarVisible: true,
      isPhone: document.documentElement.clientWidth < 993
    }
  },
  mounted() {
    window.addEventListener('resize', () => {
      this.isPhone = document.documentElement.clientWidth < 993 // 小于993视为平板及手机
      console.log("isPhone: " + this.isPhone)
    })
  },
  methods: {
    toggleSidebar() {
      this.isSidebarVisible = !this.isSidebarVisible
    }
  }
}
</script>

<style scoped>
.sidebar {
  display: block;
  --el-aside-width: 100%;
}

@media (max-width: 768px) {
  .sidebar {
    display: block;
    height: 100px;
    --el-aside-width: 100%;
  }
}
</style>
