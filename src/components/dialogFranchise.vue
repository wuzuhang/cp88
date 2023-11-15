<template>
  <el-dialog
    :title="state.type == 'add' ? '新規追加' : '編集'"
    v-model="state.visible"
    width="600px"
  >
    <el-form
      :model="state.ruleForm"
      :rules="state.rules"
      ref="formRef"
      label-width="150px"
      class="good-form"
    >
      <el-form-item label="企業コード" prop="loginUserName">
        <el-input
          type="text"
          v-model="state.ruleForm.loginUserName"
        ></el-input>
        </el-form-item>
        <el-form-item label="企業名" prop="loginUserName">
        <el-input
          type="text"
          v-model="state.ruleForm.loginUserName"
        ></el-input>
        </el-form-item>
    </el-form>
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
    type:"add",
    visible:false,
    ruleForm:{
        loginUserName:""
    },
    rules:[]
})
const submitForm = ()=>{

}
defineExpose({state})
</script>
  
