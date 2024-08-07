<template>
  <el-menu class="hidden-sm-and-down" default-active="1" @select="handleSelect" :router="true">
    <el-menu-item index="/">
      <el-icon>
        <House />
      </el-icon>
      <span>Home</span>
    </el-menu-item>
    <el-menu-item index="/books">
      <el-icon>
        <Files />
      </el-icon>
      <span>Books</span>
    </el-menu-item>
    <el-menu-item index="/search">
      <el-icon>
        <Search />
      </el-icon>
      <span>Search</span>
    </el-menu-item>
    <el-menu-item index="/setting">
      <el-icon>
        <Setting />
      </el-icon>
      <span>Setting</span>
    </el-menu-item>
  </el-menu>
  <el-menu
    class="hidden-md-and-up"
    style="width: 100%"
    default-active="1"
    @select="handleSelect"
    :router="true"
    mode="horizontal"
  >
    <el-menu-item index="/">
      <el-icon>
        <House />
      </el-icon>
      <!--      <span>Home</span>-->
    </el-menu-item>
    <el-menu-item index="/books">
      <el-icon>
        <Files />
      </el-icon>
      <!--      <span>Books</span>-->
    </el-menu-item>
    <el-menu-item index="/search">
      <el-icon>
        <Search />
      </el-icon>
      <!--      <span>Search</span>-->
    </el-menu-item>
    <el-menu-item index="/setting">
      <el-icon>
        <Setting />
      </el-icon>
      <!--      <span>Setting</span>-->
    </el-menu-item>
  </el-menu>
</template>

<script lang="ts">
export default {
  name: 'Sidebar',
  methods: {
    handleSelect(key: string, keyPath: string) {
      console.log(key, keyPath)
    }
  }
}
</script>

<style scoped>
.el-menu {
  width: 100%;
}

@media (max-width: 768px) {
  .el-menu {
    width: 100%;
  }
}
</style>
