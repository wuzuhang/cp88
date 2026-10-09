<script setup>
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { ArrowDown, ArrowUpRight, BarChart3, Check, ChevronLeft, ChevronRight, CircleHelp, History, LoaderCircle, RefreshCw, RotateCcw, Shuffle, Sparkles, Grid2X2, WifiOff } from 'lucide-vue-next'
import { analyze3D, fetchRecords, request } from './lib/api.js'
import { frequencyLevel } from './lib/analysis.js'
import { build3DMatrix, digits3D, positions3D, random3D, trialDigits3D } from './lib/fucai3d.js'

const records=ref([]),total=ref(0),loading=ref(true),loadError=ref('')
const tab=ref('matrix'),sampleWindow=ref(100),offset=ref(0),sortOrder=ref('asc')
const selected=ref([null,null,null]),result=ref(null),analyzing=ref(false),analysisError=ref('')
const search=ref(''),historyLimit=ref(12),moreLoading=ref(false),notice=ref('')
let version=0
const matrix=computed(()=>build3DMatrix(records.value,offset.value,sampleWindow.value))
const columns=computed(()=>matrix.value.columns.map(column=>sortOrder.value==='asc'?column:[...column].sort((a,b)=>b.count-a.count||a.number-b.number)))
const complete=computed(()=>selected.value.every(n=>n!==null))
const selectedCount=computed(()=>selected.value.filter(n=>n!==null).length)
const maxOffset=computed(()=>Math.max(0,records.value.length-sampleWindow.value))
const targetRecord=computed(()=>records.value[offset.value>0?offset.value-1:0])
const sampleLabel=computed(()=>{
  const sample=records.value.slice(offset.value,offset.value+sampleWindow.value)
  return sample.length?`${sample.at(-1).issue} — ${sample[0].issue}`:'等待历史数据'
})
const history=computed(()=>records.value.filter(r=>`${r.issue} ${r.draw_date}`.includes(search.value.trim())))

async function load(){
  loading.value=true;loadError.value=''
  try{const data=await fetchRecords('/fucai3d/records');records.value=data.items;total.value=data.total;offset.value=0;reset()}
  catch(error){loadError.value=error.message}finally{loading.value=false}
}
function choose(position,digit){selected.value[position]=selected.value[position]===digit?null:digit;notice.value=''}
function reset(){selected.value=[null,null,null];analysisError.value='';notice.value=''}
function randomize(){selected.value=random3D();notice.value='已生成一组随机号码'}
function setOffset(value){
  offset.value=Math.min(maxOffset.value,Math.max(0,value))
  if(offset.value===0){reset();return}
  selected.value=digits3D(records.value[offset.value-1]);notice.value='仅使用对照期之前的样本'
}
function useHistory(record){
  const index=records.value.findIndex(r=>r.issue===record.issue)
  if(index+1<=maxOffset.value){setOffset(index+1);tab.value='matrix'}
  else{selected.value=digits3D(record);tab.value='picker';notice.value='已载入号码；更早的历史样本不足，未切换对照窗口'}
}
async function analyze(){
  if(!complete.value||analyzing.value)return
  const currentVersion=version;analyzing.value=true;analysisError.value=''
  try{
    const data=await analyze3D([...selected.value])
    if(currentVersion===version){
      result.value=data
      if(window.innerWidth<=800){
        await nextTick()
        document.querySelector('.composition-results')?.scrollIntoView({behavior:window.matchMedia('(prefers-reduced-motion: reduce)').matches?'instant':'smooth',block:'center'})
      }
    }
  }
  catch(error){if(currentVersion===version)analysisError.value=error.message}
  finally{analyzing.value=false}
}
async function showMore(){
  historyLimit.value+=24
  if(historyLimit.value<=history.value.length||records.value.length>=total.value||moreLoading.value)return
  moreLoading.value=true
  try{
    const data=await request(`/fucai3d/records?limit=500&offset=${records.value.length}`)
    const seen=new Set(records.value.map(r=>r.issue))
    records.value.push(...data.items.filter(r=>!seen.has(r.issue)));total.value=data.total
  }catch(error){notice.value=error.message}finally{moreLoading.value=false}
}
watch(selected,()=>{version++;result.value=null;analysisError.value=''}, {deep:true,flush:'sync'})
watch(sampleWindow,()=>{if(offset.value>maxOffset.value)setOffset(maxOffset.value)})
watch(search,()=>historyLimit.value=12)
onMounted(load)
</script>

