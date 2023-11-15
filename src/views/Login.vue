<template>
  <div class="login-body">
    <div class="login-container">
      <div class="head">
        <!-- <img class="logo" src="https://s.yezgea02.com/1582958061265/mlogo.png" /> -->
        <div class="name">
          <div class="title">Star-Coupon</div>
          <div class="tips">商户推广平台</div>
        </div>
      </div>
      <el-form
        label-position="top"
        :rules="state.rules"
        :model="state.ruleForm"
        ref="loginForm"
        class="login-form"
      >
        <el-form-item label="账号" prop="username">
          <el-input
            type="text"
            v-model.trim="state.ruleForm.username"
            autocomplete="off"
          ></el-input>
        </el-form-item>
        <el-form-item label="密码" prop="password" style="margin-top: 16px">
          <el-input
            type="password"
            v-model.trim="state.ruleForm.password"
            autocomplete="off"
          ></el-input>
        </el-form-item>
        <el-form-item
          label="验证码"
          prop="verify"
          style="margin-top: 16px; width: 50%"
        >
          <el-input
            :clearable="true"
            v-model.trim="state.ruleForm.verify"
            autocomplete="off"
          ></el-input>
          <div class="verifyImg">
            <el-button
              style="width: 100px"
              type="primary"
              :disabled="state.disabled"
              @click="sendEmail"
              >{{ state.valiemail }}</el-button
            >
          </div>
        </el-form-item>
        <el-form-item> &nbsp; </el-form-item>
        <el-form-item>
          <el-button style="width: 100%" type="primary" @click="submitForm"
            >立即登录</el-button
          >
          <!-- <el-checkbox v-model="state.checked" @change="!state.checked">下次自动登录</el-checkbox> -->
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import axios from "@/utils/axios";
import md5 from "js-md5";
import { reactive, ref } from "vue";
import { localSet } from "@/utils";
const loginForm = ref(null);
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";

const router = useRouter();
const state = reactive({
  ruleForm: {
    username: "",
    password: "",
    verify: "",
  },
  verifyImg: "",
  checked: true,
  rules: {
    username: [{ required: "true", message: "账户不能为空", trigger: "blur" }],
    password: [{ required: "true", message: "密码不能为空", trigger: "blur" }],
    verify: [{ required: "true", message: "验证码不能为空", trigger: "blur" }],
  },
  loading: false,
  Capt: "",
  valiemail: "获取验证码",
  disabled: false,
  LoginEmail: "",
  submitLogin:null
});

const submit = () => {
  loginForm.value.validate((valid) => {
    if (valid) {
      axios
        .post("/adminUser/login", {
          userName: state.ruleForm.username || "",
          passwordMd5: md5(state.ruleForm.password),
          EmailValue: state.ruleForm.verify,
        })
        .then((res) => {
          localSet("username", state.ruleForm.username);
          localSet("token", res);
          if (state.ruleForm.username == "admin") {
            router.push({ path: "/user" });
          } else {
            router.push({ path: "/" });
          }
        }).catch(err=>{

        });
    } else {
      console.log("error submit!!");
      return false;
    }
  });
};
const submitForm = ()=>{
  _debounce(submit,1000)()
}
const _debounce = (func, delay) => {
  // let timer;  // 定时器

    return function () { 
        let context = this;  // 记录 this 值,防止在回调函数中丢失
        let args = arguments;  // 函数参数
        
        // 标识是否立即执行
        let isImmediately = !state.submitLogin;

        //如果定时器存在，则清除定时器(如果没有,也没必要进行处理)
        state.submitLogin ? clearTimeout(state.submitLogin) : null; 

        state.submitLogin = setTimeout(() => { 
          state.submitLogin = null;
        }, delay);
        
        // isImmediately 为 true 则 执行函数(即首次触发事件)
        isImmediately ? func.apply(context, args) : null;
    } 
};
const resetForm = () => {
  loginForm.value.resetFields();
};
const sendEmail = () => {
  if (state.ruleForm.username == "") {
    ElMessage.error("请输入账号");
    return;
  }
  state.disabled = true;
  axios
    .post("/user/captcha", { LoginUserName: state.ruleForm.username })
    .then((res) => {
        state.valiemail = 60;
        let send = setInterval(() => {
          state.valiemail = state.valiemail - 1;
          if (state.valiemail == 0) {
            state.valiemail = "获取验证码";
            state.disabled = false;
            clearInterval(send);
          }
        }, 1000);
      
    })
    .catch((err) => {
      // ElMessage.error('发送失败，请重试')
        state.disabled = false;
    });
};
// getVerify();
// const sendEmail = () => {
//   getVerify()

// };
</script>

<style scoped>
.login-body {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  background-color: #fff;
}
.login-container {
  width: 420px;
  height: 440px;
  background-color: #fff;
  border-radius: 4px;
  box-shadow: 0px 21px 41px 0px rgba(0, 0, 0, 0.2);
}
.head {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40px 0 20px 0;
}
.head img {
  width: 100px;
  height: 100px;
  margin-right: 20px;
}
.head .title {
  font-size: 28px;
  color: #1baeae;
  font-weight: bold;
}
.head .tips {
  font-size: 12px;
  color: #999;
}
.login-form {
  width: 70%;
  margin: 0 auto;
}
.login-form >>> .el-form--label-top .el-form-item__label {
  padding: 0;
}
.login-form >>> .el-form-item {
  margin-bottom: 0;
}
.verifyImg {
  position: absolute;
  right: -100%;
  cursor: pointer;
  /* margin-top: 5px; */
}
</style>