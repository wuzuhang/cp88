<template>
  <el-card class="category-container" v-loading="state.loading">
    <template #header>
      <div class="header">
        <!-- <el-upload class="upload-demo" :http-request="readxls" style="margin-bottom:10px">
          <el-button type="primary">Click to upload</el-button>
        </el-upload> -->
        <!-- <div>
          <label class="myLabel">样本数量：</label>
          <el-select
            style="margin-left: 20px; margin-bottom: 10px"
            v-model="state.option"
            placeholder="Select"
          >
            <el-option
              v-for="item in state.options"
              :key="item"
              :label="item"
              :value="item"
            />
          </el-select>
        </div> -->
        <div>
          <label class="myLabel">最近期次开奖结果：</label>
          <el-input-number
            style="width: 175px; margin-bottom: 10px"
            v-model="state.startNum"
          />
        </div>

        <!-- <el-button
          @click="setUp(state.option, state.startNum)"
          style="margin-left: 20px; margin-bottom: 10px"
          type="primary"
          >单样本计算</el-button
        > -->
        <el-button
          @click="random"
          style="margin-left: 20px; margin-bottom: 10px"
          type="primary"
          >随机生成开奖结果</el-button
        >
        <el-button
          @click="setAll(state.startNum)"
          style="margin-left: 20px; margin-bottom: 10px"
          type="primary"
          >全样本计算</el-button
        >
        <!-- <el-button
          @click="barList"
          style="margin-left: 20px; margin-bottom: 10px"
          type="primary"
          >柱状图</el-button
        > -->

        <el-button
          @click="reload"
          style="margin-left: 20px; margin-bottom: 10px"
          type="primary"
          >重置</el-button
        >
        <!-- @click="setUp(state.option, state.startNum)" -->
        <div class="dataTime">期次：{{ state.time }}</div>
        <div class="dataTime">
          命中数量：{{ state.rednum }}-{{ state.bluenum }}
        </div>
      </div>
    </template>
    <div class="nums">
      <div class="red">
        <div
          v-for="(item, index) in state.red"
          class="allred"
          @click="redclick(item)"
        >
          <span
            class="nonum"
            :class="{ redbacka: item.background % 2 === 0 }"
            >{{ index + 1 }}</span
          >
          <span
            class="redspan"
            :class="{
              isred: state.redyes.includes(item.value) || item.clicked,
            }"
            >{{ item.value }}&nbsp;&nbsp;&nbsp;&nbsp;{{ item.abs }} 次</span
          >
        </div>
      </div>
      <div class="blue">
        <div
          v-for="(item, index) in state.blue"
          class="allblue"
          @click="redclick(item)"
        >
          <span
            class="nonum"
            :class="{ bluebacka: item.background % 2 === 0 }"
            >{{ index + 1 }}</span
          ><span
            class="bluespan"
            :class="{
              isblue: item.value == state.blueyes || item.clicked,
            }"
            >{{ item.value }}{{ item.abs }} 次</span
          >
        </div>
      </div>
    </div>
    <div ref="redbar" class="bars"></div>
    <div ref="bluebar" class="bars"></div>
  </el-card>
</template>

