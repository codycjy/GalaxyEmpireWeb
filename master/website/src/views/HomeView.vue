<template>
  <div class="home-view">
    <el-container>
      <el-aside width="200">
        <div class="home-menu">
          <div class="open-menu">
              <i :class="isCollapse ? 'el-icon-d-arrow-right' : 'el-icon-d-arrow-left'"
                 @click="isCollapse = !isCollapse"></i>
            </div>
          <el-menu
            router
            :default-active="this.$route.path"
            class="el-menu-vertical-demo"
            @open="handleOpen"
            @close="handleClose"
            :collapse="isCollapse">
              <el-submenu index="1">
                <template slot="title">
                  <i class="el-icon-location"></i>
                  <span slot="title">账号管理</span>
                </template>
                <el-menu-item index="/account/list">账户列表</el-menu-item>
              </el-submenu>
              <el-submenu index="2">
                <template slot="title">
                  <i class="el-icon-location"></i>
                  <span slot="title">其他功能</span>
                </template>
                <el-menu-item index="#">其他</el-menu-item>
              </el-submenu>
            </el-menu>
        </div>
      </el-aside>
      <el-container>
        <el-header>导航</el-header>
        <el-main>
          <router-view/>
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>
<script>
export default {
  data () {
    return {
      isCollapse: false
    }
  },
  methods: {
    handleOpen (key, keyPath) {
      console.log(key, keyPath)
    },
    handleClose (key, keyPath) {
      console.log(key, keyPath)
    }
  }
}
</script>
<style lang='less' scoped>
  .el-header{
    background-color: #B3C0D1;
    color: #333;
    text-align: center;
    line-height: 60px;
  }
  .el-aside {
    color: #333;
    text-align: center;
    line-height: 200px;
    height: auto;
  }
.el-menu-vertical-demo:not(.el-menu--collapse) {
  width: 200px;
  height: 100%;
  text-align: left;
}
.home-menu {
  position: relative;
  height: 100vh;
}
.open-menu {
  position: absolute;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
  z-index: 1;
  cursor: pointer;
}
</style>
