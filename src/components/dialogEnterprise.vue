<template>
  <el-dialog
    :title="state.type == 'all' ? 'ステータス一括設定' : '詳細設定'"
    v-model="state.visible"
    width="600px"
  >
    <el-table :data="state.enterprise">
      <el-table-column type="selection" width="55"> </el-table-column>
      <el-table-column property="id" label="企業コード" />
      <el-table-column property="name" label="企業名" />
    </el-table>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="state.visible = false">取 消</el-button>
        <el-button type="primary" @click="submitForm">确 定</el-button>
      </span>
    </template>
  </el-dialog>
</template>
  <script setup>
import { reactive } from "vue";
const state = reactive({
  type: "all",
  visible: false,
  enterprise:[{id:"9000012345",name:"株式会社ABC"},{id:"9000987654",name:"株式会社DEF"}]
});
const submitForm = () => {};
defineExpose({ state });
</script>
  