<script setup>
import { reactive, ref, onMounted } from "vue";
import axios from "@/utils/axios";
import * as echarts from "echarts";
import { read, utils, writeFile } from "xlsx";
import { ElMessage, ElMessageBox } from "element-plus";
import { expandColumnKey } from "element-plus/es/components/table-v2/src/common";
const state = reactive({
  xlsxList: [],
  option: 100,
  options: [30, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000],
  showList: [],
  time: "",
  setredlist: [],
  setbluelist: [],
  redFooter: [],
  blueFooter: [],
  red: [],
  blue: [],
  redyes: [],
  blueyes: [],
  startNum: 0,
  redbar: [],
  bluebar: [],
  allred: 0,
  loading: false,
  rednum: 0,
  bluenum: 0,
});
const redbar = ref(null);
const bluebar = ref(null);
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
    setAll(state.startNum);
    state.loading = false;
  });
};
const setUp = (option, num) => {
  if (num > 0) {
    const data = state.xlsxList[num - 1];
    state.redyes = [
      parseFloat(data[2]),
      parseFloat(data[3]),
      parseFloat(data[4]),
      parseFloat(data[5]),
      parseFloat(data[6]),
      parseFloat(data[7]),
    ];
    state.blueyes = parseFloat(data[8]);
  } else {
    state.redyes = [];
    state.blueyes = "";
  }
  state.showList = [];
  state.setredlist = [];
  state.setbluelist = [];
  state.redFooter = [];
  state.blueFooter = [];
  for (var i = 1; i <= 33; i++) {
    state.setredlist.push(0);
    state.redFooter.push(0);
  }
  for (var i = 1; i <= 16; i++) {
    state.setbluelist.push(0);
    state.blueFooter.push(0);
  }
  let datalist = state.xlsxList.slice(num, option + num).reverse();
  // 1 101
  state.time = datalist[0][1] + "------" + datalist[datalist.length - 1][1];
  datalist.forEach((item, index) => {
    const red = [
      parseFloat(item[2]),
      parseFloat(item[3]),
      parseFloat(item[4]),
      parseFloat(item[5]),
      parseFloat(item[6]),
      parseFloat(item[7]),
    ];
    const blue = [parseFloat(item[8])];
    let num = [];
    for (var i = 1; i <= 33; i++) {
      if (red.includes(i)) {
        state.setredlist[i - 1] = 0;
        state.redFooter[i - 1] += 1;
        num.push({ value: i, class: "red", select: true });
      } else {
        state.setredlist[i - 1] += 1;
        num.push({
          value: state.setredlist[i - 1],
          class: "red",
          select: false,
        });
      }
    }
    for (var i = 1; i <= 16; i++) {
      if (blue.includes(i)) {
        state.setbluelist[i - 1] = 0;
        state.blueFooter[i - 1] += 1;
        num.push({ value: i, class: "blue", select: true });
      } else {
        state.setbluelist[i - 1] += 1;
        num.push({
          value: state.setbluelist[i - 1],
          class: "blue",
          select: false,
        });
      }
    }
    state.showList.push(num);
  });
  let r0 = setNum(state.redFooter.slice(0, 11), 1);
  let r1 = setNum(state.redFooter.slice(11, 22), 12);
  let r2 = setNum(state.redFooter.slice(22, 33), 23);
  let b0 = setNum(state.blueFooter, 1);
  // let red = r0.concat(r1).concat(r2);
  let red = setNum(state.redFooter, 1);
  red.sort((a, b) => {
    return a.abs - b.abs;
  });
  b0.sort((a, b) => {
    return a.abs - b.abs;
  });
  state.allred = 0;
  red.forEach((item) => {
    if (state.redyes.includes(item.value)) {
      state.allred += item.abs;
    }
  });
  state.red = red;
  state.blue = b0;
  return { red: red, blue: b0 };
};
const setNum = (as, start) => {
  let a = JSON.parse(JSON.stringify(as));
  a.sort((a, b) => {
    return a - b;
  });
  let a0 = a[0];
  let a1 = a[a.length - 1];
  let min = 0;
  let max = 0;
  let num = 0;
  let all = 0;
  a.forEach((item) => {
    all += item;
    if (a0 == item) {
      min += item;
      num += 1;
    }
    if (a1 == item) {
      max += item;
      num += 1;
    }
  });
  const average = (all - min - max) / (a.length - num);
  let absnum = [];
  as.forEach((item, index) => {
    let absObj = { value: index + start, abs: 0, clicked: false };
    absObj.abs = Math.abs(item - average);
    absnum.push(absObj);
  });
  return absnum;
};
const barList = () => {
  const titlered = [];
  const titleblue = [];
  for (var i = 0; i < 33; i++) {
    titlered.push(i + 1);
    state.redbar.push(0);
  }
  for (var i = 0; i < 16; i++) {
    titleblue.push(i + 1);
    state.bluebar.push(0);
  }
  for (var i = 1; i <= state.startNum; i++) {
    const data = setUp(state.option, i);
    data.red.forEach((item, index) => {
      if (state.redyes.includes(item.value)) {
        state.redbar[index] += 1;
      }
    });
    data.blue.forEach((item, index) => {
      if (state.blueyes == item.value) {
        state.bluebar[index] += 1;
      }
    });
  }
  var myChartred = echarts.init(redbar.value);
  var optionred = {
    xAxis: {
      type: "category",
      data: titlered,
    },
    yAxis: {
      type: "value",
    },
    series: [
      {
        data: state.redbar,
        type: "bar",
      },
    ],
  };
  myChartred.setOption(optionred);

  var myChartblue = echarts.init(bluebar.value);
  var optionblue = {
    xAxis: {
      type: "category",
      data: titleblue,
    },
    yAxis: {
      type: "value",
    },
    series: [
      {
        data: state.bluebar,
        type: "bar",
      },
    ],
  };
  myChartblue.setOption(optionblue);
};
const redclick = (a) => {
  a.clicked = !a.clicked;
};
const random = () => {
  console.log(state.blue);
  state.red.forEach((item) => {
    item.clicked = false;
  });
  state.blue.forEach((item) => {
    item.clicked = false;
  });
  var numbers = [];
  for (var i = 0; i < 6; i++) {
    var randomNumber;

    do {
      randomNumber = Math.floor(Math.random() * state.red.length) + 1;
    } while (numbers.includes(randomNumber));

    numbers.push(randomNumber);
    state.red[randomNumber - 1].clicked = true;
  }
  var bluerandomNumber = Math.floor(Math.random() * state.blue.length) + 1;
  state.blue[bluerandomNumber - 1].clicked = true;
};
const setAll = (startNum) => {
  state.rednum = 0;
  state.bluenum = 0;
  let list = [];
  let redlist = [];
  let bluelist = [];
  for (var i = 0; i < state.options.length; i++) {
    list.push(setUp(state.options[i], startNum));
  }
  list.forEach((item, index) => {
    let redabs = item.red[5].abs;
    let blueabs = item.blue[0].abs;
    item.red.forEach((reditem, redindex) => {
      if (redindex < 6) {
        redlist.push(reditem.value);
      }
      if (redindex >= 6 && reditem.abs === redabs) {
        redlist.push(reditem.value);
      }
    });
    item.blue.forEach((blueitem, blueindex) => {
      if (blueindex == 0) {
        bluelist.push(blueitem.value);
      }
      if (blueindex > 0 && blueitem.abs == blueabs) {
        bluelist.push(blueitem.value);
      }
    });
  });
  let ared = countAndSortDuplicates(redlist);
  let bblue = countAndSortDuplicates(bluelist);
  let redbackground = ared[0].abs;
  let bluebackground = bblue[0].abs;
  let redcolor = 0;
  let bluecolor = 0;
  ared.forEach((item, index) => {
    if (state.redyes.includes(item.value)) {
      state.rednum += 1;
      item.clicked = true;
    }
    if (item.abs === redbackground) {
      item.background = redcolor + 1;
    } else {
      redbackground = item.abs;
      redcolor += 1;
      item.background = redcolor + 1;
    }
  });
  bblue.forEach((item, index) => {
    if (state.blueyes == item.value) {
      state.bluenum += 1;
      item.clicked = true;
    }
    if (item.abs === bluebackground) {
      item.background = bluecolor + 1;
    } else {
      bluebackground = item.abs;
      bluecolor += 1;
      item.background = bluecolor + 1;
    }
  });
  state.red = ared;
  state.blue = bblue;
  console.log(ared, bblue);
};
const countAndSortDuplicates = (arr) => {
  // 使用一个对象来存储数字出现的次数
  var countMap = {};

  // 遍历数组，统计每个数字的出现次数
  arr.forEach(function (num) {
    countMap[num] = (countMap[num] || 0) + 1;
  });

  // 将结果转换为数组，其中每个元素是包含数字和出现次数的对象
  var countArray = Object.keys(countMap).map(function (key) {
    return {
      value: parseInt(key),
      abs: countMap[key],
      clicked: false,
      background: 0,
    };
  });

  // 根据出现次数进行排序
  countArray.sort(function (a, b) {
    return b.abs - a.abs;
  });

  return countArray;
};
const reload = () => {
  location.reload();
};
</script>

