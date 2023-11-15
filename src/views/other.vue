<template>
  <el-card class="category-container">
    <template #header>
      <div class="header">
        <el-upload class="upload-demo" :http-request="readxls">
          <el-button type="primary">Click to upload</el-button>
        </el-upload>
        <el-select
          style="margin-left: 20px"
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
        <el-input-number v-model="state.startNum" />
        <el-button
          @click="setUp(state.option, state.startNum)"
          style="margin-left: 20px"
          type="primary"
          >计算</el-button
        >
        <el-button @click="barList" style="margin-left: 20px" type="primary"
          >柱状图</el-button
        >
        <el-button @click="random" style="margin-left: 20px" type="primary"
          >随机数</el-button
        >
        <!-- @click="setUp(state.option, state.startNum)" -->
        <div class="dataTime">{{ state.time }}</div>
        <div class="dataTime">{{ state.allred }}</div>
      </div>
    </template>
    <div class="border" v-if="false">
      <div class="head">
        <div
          v-for="item in 33"
          :class="{ right: item == 11 || item == 22 || item == 33 }"
        >
          {{ item }}
        </div>
        <div v-for="item in 16">{{ item }}</div>
      </div>
      <div class="body">
        <div class="row" v-for="(item, index) in state.showList">
          <div
            class="cow"
            v-for="(items, indexs) in item"
            :class="{
              right:
                (indexs == 10 && items.class != 'blue') ||
                indexs == 21 ||
                indexs == 32,
              red: items.select && items.class == 'red',
              blue: items.select && items.class == 'blue',
            }"
          >
            {{ items.value }}
          </div>
        </div>
      </div>
      <div class="footer">
        <div class="footers">
          <div
            v-for="(item, index) in state.redFooter"
            :class="{ right: index == 10 || index == 21 || index == 32 }"
          >
            {{ item }}
          </div>
          <div v-for="(item, index) in state.blueFooter">
            {{ item }}
          </div>
        </div>
      </div>
    </div>
    <div class="nums">
      <div class="red">
        <div
          v-for="(item, index) in state.red"
          class="allred"
          :class="{ isred: state.redyes.includes(item.value) || item.clicked }"
          @click="redclick(item)"
        >
          <span class="nonum">{{ index + 1 }}</span
          >&nbsp;&nbsp;&nbsp;&nbsp; <span>{{ item.value }}</span
          >&nbsp;&nbsp;&nbsp;&nbsp;
          <span>{{ item.abs }}</span>
        </div>
      </div>
      <div class="blue">
        <div
          v-for="(item, index) in state.blue"
          class="allblue"
          :class="{ isblue: item.value == state.blueyes || item.clicked }"
          @click="redclick(item)"
        >
          <span class="nonum">{{ index + 1 }}</span
          >&nbsp;&nbsp;&nbsp;&nbsp; <span>{{ item.value }}</span
          >&nbsp;&nbsp;&nbsp;&nbsp;
          <span>{{ item.abs }}</span>
        </div>
      </div>
    </div>
    <div ref="redbar" class="bars"></div>
    <div ref="bluebar" class="bars"></div>
  </el-card>
</template>

<script setup>
import { reactive, ref, onMounted } from "vue";
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
});
const redbar = ref(null);
const bluebar = ref(null);
onMounted(() => {
  if (localStorage.getItem("data")) {
    state.xlsxList = JSON.parse(localStorage.getItem("data"));
  }
});
const readxls = async (e) => {
  const ab = await e.file.arrayBuffer();
  /* parse */
  const wb = read(ab);

  /* generate array of objects from first worksheet */
  const ws = wb.Sheets[wb.SheetNames[0]]; // get the first worksheet
  const data = utils.sheet_to_json(ws); // generate objects
  state.xlsxList = data.reverse();
  localStorage.setItem("data", JSON.stringify(state.xlsxList));
  /* update state */
  //   pres.value = data;
};
const setUp = (option, num) => {
  if (num > 0) {
    const data = state.xlsxList[num - 1];
    state.redyes = [
      data["__EMPTY_2"],
      data["__EMPTY_3"],
      data["__EMPTY_4"],
      data["__EMPTY_5"],
      data["__EMPTY_6"],
      data["__EMPTY_7"],
    ];
    state.blueyes = data["__EMPTY_8"];
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
  state.time =
    datalist[0]["__EMPTY_1"] +
    "------" +
    datalist[datalist.length - 1]["__EMPTY_1"];
  datalist.forEach((item, index) => {
    const red = [
      item["__EMPTY_2"],
      item["__EMPTY_3"],
      item["__EMPTY_4"],
      item["__EMPTY_5"],
      item["__EMPTY_6"],
      item["__EMPTY_7"],
    ];
    const blue = [item["__EMPTY_8"]];
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
  console.log(state.red,state.blue)
  var numbers = [];
  for (var i = 0; i < 6; i++) {
    var randomNumber;
    
    do {
      randomNumber = Math.floor(Math.random() * 33) + 1;
    } while (numbers.includes(randomNumber));

    numbers.push(randomNumber);
    state.red[randomNumber-1].clicked = true
  }
  var bluerandomNumber = Math.floor(Math.random() * 16) + 1;
  state.blue[bluerandomNumber-1].clicked = true
};
</script>

<style lang="less" scoped>
.header {
  display: flex;
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
  // background-color: rgb(247, 145, 145);
  // color: #000000;
}
.blue {
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
  width: 20vw;
  border-bottom: 2px solid #e62727;
}
.allblue {
  width: 20vw;
  border-bottom: 2px solid #273ae6;
}
.isred {
  width: 20vw;
  background-color: rgb(247, 145, 145);
}
.isblue {
  width: 20vw;
  background-color: rgb(82, 125, 244);
}
.nonum {
  display: inline-block;
  width: 20px;
  text-align: center;
  background-color: #b5b4b4;
}
.nums {
  display: flex;
  cursor: pointer;
}
.bars {
  width: 100%;
  height: 400px;
}
</style>
