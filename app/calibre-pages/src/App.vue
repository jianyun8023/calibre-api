<template>
  <el-container>
    <el-header>
      <Header />
    </el-header>
    <el-container>
      <el-aside width="200px" class="sidebar hidden-sm-and-down">
        <Sidebar/>
      </el-aside>
      <el-container>
        <el-main>
          <router-view></router-view>
        </el-main>
      </el-container>
    </el-container>
    <el-footer>
      <Footer />
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
    }
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
}

@media (max-width: 768px) {
  .sidebar {
    display: none;
  }
}
</style>