<style lang="less" scoped>
.header {
  // width:30%;
  display: flex;
  flex-wrap:wrap;
  // flex-direction:column;
  justify-content:center;
}
.head {
  display: flex;
  width: 100%;
  div {
    border-right: 1px solid #000000;
    border-bottom: 1px solid #000000;
    width: 2%;
    text-align: center;
  }
}
.body {
  color: #7d7c7c;
  width: 100%;
  .row {
    display: flex;
    width: 100%;
    .cow {
      width: 2%;
      text-align: center;
      border-right: 1px solid #000000;
    }
  }
}
.border {
  border: 1px solid #000000;
}
.dataTime {
  height: 32px;
  line-height: 32px;
  margin-left: 20px;
}
.right {
  border-right: 2px solid #000000 !important;
}
.red {
  width:calc(50% - 10px);
  border: 1px solid #000000;
  // background-color: rgb(247, 145, 145);
  // color: #000000;
}
.blue {
  width:calc(50% - 10px);
  border: 1px solid #000000;
  // background-color: rgb(82, 125, 244);
  // color: #000000;
}
.footer {
  width: 100%;
  border-top: 1px solid #000000;
  .footers {
    display: flex;
    width: 100%;
    justify-content: space-between;
    div {
      border-right: 1px solid #000000;
      width: 2%;
      text-align: center;
    }
  }
}
.allred {
  // width: 20vw;
  border-bottom: 2px solid #e62727;
}
.redbacka{
  background-color: rgba(0,0,0,0.3);
}

.allblue {
  // width: 20vw;
  border-bottom: 2px solid #273ae6;
}
.bluebacka{
  background-color: rgba(0,0,0,0.3);
}
.isred {
  // width: 20vw;
  background-color: rgb(247, 145, 145);
}
.isblue {
  // width: 20vw;
  background-color: rgb(82, 125, 244);
}
.nonum {
  display: inline-block;
  width: 20px;
  text-align: center;
  border-right: 1px solid #000000;
  // background-color: #b5b4b4;
}
.nums {
  display: flex;
  cursor: pointer;
   flex-wrap:wrap;
}
.bars {
  width: 100%;
  height: 400px;
}
.el-card__body{
padding:5px!import;
}
.myLabel{
  height:32px;
  line-height:32px
}
.redspan{
  display: inline-block;
  width:calc(100% - 36px);
  padding-left: 15px;
}
.bluespan{
  display: inline-block;
  width:calc(100% - 36px);
  padding-left: 15px;
}
</style>