<template>
  <div class="app-shell fucai3d-page">
    <header class="site-header">
      <a class="brand" href="#/fucai3d"><span class="brand-mark"><i></i><i></i></span><span>球序<span class="brand-sub">3D LAB</span></span></a>
      <div class="header-caption">三个位置，更多观察。</div>
      <span class="header-status"><span :class="['status-dot',{offline:loadError}]"></span>{{loadError?'连接异常':loading?'同步中':'福彩 3D 分析'}}</span>
    </header>
    <main class="workspace">
      <div class="page-heading">
        <div><div class="eyebrow">OBSERVE. COMPARE. EXPLORE.</div><h1>福彩 3D<span>位置分析</span><span class="heading-dot">.</span></h1><p>从百、十、个位出发，观察 0–9 的历史频次与组合特征。</p></div>
        <button class="refresh-button" @click="load" :disabled="loading" aria-label="刷新福彩3D数据"><RefreshCw :size="15" :class="{spinning:loading}"/><span>刷新数据</span></button>
      </div>
      <div v-if="loadError" class="connection-error" role="alert"><WifiOff :size="22"/><div><strong>暂时无法读取福彩 3D 数据</strong><p>{{loadError}}</p></div><button class="button secondary" @click="load">重试</button></div>
      <div v-else-if="loading" class="loading-state"><LoaderCircle class="spinning" :size="24"/><span>正在读取开奖历史…</span></div>
      <div v-else-if="!records.length" class="connection-error"><History :size="22"/><div><strong>还没有福彩 3D 开奖记录</strong><p>导入历史数据后，刷新即可开始分析。</p></div></div>
      <template v-else>
        <section class="overview-grid">
          <article class="latest-card card">
            <div class="card-kicker"><span><span class="live-dot"></span>{{offset?'历史对照':'最新开奖'}}</span><span class="date">{{targetRecord.draw_date}}</span></div>
            <div class="issue-line"><h2>第 {{targetRecord.issue}} 期</h2><span class="subtle-tag">{{offset?'仅用此前样本':'福彩 3D'}}</span></div>
            <div class="three-draw"><div v-for="(digit,p) in digits3D(targetRecord)" :key="p"><span class="draw-ball red">{{digit}}</span><small>{{positions3D[p]}}</small></div></div>
            <div class="draw-sequence"><span>试机号</span><span>{{trialDigits3D(targetRecord).join(' · ')}}</span><span>机型 {{targetRecord.machine}} · 球号 {{targetRecord.ball_set}}</span></div>
          </article>
          <article class="insight-card card"><div class="card-kicker"><span>样本中的高频数字</span><BarChart3 :size="17"/></div><div class="hot-balls"><div v-for="item in matrix.hot" :key="item.number"><span>{{item.number}}</span><small>{{item.count}} 次</small></div></div><p>近 {{matrix.count}} 期 · 三个位置合计出现次数</p></article>
          <article class="archive-card card"><div class="card-kicker"><span>历史档案</span><History :size="17"/></div><div class="archive-number">{{total.toLocaleString()}}<span>期</span></div><div class="archive-foot"><span>数据库已收录</span><button @click="tab='history'">查看记录 <ArrowUpRight :size="15"/></button></div></article>
        </section>
        <div class="analysis-layout">
          <section class="main-panel card">
            <nav class="tabs" aria-label="福彩3D分析视图"><button :class="{active:tab==='matrix'}" @click="tab='matrix'"><Grid2X2 :size="16"/>频次矩阵</button><button :class="{active:tab==='picker'}" @click="tab='picker'"><Sparkles :size="16"/>号码选择</button><button :class="{active:tab==='history'}" @click="tab='history'"><History :size="16"/>开奖历史</button></nav>
            <div v-if="tab==='matrix'" class="panel-content matrix-content">
              <div class="section-heading"><div><h2>号码频次矩阵 <span class="mini-label">POSITION FREQUENCY</span></h2><p>百、十、个位分别计数，每列独立排序。</p></div><button class="sort-button" @click="sortOrder=sortOrder==='asc'?'desc':'asc'"><ArrowDown :size="14"/>{{sortOrder==='asc'?'低频优先':'高频优先'}}</button></div>
              <div class="sample-toolbar"><div class="sample-selects"><label>号码样本 <select v-model.number="sampleWindow"><option v-for="n in 10" :key="n" :value="n*100">近 {{n*100}} 期</option></select></label></div><div class="period-controls"><button @click="setOffset(offset+1)" :disabled="offset>=maxOffset" aria-label="向前对照一期"><ChevronLeft :size="17"/></button><span>{{offset?`向前 ${offset} 期`:'最新样本'}}</span><button @click="setOffset(offset-1)" :disabled="offset===0" aria-label="返回较新一期"><ChevronRight :size="17"/></button></div></div>
              <div class="matrix-meta"><span>样本期号 <b>{{sampleLabel}}</b></span><span>实际 {{matrix.count}} 期</span></div>
              <div class="matrix-wrap"><table class="frequency-table three-frequency"><caption class="sr-only">0 到 9 总频次与百、十、个位独立频次排名</caption><thead><tr><th class="total-heading">数字<small>三位合计</small></th><th v-for="position in positions3D" :key="position">{{position}}<small>{{matrix.count}} 期</small></th></tr></thead><tbody><tr v-for="(_,i) in 10" :key="i"><td class="total-cell"><span class="three-total"><b>{{i}}</b><small>{{matrix.totals[i]}}</small></span></td><td v-for="(column,p) in columns" :key="p" :class="['heat-cell',`heat-${frequencyLevel(column[i].count,matrix.count,10)}`]"><button :class="{selected:selected[p]===column[i].number}" @click="choose(p,column[i].number)" :aria-pressed="selected[p]===column[i].number" :aria-label="`${positions3D[p]}数字${column[i].number}，出现${column[i].count}次`"><b>{{column[i].number}}</b><span>{{column[i].count}}</span></button></td></tr></tbody></table></div>
              <div class="matrix-bottom"><div class="heat-legend"><span>相对频次</span><i v-for="n in 5" :key="n" :class="`heat-${n-1}`"></i><span>低 → 高</span></div><span class="three-sample-note">平均频次 = 实际期数 ÷ 10</span></div>
              <div class="matrix-tip"><CircleHelp :size="14"/><span>点击格子选择该位置的数字。允许重复，0 为有效号码；颜色随实际样本期数调整。同一行不代表一组号码。</span></div>
            </div>
            <div v-if="tab==='picker'" class="panel-content picker-content">
              <div class="section-heading"><div><h2>构建你的号码组合</h2><p>每位选择一个 0–9 的数字，允许重复。</p></div><span class="selection-count">{{selectedCount}} / 3</span></div>
              <div v-for="(position,p) in positions3D" :key="position" class="three-picker-group"><div class="picker-title"><span class="color-dot red-dot"></span><h3>{{position}}</h3><span>0 — 9</span></div><div class="number-grid red-grid three-number-grid"><button v-for="(_,digit) in 10" :key="digit" :class="{selected:selected[p]===digit}" @click="choose(p,digit)" :aria-pressed="selected[p]===digit" :aria-label="`${position}选择${digit}`"><b>{{digit}}</b><small>{{matrix.columns[p].find(item=>item.number===digit)?.count||0}} 次</small><Check v-if="selected[p]===digit" :size="11"/></button></div></div>
              <div class="picker-note"><CircleHelp :size="15"/>号码按百、十、个位保留顺序，不做排序。</div>
            </div>
            <div v-if="tab==='history'" class="panel-content history-content">
              <div class="section-heading"><div><h2>福彩 3D 开奖历史</h2><p>共 {{total.toLocaleString()}} 期，当前已加载 {{records.length}} 期。</p></div><History :size="20"/></div>
              <input class="history-search" v-model="search" placeholder="筛选已加载的期号或日期，如 2026267" aria-label="筛选福彩3D开奖历史"/>
              <div class="history-list"><button v-for="record in history.slice(0,historyLimit)" :key="record.issue" class="history-row" @click="useHistory(record)"><div class="history-period"><strong>{{record.issue}} <small>期</small></strong><span>{{record.draw_date}}</span></div><div class="history-balls"><span v-for="(digit,p) in digits3D(record)" :key="p" class="small-ball red">{{digit}}</span></div><span class="three-trial">试机 {{trialDigits3D(record).join('')}}</span><ChevronRight :size="16"/></button></div>
              <p v-if="!history.length" class="empty-search">已加载的数据中没有匹配记录。</p><button v-if="historyLimit<history.length||records.length<total" class="load-more" @click="showMore" :disabled="moreLoading">{{moreLoading?'加载中…':'加载更多历史'}}</button>
            </div>
          </section>
          <aside class="selection-panel card">
            <div class="selection-heading"><div><span class="eyebrow">YOUR COMBINATION</span><h2>我的 3D 号码</h2></div><span class="selection-icon"><Sparkles :size="19"/></span></div><p class="selection-description">按位置选择，观察组合特征。</p>
            <div class="three-selected"><div v-for="(digit,p) in selected" :key="p"><small>{{positions3D[p]}}</small><button class="draw-ball" :class="digit!==null?'red':'empty'" @click="selected[p]=null" :aria-label="`清空${positions3D[p]}号码`">{{digit!==null?digit:'—'}}</button></div></div>
            <div class="selected-status"><span>{{selectedCount}} / 3 位已选</span><span v-if="complete"><Check :size="12"/>已选齐</span></div>
            <div class="selection-actions"><button class="button secondary" @click="randomize"><Shuffle :size="15"/>随机一组</button><button class="reset-button" @click="reset" aria-label="清空已选号码"><RotateCcw :size="16"/></button></div>
            <button class="button analyze-button" @click="analyze" :disabled="!complete||analyzing"><LoaderCircle v-if="analyzing" :size="17" class="spinning"/><BarChart3 v-else :size="17"/>{{analyzing?'分析中…':'分析号码组合'}}<ArrowUpRight :size="17"/></button>
            <p v-if="notice" class="three-notice" role="status">{{notice}}</p><p v-if="analysisError" class="analysis-error" role="alert">{{analysisError}} <button @click="analyze">重试</button></p>
            <div v-if="result" class="composition-results"><div class="result-title"><span>组合特征</span><span>{{result.pattern}} <Check :size="12"/></span></div><div class="metrics-grid"><div><span>和值</span><b>{{result.sum}}</b></div><div><span>跨度</span><b>{{result.span}}</b></div><div><span>奇偶比例</span><b>{{result.odd_count}} : {{result.even_count}}</b></div><div><span>大小比例</span><b>{{result.big_count}} : {{result.small_count}}</b></div></div><p>小数 0–4 · 大数 5–9<br/>三同为豹子、两同为组三、各不相同为组六。</p></div>
            <div v-else class="result-placeholder"><span class="placeholder-chart"><BarChart3 :size="23"/></span><span>选齐三位数字后，查看和值、跨度<br/>奇偶、大小比例与号码形态。</span></div>
            <div class="selection-footnote"><CircleHelp :size="15"/><p>统计用于观察历史分布，不代表未来开奖概率。</p></div>
          </aside>
        </div>
      </template>
      <footer class="page-footer"><span class="footer-brand">球序 <span>3D LAB</span></span><span>观察历史 · 理性选择</span><span>数据来自历史开奖记录</span></footer>
    </main>
    <div v-if="records.length&&!loading&&!loadError" class="mobile-dock"><button class="dock-selection" @click="tab='picker'"><span>我的号码 <ChevronRight :size="12"/></span><b>{{selected.map(n=>n===null?'—':n).join(' · ')}} <small>{{selectedCount}} / 3 位</small></b></button><button class="dock-random" @click="randomize" aria-label="随机生成3D号码"><Shuffle :size="18"/></button><button class="button dock-analyze" @click="analyze" :disabled="!complete||analyzing"><BarChart3 :size="14"/>{{analyzing?'分析中':'分析组合'}}</button></div>
  </div>
