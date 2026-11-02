<script setup>
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { ArrowUpRight, LoaderCircle, Sparkles } from 'lucide-vue-next'
import { ball } from '../lib/analysis.js'
import { comparePrediction, generateReferences } from '../lib/prediction.js'

const props=defineProps({game:{type:String,required:true},sample:{type:Array,required:true},targetLabel:{type:String,required:true},winningRecord:{type:Object,default:null}})
defineEmits(['use'])
const output=ref(null),generating=ref(false),error=ref('')
const gameName=computed(()=>props.game==='ssq'?'双色球':'福彩 3D')
const comparedRows=computed(()=>output.value?.rows.map(row=>({...row,comparison:comparePrediction(props.game,row,props.winningRecord)})) || [])
let version=0
function invalidate(){version++;output.value=null;generating.value=false;error.value=''}
watch([()=>props.sample,()=>props.game,()=>props.targetLabel],invalidate,{flush:'sync'})
onBeforeUnmount(()=>version++)
async function generate(){
  if(generating.value||!props.sample.length)return
  const current=++version,sample=props.sample,game=props.game
  generating.value=true;error.value=''
  await nextTick()
  await new Promise(resolve=>setTimeout(resolve,0))
  if(current!==version)return
  try{const data=generateReferences(game,sample);if(current===version)output.value=data}
  catch(err){if(current===version){output.value=null;error.value=err.message}}
  finally{if(current===version)generating.value=false}
}
</script>

<template>
  <section class="reference-panel card" :aria-label="`${gameName}预测参考`">
    <div class="reference-heading"><div><span class="eyebrow">STATISTICAL REFERENCE</span><h2>{{gameName}}预测参考</h2><p>{{targetLabel}} · 实际 {{sample.length}} 期样本 · 贝叶斯平滑与{{game==='ssq'?'独立抽样':'加权抽样'}}</p></div><button class="button reference-generate" @click="generate" :disabled="generating||!sample.length"><LoaderCircle v-if="generating" :size="16" class="spinning"/><Sparkles v-else :size="16"/>{{generating?'分析样本中…':`预测10注${game==='ssq'?'双色球':'3D'}`}}</button></div>
    <p v-if="error" class="analysis-error" role="alert">{{error}}</p>
    <template v-if="output">
      <div class="reference-summary"><div><small>样本期号</small><b>{{output.firstIssue}} — {{output.lastIssue}}</b></div><div><small>{{game==='ssq'?'红球':'三位'}}和值均值 ± 标准差</small><b>{{output.meanSum.toFixed(1)}} ± {{output.sdSum.toFixed(1)}}</b></div><div><small>平均跨度</small><b>{{output.meanSpan.toFixed(1)}}</b></div><div><small title="归一化信息熵取值0到1，越接近1表示各位置分布越均匀；不是预测准确率">位置平均信息熵</small><b>{{output.entropy.toFixed(3)}} <small>/ 1</small></b></div></div>
      <p v-if="game==='ssq'" class="reference-disclaimer">10注按同一个样本模型独立抽取，注间允许重复；每注内部六个红球不重复。</p>
      <p v-if="winningRecord" class="reference-match-legend">对照第 {{winningRecord.issue}} 期 · ✓ 标记命中号码{{game==='fucai3d'?'（按百、十、个位对应比较）':'（红球不分出球顺序，蓝球单独比较）'}}</p>
      <div class="reference-list">
        <article v-for="(row,index) in comparedRows" :key="`${index}:${row.key}`" class="reference-row">
          <span class="reference-index">{{String(index+1).padStart(2,'0')}}</span>
          <div class="reference-combination">
            <div class="reference-balls">
              <template v-if="game==='ssq'">
                <span v-for="(n,p) in row.reds" :key="n" class="small-ball red" :class="{'matched':row.comparison?.matches[p],'not-matched':row.comparison&&!row.comparison.matches[p]}" :aria-label="`红球${ball(n)}${row.comparison?(row.comparison.matches[p]?'，命中':'，未命中'):''}`" :title="row.comparison?(row.comparison.matches[p]?'红球命中':'红球未命中'):undefined">{{ball(n)}}</span>
                <span class="small-ball blue" :class="{'matched':row.comparison?.blueMatch,'not-matched':row.comparison&&!row.comparison.blueMatch}" :aria-label="`蓝球${ball(row.blue)}${row.comparison?(row.comparison.blueMatch?'，命中':'，未命中'):''}`" :title="row.comparison?(row.comparison.blueMatch?'蓝球命中':'蓝球未命中'):undefined">{{ball(row.blue)}}</span>
              </template>
              <template v-else>
                <span v-for="(n,p) in row.numbers" :key="p" class="small-ball red" :class="{'matched':row.comparison?.matches[p],'not-matched':row.comparison&&!row.comparison.matches[p]}" :aria-label="`${['百位','十位','个位'][p]}${n}${row.comparison?(row.comparison.matches[p]?'，同位命中':'，未命中'):''}`" :title="row.comparison?(row.comparison.matches[p]?'同位置命中':'该位置未命中'):undefined">{{n}}</span>
                <span class="reference-pattern">{{row.pattern}}</span>
              </template>
            </div>
            <p>和值 {{row.sum}} · 跨度 {{row.span}} · 奇偶 {{row.odd}}:{{(game==='ssq'?6:3)-row.odd}}</p>
            <p v-if="row.comparison" class="reference-hit-summary"><template v-if="game==='ssq'">红球命中 {{row.comparison.redCount}} / 6 · 蓝球{{row.comparison.blueMatch?'命中':'未中'}}</template><template v-else>同位命中 {{row.comparison.positionCount}} / 3</template></p>
          </div>
          <button class="reference-use" @click="$emit('use',row)" :aria-label="`载入参考第${index+1}注`">载入<ArrowUpRight :size="13"/></button>
        </article>
      </div>
      <details class="reference-method"><summary>查看分析方法与统计口径</summary><p>各位置采用对称 Dirichlet 先验，总伪计数为 50。平滑权重为（出现次数＋50÷号码数量）÷（实际样本期数＋50），降低短期波动。信息熵用于描述分布均匀程度；遗漏用于观察历史，不作为“该出了”的加分依据。</p><p v-if="game==='ssq'">每注都按同一个样本模型独立计算。动态规划汇总全部合法红球分配的条件权重，按条件分布抽出六个互不重复红球，再独立抽取蓝球；最终红球升序显示。各注不去重、不互相调整权重，允许重复号码和完全相同的组合。</p><p v-else>完整枚举 000–999，按百、十、个位的平滑权重乘积抽样。这是独立位置模型，允许 0、重复数字和豹子；参考号码按单选的百十个位顺序展示。</p><p v-if="game==='fucai3d'">3D保留分散规则：对相似组合施加分散惩罚，再不放回抽取10注。</p><p>序号仅作编号，不代表中奖率排名。切换样本、历史目标期或更新数据后，旧结果会清空。</p></details>
    </template>
    <p class="reference-disclaimer">历史模型不能改变独立随机开奖的理论概率。这10注仅供统计参考，不保证中奖或提高中奖率。</p>
  </section>
