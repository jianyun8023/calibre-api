<template>
  <!--  <el-container class="full-height">-->
  <!--    <el-col :span='24' class="full-height">-->
  <div class="full-height">
    <vue-reader ref="reader" :url="bookUrl"
                :location="location"
                :getRendition="getRendition"
                @update:location="locationChange"
    />
  </div>
  <el-button @click="jumpToPath(initPath)">跳转</el-button>
  <!--    </el-col>-->

  <!--  </el-container>-->
</template>
<script lang="ts">
import {VueReader} from 'vue-reader'
import {useStorage} from '@vueuse/core'

import {ElContainer, ElRow} from "element-plus";

export default {
  name: 'ReadBook',
  components: {ElContainer, ElRow, VueReader},
  data() {
    return {
      bookId: '',
      bookUrl: '',
      initPath: '',
      location: useStorage('book-progress', 0, undefined, {
        serializer: {
          read: (v) => JSON.parse(v),
          write: (v) => JSON.stringify(v),
        },
      }),
      rendition: {}

    }
  },
  created() {
    // this.$refs.reader.getRendition
    this.bookId = (this.$route as any).params.id
    this.bookUrl = `/api/download/book/${this.bookId}.epub`
    if (this.$route.query.path){
      this.initPath = this.$route.query.path as string
    }
  },
  methods: {
    locationChange: (epubcifi) => {
      console.log(epubcifi)
      location.value = epubcifi
    },
    jumpToPath(path: any) {
      console.log(this.rendition)
      this.$refs.reader.setLocation(path);
    },
    getRendition(val) {
      this.rendition = val
      this.rendition.themes.default({
        '::selection': {
          background: 'orange',
        },
      })
      // this.renditionrendition.on('selected', setRenderSelection)
    }
  }
}
</script>
<style scoped>
.full-height {
  height: 90vh;
  display: flex;
  flex-direction: row;

  .container {
    width: 100%;
  }
}

.sidebar {
  display: none !important;
}

.el-aside {
  display: none !important;
}

.sidebar-parent {
  display: none !important;
}


</style>