</template>

<style>
.three-draw{display:flex;gap:15px;margin:14px 0}.three-draw>div,.three-selected>div{display:flex;align-items:center;flex-direction:column;gap:6px}.three-draw small,.three-selected small{font-size:10px;color:#9aa390}.three-draw .draw-ball{width:40px;height:40px;font-size:18px}.three-selected{display:flex;gap:22px;margin:22px 0}.three-selected .draw-ball{width:46px;height:46px;font-size:18px}.three-frequency{min-width:0;width:100%;table-layout:fixed}.three-total{display:flex;align-items:center;justify-content:center;gap:12px;height:44px;font-size:13px}.three-total small{font-size:10px;color:#8d9a81}.three-frequency th:first-child{width:25%}.fucai3d-page .three-number-grid{grid-template-columns:repeat(10,minmax(0,1fr))}.three-picker-group{margin:20px 0 26px}.three-sample-note,.three-trial,.three-notice{font-size:10px;color:#869778}.three-notice{line-height:1.7;margin-top:12px}.fucai3d-page .draw-sequence{flex-wrap:wrap}.fucai3d-page .matrix-wrap{overflow:auto}
@media(max-width:800px){.fucai3d-page .three-number-grid{grid-template-columns:repeat(5,minmax(0,1fr))}.three-trial{display:none}.fucai3d-page .matrix-meta{gap:8px;flex-wrap:wrap}.three-sample-note{font-size:9px}.fucai3d-page .matrix-bottom{gap:8px;flex-wrap:wrap}}
</style>
