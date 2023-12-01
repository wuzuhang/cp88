<template>
  <div class="list" v-loading="state.loading">
    <div class="header">
      <div class="marginbtn">
        <span>上一期</span>&nbsp;&nbsp;
        <el-input-number
          v-model="state.startnum"
          :min="0"
          @change="handleChange"
        />
        &nbsp;&nbsp;<span>下一期</span>
      </div>
      <!-- <div>{{ frequency / 2 }}</div> -->
      <!-- <div class="marginbtn sample" v-show="state.time">
        {{ state.time }}:
        <span class="redhit" v-for="item in state.redhitd">{{ item }}</span>
        <span class="bluehit">{{ state.bluehitd }}</span>
      </div> -->
      <el-button type="primary" class="marginbtn" @click="sampleInit"
        >随机生成开奖号码</el-button
      >
      <div class="marginbtn sample" v-show="state.redlistSort.length > 0">
        <span class="redhit" v-for="item in state.redlistSort">{{ item }}</span>
        <span class="bluehit">{{ state.bluehit }}</span>
      </div>
      <el-button type="primary" class="marginbtn" @click="reload"
        >重置</el-button
      >
    </div>
    <div class="border">
      <div class="row">
        <div
          class="red"
          :class="{ blue: index == 7, hidden: index == 0 }"
          v-for="(item, index) in 8"
        >
          {{ index }}
        </div>
      </div>
      <div class="cow">
        <div class="cow-red redNo">
          <div
            class="red"
            v-for="(item, index) in 33"
            :class="{
              redNohit: setnum(state.allredlist, state.redhitd, item).hit,
            }"
          >
            {{ item }}-<span style="color: #dddddd"
              >{{ setnum(state.allredlist, state.redhitd, item).num }}
            </span>
          </div>
        </div>
        <div class="cow-red rednum" v-for="(item, index) in state.result">
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
            @click="numclick(items)"
          >
            <span :class="{ hitred: items.hit }">{{ items.rednum }}</span>
          </div>
        </div>
        <div class="cow-red rednum">
          <div
            class="setred"
            v-for="(items, indexs) in state.resultblue"
            :class="{
              class0: items.order == 0,
              class1: items.order == 1,
              class2: items.order == 2,
              class3: items.order == 3,
              class4: items.order == 4,
            }"
            @click="numclick(items)"
          >
            <span :class="{ hitblue: items.hit }">{{ items.bluenum }}</span>
          </div>
        </div>
        <!-- <div class="cow-red blueNo">
          <div
            class="blue"
            v-for="(item, index) in 16"
            :class="{ blueNohit: setnum(item).hit }"
          >
            {{ item }}-<span style="color: #dddddd"
              >{{ setnum(item).num }}
            </span>
          </div>
        </div> -->
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted, computed } from "vue";
import axios from "@/utils/axios";
const state = reactive({
  startnum: 0,
  xlsxList: [],
  loading: false,
  redlist: [],
  redlistSort: [],
  numlist: [],
  result: [],
  resultblue: [],
  class0: [],
  bluehit: "",
  time: "",
  redhitd: [],
  bluehitd: "",
  allredlist: []
});
onMounted(() => {
  init();
});
const setfrequency = ()=>{
   if (Array.isArray(allredlist) && allredlist.length > 0) {
    allredlist.forEach((item) => {
      if (item.rednum == data) {
        obj.num = item.order;
      }
    });
  }

}
const setnum = computed(() => (allredlist, redhitd, data) => {
  let obj = { num: 0, hit: false };
  if (Array.isArray(allredlist) && allredlist.length > 0) {
    allredlist.forEach((item) => {
      if (item.rednum == data) {
        obj.num = item.order;
      }
    });
  }
  if (Array.isArray(redhitd) && redhitd.length > 0) {
    redhitd.forEach((item) => {
      if (item == data) {
        obj.hit = true;
      }
    });
  }
  return obj;
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
  state.time = "";
  state.startnum = 0;
  setred();
  let redlist = [];
  let bluelist = Math.floor(Math.random() * 16) + 1;
  for (var i = 1; i <= 33; i++) {
    redlist.push(i);
  }
  state.redlist = [];
  state.bluehit = bluelist;
  sample(redlist);
  state.redlistSort = JSON.parse(JSON.stringify(state.redlist)).sort(
    (a, b) => a - b
  );
  state.redhitd = state.redlistSort;
  state.result.forEach((item, index) => {
    item.forEach((items) => {
      items.hit = false;
      if (items.rednum == state.redlist[index]) {
        items.hit = true;
      }
      // if (state.redlist.includes(items.rednum)) {
      //   items.hit = true;
      // }
    });
  });
  state.resultblue.forEach((item) => {
    item.hit = false;
    if (item.bluenum == bluelist) {
      item.hit = true;
    }
  });
};
const setred = () => {
  let redlist = [];
  state.xlsxList.forEach((item) => {
    redlist.push(item.slice(9, 15));
  });
  let num = 1;
  let hitlist = state.startnum == 0 ? [] : redlist[state.startnum - 1];
  var intArray = hitlist.map(function (str) {
    return parseInt(str, 10);
  });
  state.redhitd = JSON.parse(JSON.stringify(intArray)).sort((a, b) => a - b);

  state.class0 = [];
  getrednum(
    redlist.slice(state.startnum, (33 + state.startnum) * num),
    intArray
  );
};

const getrednum = (redlist, hitlist) => {
  let allredlist = [];
  let red1 = [];
  let red2 = [];
  let red3 = [];
  let red4 = [];
  let red5 = [];
  let red6 = [];
  redlist.forEach((item) => {
    allredlist = allredlist.concat(item);
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

  let bluelist = [];
  state.xlsxList.forEach((item) => {
    bluelist.push(item[8]);
  });
  let num = 1;
  let resultblue = countOccurrencesblue(
    bluelist.slice(state.startnum, (16 + state.startnum) * num)
  );
  state.resultblue = resultblue;

  state.allredlist = countOccurrences(allredlist);
  let class0 = Array.from(new Set(state.class0));
  class0.sort((a, b) => a - b);
};

function countOccurrences(arr, hitlist) {
  // console.log(hitlist.length)
  // let hit_list = []
  // if(Array.isArray(hitlist)&&hitlist.length>0){
  //    hit_list = hitlist
  // }
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
    if (item == 0) {
      state.class0.push(index + 1);
    }
  });
  list.sort((a, b) => a.order - b.order);
  return list;
}
function countOccurrencesblue(arr) {
  let hitlist =
    state.startnum == 0 ? 0 : parseInt(state.xlsxList[state.startnum - 1][8]);
  state.bluehitd = hitlist;
  state.numlist = [];
  for (var i = 0; i < 16; i++) {
    state.numlist.push(0);
  }
  let list = [];
  arr.forEach((item) => {
    state.numlist[item - 1] += 1;
  });
  state.numlist.forEach((item, index) => {
    list.push({
      bluenum: index + 1,
      order: item,
      hit: hitlist == index + 1,
      color: 0,
    });
    if (item == 0) {
      state.class0.push(index + 1);
    }
  });
  list.sort((a, b) => a.order - b.order);
  return list;
}
const handleChange = (num) => {
  if (num == 0) {
    state.time = "";
  } else {
    state.time = state.xlsxList[state.startnum - 1][1];
  }
  state.redlistSort = [];
  setred();
};
const numclick = (a) => {
  a.hit = !a.hit;
};
</script>

<style lang="less" scoped>
* {
  box-sizing: border-box;
  font-size: 14px;
}
.list {
  .header {
    display: flex;
    flex-direction: column;
    align-items: center;
    .marginbtn {
      margin-bottom: 10px;
    }
    .sample {
      .redhit {
        display: inline-block;
        width: 26px;
        height: 26px;
        background-color: rgb(245, 56, 56);
        color: #ffffff;
        border-radius: 50%;
        margin-right: 5px;
        text-align: center;
        line-height: 26px;
      }
      .bluehit {
        display: inline-block;
        width: 26px;
        height: 26px;
        background-color: rgb(73, 79, 249);
        color: #ffffff;
        border-radius: 50%;
        text-align: center;
        line-height: 26px;
      }
      .rednum {
        display: inline-block;
        width: 26px;
        height: 26px;
        color: #333333;
        margin-right: 5px;
        text-align: center;
        line-height: 26px;
      }
    }
  }
  .border {
    .row {
      display: flex;
      justify-content: center;
      font-weight: bold;
      color: #333333;
      .red {
        text-align: center;
        width: 12vw;
        height: 6vw;
        line-height: 6vw;
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
      justify-content: center;
      .cow-red {
        .red {
          text-align: center;
          width: 12vw;
          height: 6vw;
          line-height: 6vw;
        }
      }
      .redNo {
        background-color: rgb(255, 127, 127);
        // font-weight: bold;
        color: #333333;
        border: 1px solid #333333;
        .red {
          width: 12vw;
          height: 6vw;
          line-height: 6vw;
          border-bottom: 1px solid #333333;
        }
        .redNohit {
          background-color: rgb(245, 56, 56);
        }
      }
      .blueNo {
        background-color: rgb(127, 131, 255);
        font-weight: bold;
        color: #333333;
        border: 1px solid #333333;
        .blue {
          width: 12vw;
          height: 6vw;
          line-height: 6vw;
          border-bottom: 1px solid #333333;
        }
        .blueNohit {
          background-color: rgb(73, 79, 249);
        }
      }
      .setred {
        text-align: center;
        width: 12vw;
        height: 6vw;
        line-height: 6vw;
        span {
          display: inline-block;
        }
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
        width: 50%;
        background-color: rgb(245, 56, 56);
        color: #ffffff;
        font-weight: bold;
        border-radius: 50%;
      }
      .hitblue {
        width: 50%;
        background-color: rgb(73, 79, 249);
        color: #ffffff;
        font-weight: bold;
        border-radius: 50%;
      }
    }
  }
}
</style>
