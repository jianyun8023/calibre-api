<template>
  <el-container>
    <el-header>
      <SiteHeader />
    </el-header>
    <el-container>
      <el-row style="width: 100%">
        <el-col class="sidebar-parent" :span="4" :sm="24" :xs="24" :lg="2">
          <el-aside class="sidebar">
            <Sidebar />
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
      <SiteFooter />
    </el-footer>
  </el-container>
</template>

<script lang="ts">
import SiteHeader from './components/Header.vue'
import Sidebar from './components/Sidebar.vue'
import SiteFooter from './components/Footer.vue'

export default {
  name: 'App',
  components: {
    SiteHeader,
    Sidebar,
    SiteFooter
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
.sidebar-parent{

}
</style>
