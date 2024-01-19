<template>
  <div class="account-list">
    <el-table
      ref="multipleTable"
      :data="accountDatas"
      tooltip-effect="dark"
      style="width: 100%"
      @selection-change="handleSelectionChange">
      <el-table-column
        type="selection"
        width="120">
      </el-table-column>
      <el-table-column
        prop="username"
        label="用户名"
        width="120">
      </el-table-column>
      <el-table-column
        prop="email"
        label="邮件"
        width="240">
      </el-table-column>
      <el-table-column
        prop="server"
        label="服务"
        width="240"
        show-overflow-tooltip>
      </el-table-column>
    </el-table>
  </div>
</template>
<script>
// import { getAccountsByUserId } from '@/api/account'
export default {
  created () {
    // getAccountsByUserId(this.id).then(response => {
    //   this.accountDatas = response.data.accounts
    // })
  },
  data () {
    return {
      id: '1',
      accountDatas: [
        {
          id: '123',
          username: 'wzflag',
          email: 'wzflag@qq.com',
          server: 'wzf'
        }
      ],
      multipleSelection: []
    }
  },
  methods: {
    handleSelectionChange (val) {
      this.multipleSelection = val
    }
  }
}
</script>
<style lang="less" scoped>
/deep/ .el-table th.el-table__cell>.cell{
  padding: 0 14px;
}
</style>
