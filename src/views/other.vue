<template>
  <div class="list">
    <div class="header">
      <el-input-number
        v-model="state.startnum"
        :min="0"
        @change="handleChange"
      />
    </div>
    <div class="border">
      <div class="row" v-show="false">
        <div
          class="red"
          :class="{ hidden: index == 0 }"
          v-for="(item, index) in 7"
        >
          {{ index }}
        </div>
      </div>
      <div class="cow">
        <div class="cow-red">
          <div class="red" v-for="(item, index) in 33">
            {{ item }}
          </div>
        </div>
        <div class="cow-red" v-for="(item, index) in state.result">
          <div
            class="setred"
            :class="{
              class0: items.order == 0,
              class1: items.order == 1,
              class2: items.order == 2,
              class3: items.order == 3,
              class4: items.order == 4,
            }"
            v-for="(items, indexs) in item"
          >
            <span :class="{hitred: items.hit,}">{{items.rednum}}|{{items.order}}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from "vue";
import axios from "@/utils/axios";
const state = reactive({
  startnum: 0,
  xlsxList: [],
  loading: false,
  redlist: [],
  numlist: [],
  result: [],
  resultblue: [],
  class0:[]
});
onMounted(() => {
  init();
});
const init = () => {
  state.loading = true;
  axios.get("ssq.TXT").then((res) => {
    let list = res.split("\n").reverse();
    let datalist = [];
    list.forEach((item) => {
      let ddlist = item.split(" ");
      datalist.push(ddlist);
    });
    state.xlsxList = datalist;
    let numlist = [];
    state.xlsxList.forEach((item) => {
      numlist.push(item.slice(9, 15));
    });
    setred();
    sampleInit();
    state.loading = false;
  });
};
const reload = () => {
  location.reload();
};

const sample = (list) => {
  let num = Math.floor(Math.random() * list.length);
  state.redlist.push(list[num]);
  if (state.redlist.length < 6) {
    list.splice(num, 1);
    sample(list);
  }
};
const sampleInit = () => {
  let redlist = [];
  let bluelist = Math.floor(Math.random() * 16) + 1;
  for (var i = 1; i <= 33; i++) {
    redlist.push(i);
  }
  state.redlist = [];
  sample(redlist);
  let list = [];
  list.push(state.redlist);
  return list;
};
const setred = () => {
  let redlist = [];
  state.xlsxList.forEach((item) => {
    redlist.push(item.slice(9, 15));
  });
  console.log("zzzzzzzzzzzzz", redlist);
  let num = 1;
  let hitlist = state.startnum == 0 ? [] : redlist[state.startnum - 1];
  var intArray = hitlist.map(function (str) {
    return parseInt(str, 10);
  });
  console.log(intArray);
  state.class0 = []
  getrednum(
    redlist.slice(state.startnum, (33 + state.startnum) * num),
    intArray
  );
};

const getrednum = (redlist, hitlist) => {
  let red1 = [];
  let red2 = [];
  let red3 = [];
  let red4 = [];
  let red5 = [];
  let red6 = [];
  redlist.forEach((item) => {
    red1.push(parseInt(item[0]));
    red2.push(parseInt(item[1]));
    red3.push(parseInt(item[2]));
    red4.push(parseInt(item[3]));
    red5.push(parseInt(item[4]));
    red6.push(parseInt(item[5]));
  });
  state.result = [];
  let result1 = countOccurrences(red1, hitlist[0]);
  let result2 = countOccurrences(red2, hitlist[1]);
  let result3 = countOccurrences(red3, hitlist[2]);
  let result4 = countOccurrences(red4, hitlist[3]);
  let result5 = countOccurrences(red5, hitlist[4]);
  let result6 = countOccurrences(red6, hitlist[5]);
  state.result = [result1, result2, result3, result4, result5, result6];
  let class0 = Array.from(new Set(state.class0))
  class0.sort((a,b)=>a-b)
  console.log(class0);
};

function countOccurrences(arr, hitlist) {
  state.numlist = [];
  for (var i = 0; i < 33; i++) {
    state.numlist.push(0);
  }
  let list = [];
  arr.forEach((item) => {
    state.numlist[item - 1] += 1;
  });
  state.numlist.forEach((item, index) => {
    list.push({
      rednum: index + 1,
      order: item,
      hit: hitlist == index + 1,
      color: 0,
    });
    if(item==0){
      state.class0.push(index + 1)
    }
  });
  list.sort((a, b) => a.order - b.order);
  return list;
}
const handleChange = (num) => {
  setred();
};
</script>

<style lang="less" scoped>
.list {
  // padding: 2vw;
  .border {
    .row {
      display: flex;
      justify-content: space-around;
      font-weight: bold;
      color: #333333;
      .red {
        text-align: center;
        width: 100%;
        background-color: rgb(255, 127, 127);
        border: 1px solid #333333;
      }
      .blue {
        background-color: rgb(127, 131, 255);
      }
      .hidden {
        visibility: hidden;
      }
    }
    .cow {
      display: flex;
      justify-content: space-around;
      .cow-red {
        width: 100%;
      }
      .setred {
        text-align: center;
        span{
          display: inline-block;
        }
        // background-color: aquamarine;
      }
      .class0 {
        background-color: rgba(0, 0, 0, 0.6);
      }
      .class1 {
        background-color: rgba(0, 0, 0, 0.5);
      }
      .class2 {
        background-color: rgba(0, 0, 0, 0.4);
      }
      .class3 {
        background-color: rgba(0, 0, 0, 0.3);
      }
      .class4 {
        background-color: rgba(0, 0, 0, 0.2);
      }

      .hitred {
        width: 60%;
        background-color: red;
      }
    }
  }
}
</style>
