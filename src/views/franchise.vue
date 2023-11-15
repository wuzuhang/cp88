<template>
  <el-card class="category-container">
    <template #header>
      <div class="header">
        <el-row>
          <el-col :span="24"
            ><el-input
              style="width: 200px; margin-right: 10px"
              placeholder="キーワード検索"
              clearable
            ></el-input>
            <el-button type="primary" :icon="Search"
              >キーワード検索</el-button
            ></el-col
          >
        </el-row>
        <el-row>
          <el-col :span="16">
            <el-button type="primary" @click="addCommodity">新規追加</el-button>
            <el-popconfirm
              title="确定删除吗？"
              confirmButtonText="确定"
              cancelButtonText="取消"
            >
              <template #reference>
                <el-button type="danger" :icon="Delete">一括削除</el-button>
              </template>
            </el-popconfirm>
          </el-col>
        </el-row>
      </div>
    </template>
    <el-table
      :loading="state.loading"
      :data="state.tableData"
      tooltip-effect="dark"
      style="width: 100%"
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="55"> </el-table-column>
      <el-table-column prop="merchantsName" label="企業コード">
      </el-table-column>
      <el-table-column prop="merchantsCode" label="企業名"> </el-table-column>
      <el-table-column prop="merchantsSet" label="キャンペーン商品数" align="center">
      </el-table-column>
      <el-table-column label="操作" width="100" align="center">
        <template #default="scope">
          <a style="cursor: pointer" @click="handleEdit(scope.row)">編集</a>
          <el-popconfirm
            title="确定删除吗？"
            confirmButtonText="确定"
            cancelButtonText="取消"
          >
            <template #reference>
              <a
                style="cursor: pointer; margin-left: 20px"
                @click="handleDelete(scope.row)"
                >削除</a
              >
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      background
      layout="prev, pager, next"
      :total="state.total"
      :page-size="state.pageSize"
      :current-page="state.currentPage"
      @current-change="changePage"
    />
    <DialogFranchise ref="franchise"></DialogFranchise>
  </el-card>
</template>

<script setup>
import { reactive, ref } from "vue";
import { Plus, Delete, Search, Upload } from '@element-plus/icons-vue'
import axios from "@/utils/axios";
import DialogFranchise from "@/components/dialogFranchise.vue";
const state = reactive({
  loading: false,
  tableData: [
    {
      merchantsName: "4901777235984",
      merchantsCode: "ペプシスペシャルゼロ500",
      merchantsSet: "15",
    },
  ],
  total: 0, // 总条数
  pageSize: 10, // 分页大小
  currentPage: 1, // 当前页
});
const franchise = ref(null);
const enterprise = ref(null);
const changePage = (val) => {
  state.currentPage = val;
};
// 修改分类
const handleEdit = (data) => {
  franchise.value.state.type = "edit";
  franchise.value.state.ruleForm = data;
  franchise.value.state.visible = true;
};
const handleDelete = (data) => {};
// // 选择项
const handleSelectionChange = (val) => {
  state.multipleSelection = val;
};
const enterpriseClick = () => {
  enterprise.value.state.type = "all";
  enterprise.value.state.visible = true;
};
const addCommodity = () => {
  franchise.value.state.type = "add";
  franchise.value.state.visible = true;
};
</script>

<style lang="less" scoped>
.el-row {
  margin-bottom: 20px;
}
.mright {
  text-align: right;
}
.mlift {
  cursor: pointer;
  margin-left: 30px;
}
</style>
