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
            <el-button type="primary" @click="enterpriseClick"
              >ステータス一括設定</el-button
            >
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
          <el-col :span="8" class="mright">
            <el-popover placement="bottom" :width="400" trigger="click">
              <template #reference>
                <el-button>変更実施確認 </el-button>
              </template>
              <div>これまでの変更を反映させますか？</div>
              <el-button>変更を反映する</el-button>
              <el-button>変更棄却</el-button>
              <el-button>キャンセル</el-button>
            </el-popover>
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
      <el-table-column prop="merchantsName" label="JAN Code" width="220">
      </el-table-column>
      <el-table-column prop="merchantsCode" label="商品名"> </el-table-column>
      <el-table-column prop="merchantsSet" label="対象加盟店設定">
        <template #default="scope">
          <label>有效:</label>
          <span>{{ scope.row.merchantsSet }}</span>
          <a class="mlift" @click="enterpriseClick">詳細設定</a>
        </template>
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
    <DialogCommodity ref="commodity"></DialogCommodity>
    <DialogEnterprise ref="enterprise"></DialogEnterprise>
  </el-card>
</template>

<script setup>
import { reactive, ref } from "vue";
import { Plus, Delete, Search, Upload } from '@element-plus/icons-vue'
import axios from "@/utils/axios";
import DialogCommodity from "@/components/DialogCommodity.vue";
import DialogEnterprise from "@/components/dialogEnterprise.vue";
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
const commodity = ref(null);
const enterprise = ref(null);
const changePage = (val) => {
  state.currentPage = val;
};
// 修改分类
const handleEdit = (data) => {
  commodity.value.state.type = "edit";
  commodity.value.state.ruleForm = data;
  commodity.value.state.visible = true;
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
  commodity.value.state.type = "add";
  commodity.value.state.visible = true;
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