</template>

<style>
.reference-panel{padding:22px 24px;margin-bottom:24px}.reference-heading{display:flex;align-items:center;justify-content:space-between;gap:18px}.reference-heading h2{font-size:17px;margin:7px 0}.reference-heading p{font-size:11px;color:#89977e;line-height:1.8}.reference-generate{background:#304d3e;color:#fff;flex-shrink:0;min-height:40px}.reference-summary{display:grid;grid-template-columns:1.6fr 1.25fr 1fr 1fr;gap:16px;margin:20px 0;padding:16px;background:#f6f8f2;border-radius:8px}.reference-summary>div{display:flex;flex-direction:column;gap:7px}.reference-summary small{font-size:10px;color:#8d9a81}.reference-summary b{font-size:13px;font-weight:500;color:#526b43}.reference-list{display:grid;grid-template-columns:1fr 1fr;gap:8px 18px}.reference-row{display:flex;align-items:center;gap:12px;padding:13px 0;border-bottom:1px solid #edf0e8;min-width:0}.reference-index{font-size:11px;color:#a0ac96}.reference-combination{flex:1;min-width:0}.reference-balls{display:flex;align-items:center;gap:6px;flex-wrap:wrap}.reference-combination p{font-size:10px;color:#94a187;margin-top:7px;line-height:1.7}.reference-use{display:flex;align-items:center;gap:3px;font-size:10px;color:#70865f;white-space:nowrap}.reference-pattern{font-size:10px;color:#8c9c7d;margin-left:5px}.reference-method{margin-top:18px;color:#88967d;font-size:11px;line-height:1.9}.reference-method summary{cursor:pointer;color:#637d51}.reference-method p{margin-top:8px}.reference-disclaimer{font-size:10px;line-height:1.9;color:#97a28c;margin-top:15px}
@media(max-width:800px){.reference-panel{padding:20px}.reference-heading{align-items:flex-start;flex-wrap:wrap}.reference-generate{margin-left:auto}.reference-summary{grid-template-columns:1fr 1fr;gap:15px}.reference-list{grid-template-columns:1fr}}
@media(max-width:360px){.reference-row{gap:7px}.reference-balls{gap:4px}.reference-balls .small-ball{width:22px;height:22px;font-size:10px}.reference-use{font-size:9px}.reference-summary{padding:12px;gap:13px}.reference-summary b{font-size:11px}.reference-panel{padding:18px 14px}}
.reference-balls .small-ball.matched{position:relative;background:#ce5b57;color:#fff;border:1px solid #c65450;box-shadow:0 0 0 2px #f5ded9}.reference-balls .small-ball.blue.matched{background:#547fac;border-color:#4d76a2;box-shadow:0 0 0 2px #dfe9f5}.reference-balls .small-ball.matched::after{content:'✓';position:absolute;right:-3px;bottom:-3px;display:flex;align-items:center;justify-content:center;width:11px;height:11px;border-radius:50%;background:#304d3e;color:white;border:1px solid white;font-size:8px;line-height:1}.reference-balls .small-ball.not-matched{opacity:.48}.reference-match-legend{font-size:11px;color:#6c825c;line-height:1.9;margin-top:13px}.reference-combination .reference-hit-summary{color:#5c7c49;font-size:10px}
</style>
