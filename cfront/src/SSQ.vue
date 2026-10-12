<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ArrowDown, ArrowUpRight, BarChart3, Check, ChevronDown, ChevronLeft, ChevronRight, CircleHelp, Clock3, Grid2X2, History, LoaderCircle, RefreshCw, RotateCcw, Shuffle, Sparkles, WifiOff } from 'lucide-vue-next'
import { analyzeBalls, fetchRecords, request } from './lib/api.js'
import { ball, buildMatrix, formatOmission, frequencyLevel, positionFrequencyTotal, predictionTargetIndex, randomCombination, redBalls, sampleBeforeTarget, sequenceBalls } from './lib/analysis.js'
import WinningFrequency from './components/WinningFrequency.vue'

const records = ref([])
const total = ref(0)
const loading = ref(true)
const loadError = ref('')
const tab = ref('matrix')
const offset = ref(0)
const redWindow = ref(100)
const blueWindow = ref(100)
const sortOrder = ref('asc')
const expanded = ref(false)
const selected = ref(Array(6).fill(null))
const blue = ref(null)
const result = ref(null)
const analyzing = ref(false)
const analysisError = ref('')
const notice = ref('')
const search = ref('')
const historyLimit = ref(12)
const moreLoading = ref(false)
const matrix = computed(() => buildMatrix(records.value,predictionTargetIndex(offset.value),redWindow.value,blueWindow.value))
const columns = computed(() => matrix.value.columns.map(list => sortOrder.value === 'asc' ? list : [...list].sort((a,b) => b.count-a.count || a.number-b.number)))
const blueColumn = computed(() => sortOrder.value === 'asc' ? matrix.value.blue : [...matrix.value.blue].sort((a,b) => b.count-a.count || a.number-b.number))
const chosenReds = computed(() => selected.value.filter(n => n !== null).sort((a,b) => a-b))
const complete = computed(() => chosenReds.value.length === 6 && blue.value !== null)
const targetRecord = computed(() => records.value[predictionTargetIndex(offset.value)])
const winningNumbers = computed(() => targetRecord.value ? [...sequenceBalls(targetRecord.value),targetRecord.value.blue] : [])
const winningFrequency = computed(() => positionFrequencyTotal([...matrix.value.columns,matrix.value.blue],winningNumbers.value))
const winningOmissions = computed(() => winningNumbers.value.map((number,p)=>p<6?matrix.value.omissions[p][number-1]:matrix.value.blueOmissions[number-1]))
const maxOffset = computed(() => Math.max(0,records.value.length-Math.max(redWindow.value,blueWindow.value)))
const sampleLabel = computed(() => {
  const items=sampleBeforeTarget(records.value,predictionTargetIndex(offset.value),redWindow.value)
  return items.length ? `${items.at(-1).issue} — ${items[0].issue}` : '等待历史数据'
})
const filteredHistory = computed(() => records.value.filter(r => `${r.issue} ${r.draw_date}`.includes(search.value.trim())))
const historyItems = computed(() => filteredHistory.value.slice(0,historyLimit.value))
let noticeTimer
let analysisVersion=0

function notify(message) {
  notice.value=message
  clearTimeout(noticeTimer)
  noticeTimer=setTimeout(()=>notice.value='',3000)
}

async function load() {
  loading.value=true
  loadError.value=''
  try {
    const data=await fetchRecords()
    records.value=data.items
    total.value=data.total
    offset.value=0
    selected.value=Array(6).fill(null)
    blue.value=null
  } catch(error) { loadError.value=error.message }
  finally { loading.value=false }
}

function toggleRed(number) {
  const index=selected.value.indexOf(number)
  if(index>=0) { selected.value[index]=null; return }
  const empty=selected.value.indexOf(null)
  if(empty<0) { notify('已选满 6 个红球，先取消一个再选择'); return }
  selected.value[empty]=number
}

function selectPosition(position,number) {
  if(selected.value[position]===number) { selected.value[position]=null; return }
  if(selected.value.includes(number)) { notify(`${ball(number)} 已在其他出球位置中选中`); return }
  selected.value[position]=number
}

function randomize() {
  const combination=randomCombination()
  selected.value=combination.reds
  blue.value=combination.blue
  notify('已生成一组随机号码')
}

function reset() { selected.value=Array(6).fill(null); blue.value=null; analysisError.value='' }

function setOffset(value) {
  offset.value=Math.min(maxOffset.value,Math.max(0,value))
  const record=records.value[predictionTargetIndex(offset.value)]
  if(!record) { reset(); return }
  selected.value=sequenceBalls(record)
  blue.value=record.blue
}

function useHistory(record) {
  const index=records.value.findIndex(r=>r.issue===record.issue)
  if(index+1<=maxOffset.value) {
    setOffset(index+1)
    notify(`已对照 ${record.issue} 期，样本仅使用更早的开奖`)
    tab.value='matrix'
  } else {
    selected.value=sequenceBalls(record)
    blue.value=record.blue
    tab.value='picker'
    notify('已载入这期号码；历史样本不足，未切换对照窗口')
  }
}

async function analyze() {
  if(!complete.value || analyzing.value) return
  const version=analysisVersion
  analyzing.value=true
  analysisError.value=''
  try {
    const data=await analyzeBalls(chosenReds.value,blue.value)
    if(version===analysisVersion) {
      result.value=data
      if(window.innerWidth<=800) {
        await nextTick()
        document.querySelector('.composition-results')?.scrollIntoView({behavior:window.matchMedia('(prefers-reduced-motion: reduce)').matches?'instant':'smooth',block:'center'})
      }
    }
  } catch(error) { if(version===analysisVersion) analysisError.value=error.message }
  finally { analyzing.value=false }
}

async function showMoreHistory() {
  historyLimit.value+=24
  if(historyLimit.value <= filteredHistory.value.length || records.value.length>=total.value || moreLoading.value) return
  moreLoading.value=true
  try {
    const data=await request(`/records?limit=500&offset=${records.value.length}`)
    const issues=new Set(records.value.map(r=>r.issue))
    records.value.push(...data.items.filter(r=>!issues.has(r.issue)))
    total.value=data.total
  } catch(error) { notify(error.message) }
  finally { moreLoading.value=false }
}

watch([selected,blue],()=>{ analysisVersion++; result.value=null; analysisError.value='' },{deep:true,flush:'sync'})
watch([redWindow,blueWindow],()=>{ if(offset.value>maxOffset.value) setOffset(maxOffset.value) })
watch(search,()=>historyLimit.value=12)
onMounted(load)
onBeforeUnmount(()=>clearTimeout(noticeTimer))
</script>

<template>
  <div class="app-shell">
    <header class="site-header">
      <a class="brand" href="/" aria-label="球序首页"><span class="brand-mark"><i></i><i></i></span><span>球序<span class="brand-sub">SSQ LAB</span></span></a>
      <div class="header-caption">让每一个号码，有迹可循。</div>
      <span class="header-status"><span :class="['status-dot',{offline:loadError}]"></span>{{ loadError ? '连接异常' : loading ? '同步中' : '历史数据分析' }}</span>
    </header>

    <main class="workspace">
      <div class="page-heading">
        <div><div class="eyebrow">OBSERVE. COMPARE. EXPLORE.</div><h1>发现号码的<span>另一面</span><span class="heading-dot">.</span></h1><p>从历史出球位置出发，观察频次，探索你的号码组合。</p></div>
        <button class="refresh-button" @click="load" :disabled="loading"><RefreshCw :size="15" :class="{spinning:loading}" /><span>刷新数据</span></button>
      </div>

      <div v-if="loadError" class="connection-error" role="alert"><WifiOff :size="22"/><div><strong>暂时无法读取开奖数据</strong><p>{{ loadError }}</p></div><button class="button secondary" @click="load">重试</button></div>
      <div v-else-if="loading" class="loading-state"><LoaderCircle class="spinning" :size="24"/><span>正在读取开奖历史…</span></div>
      <div v-else-if="!records.length" class="connection-error"><History :size="22"/><div><strong>还没有开奖记录</strong><p>导入历史数据后，刷新即可开始分析。</p></div></div>

      <template v-else>
        <section class="overview-grid">
          <article class="latest-card card">
            <template v-if="targetRecord">
            <div class="card-kicker"><span><span class="live-dot"></span>{{ offset ? '历史对照' : '最新开奖' }}</span><span class="date">{{ targetRecord.draw_date }}</span></div>
            <div class="issue-line"><h2>第 {{ targetRecord.issue }} 期</h2><span class="subtle-tag">{{ offset ? '仅用此前样本' : '双色球' }}</span></div>
            <div class="draw-balls"><span v-for="n in redBalls(targetRecord)" :key="n" class="draw-ball red">{{ ball(n) }}</span><span class="ball-divider"></span><span class="draw-ball blue">{{ ball(targetRecord.blue) }}</span></div>
            <div class="draw-sequence"><span>出球顺序</span><span>{{ sequenceBalls(targetRecord).map(ball).join(' · ') }}</span></div>
            <WinningFrequency :frequency="winningFrequency" :omissions="winningOmissions" :numbers="winningNumbers.map(ball)" :labels="['红1','红2','红3','红4','红5','红6','蓝球']" :sample-label="`红球 ${matrix.redCount} 期 / 蓝球 ${matrix.blueCount} 期 · 仅用此前样本`"/>
            </template>
            <template v-else>
              <div class="card-kicker"><span><span class="live-dot"></span>下一期预测</span><span>双色球</span></div>
              <div class="issue-line"><h2>下一期尚未开奖</h2><span class="subtle-tag">等待开奖</span></div>
              <p class="prediction-pending">暂无中奖号码。最新已开奖 {{records[0].issue}} 期已计入频次和遗漏样本。</p>
              <div class="draw-sequence"><span>红球 {{matrix.redCount}} 期 · 蓝球 {{matrix.blueCount}} 期</span></div>
            </template>
          </article>
          <article class="insight-card card">
            <div class="card-kicker"><span>样本中的高频号码</span><BarChart3 :size="17" /></div>
            <div class="hot-balls"><div v-for="item in matrix.hot" :key="item.number"><span>{{ ball(item.number) }}</span><small>{{ item.count }} 次</small></div></div>
            <p>近 {{ matrix.redCount }} 期红球 · 出现次数前 5 位</p>
          </article>
          <article class="archive-card card"><div class="card-kicker"><span>历史档案</span><History :size="17"/></div><div class="archive-number">{{ total.toLocaleString() }}<span>期</span></div><div class="archive-foot"><span>数据库已收录</span><button @click="tab='history'">查看记录 <ArrowUpRight :size="15"/></button></div></article>
        </section>

        <div class="analysis-layout">
          <section class="main-panel card">
            <nav class="tabs" aria-label="分析视图"><button :class="{active:tab==='matrix'}" @click="tab='matrix'"><Grid2X2 :size="16"/>频次矩阵</button><button :class="{active:tab==='picker'}" @click="tab='picker'"><Sparkles :size="16"/>号码选择</button><button :class="{active:tab==='history'}" @click="tab='history'"><History :size="16"/>开奖历史</button></nav>

            <div v-if="tab==='matrix'" class="panel-content matrix-content">
              <div class="section-heading"><div><h2>号码频次矩阵 <span class="mini-label">POSITION FREQUENCY</span></h2><p>按实际出球顺序分列，每列独立排序。</p></div><button class="sort-button" @click="sortOrder=sortOrder==='asc'?'desc':'asc'"><ArrowDown :size="14" :class="{flipped:sortOrder==='desc'}"/>{{ sortOrder==='asc'?'低频优先':'高频优先' }}</button></div>
              <div class="sample-toolbar"><div class="sample-selects"><label>红球样本 <select v-model.number="redWindow"><option v-for="n in 10" :key="n" :value="n*100">近 {{ n*100 }} 期</option></select></label><span class="toolbar-divider"></span><label>蓝球样本 <select v-model.number="blueWindow"><option v-for="n in 10" :key="n" :value="n*100">近 {{ n*100 }} 期</option></select></label></div><div class="period-controls"><button @click="setOffset(offset+1)" :disabled="offset>=maxOffset" aria-label="向前对照一期"><ChevronLeft :size="17"/></button><span>{{ offset ? `向前 ${offset} 期` : '下一期预测' }}</span><button @click="setOffset(offset-1)" :disabled="offset===0" aria-label="返回较新一期"><ChevronRight :size="17"/></button></div></div>
              <div class="matrix-meta"><span>样本期号 <b>{{ sampleLabel }}</b></span><span><i class="tiny-ring"></i> 点击格子选择号码</span></div>
              <div class="matrix-wrap"><table class="frequency-table"><caption class="sr-only">红球总频次、六个出球位置独立排名和蓝球独立排名</caption><thead><tr><th class="total-heading">号码<small>总频次</small></th><th v-for="n in 6" :key="n">红 {{ n }}<small>第 {{ n }} 位</small></th><th class="blue-heading">蓝球<small>{{ matrix.blueCount }} 期</small></th></tr></thead><tbody><tr v-for="(_,i) in (expanded?33:16)" :key="i"><td class="total-cell"><button @click="toggleRed(i+1)" :class="{chosen:chosenReds.includes(i+1)}" :aria-label="`红球${i+1}，总出现${matrix.totals[i]}次`"><b>{{ ball(i+1) }}</b><span>{{ matrix.totals[i] }}</span></button></td><td v-for="(column,position) in columns" :key="position" :class="['heat-cell',`heat-${frequencyLevel(column[i].count,matrix.redCount,33)}`]"><button :class="{selected:selected[position]===column[i].number}" @click="selectPosition(position,column[i].number)" :aria-pressed="selected[position]===column[i].number" :aria-label="`出球位置${position+1}，红球${column[i].number}，出现${column[i].count}次`"><b>{{ ball(column[i].number) }}</b><span>{{ column[i].count }}</span><small class="matrix-omission">漏{{formatOmission(matrix.omissions[position][column[i].number-1])}}</small></button></td><td class="blue-cell" :class="i<16?`blue-heat-${frequencyLevel(blueColumn[i].count,matrix.blueCount,16)}`:'blank-cell'"><button v-if="i<16" :class="{selected:blue===blueColumn[i].number}" @click="blue=blue===blueColumn[i].number?null:blueColumn[i].number" :aria-pressed="blue===blueColumn[i].number" :aria-label="`蓝球${blueColumn[i].number}，出现${blueColumn[i].count}次`"><b>{{ ball(blueColumn[i].number) }}</b><span>{{ blueColumn[i].count }}</span><small class="matrix-omission">漏{{formatOmission(matrix.blueOmissions[blueColumn[i].number-1])}}</small></button><span v-else>—</span></td></tr></tbody></table></div>
              <div class="matrix-bottom"><div class="heat-legend"><span title="按实际样本期数归一化：平均频次 = 期数 ÷ 号码数量">相对频次</span><i v-for="n in 5" :key="n" :class="`heat-${n-1}`"></i><span>低 → 高</span></div><button @click="expanded=!expanded">{{ expanded?'收起矩阵':'展开全部 33 行' }}<ChevronDown :size="15" :class="{flipped:expanded}"/></button></div>
              <div class="matrix-tip"><CircleHelp :size="14"/><span>小数字为出现次数，颜色按实际样本期数动态调整。红球各位置按 33 个号码、蓝球按 16 个号码分别计算平均频次。各列独立排序，同一行不代表一组号码。</span></div>
            </div>

            <div v-if="tab==='picker'" class="panel-content picker-content"><div class="section-heading"><div><h2>构建你的号码组合</h2><p>选择 6 个不同红球和 1 个蓝球，再查看组合特征。</p></div><span class="selection-count">{{ chosenReds.length }} / 6</span></div><div class="picker-title"><span class="color-dot red-dot"></span><h3>红球</h3><span>01 — 33</span></div><div class="number-grid red-grid"><button v-for="n in 33" :key="n" :class="{selected:chosenReds.includes(n)}" @click="toggleRed(n)" :aria-pressed="chosenReds.includes(n)"><b>{{ ball(n) }}</b><small>{{ matrix.totals[n-1] }} 次</small><Check v-if="chosenReds.includes(n)" :size="11"/></button></div><div class="picker-title"><span class="color-dot blue-dot"></span><h3>蓝球</h3><span>01 — 16</span></div><div class="number-grid blue-grid"><button v-for="n in 16" :key="n" :class="{selected:blue===n}" @click="blue=blue===n?null:n" :aria-pressed="blue===n"><b>{{ ball(n) }}</b><small>{{ matrix.blue.find(item=>item.number===n)?.count || 0 }} 次</small><Check v-if="blue===n" :size="11"/></button></div><div class="picker-note"><CircleHelp :size="15"/>小数字为当前样本中的出现次数。红球自动按选择先后填入出球位置。</div></div>

            <div v-if="tab==='history'" class="panel-content history-content"><div class="section-heading"><div><h2>开奖历史</h2><p>数据库共 {{ total.toLocaleString() }} 期，当前已加载 {{ records.length }} 期。</p></div><Clock3 :size="20" class="muted-icon"/></div><input class="history-search" v-model="search" placeholder="筛选已加载的期号或日期，如 2026115" aria-label="筛选开奖历史"/><div class="history-list"><button v-for="record in historyItems" :key="record.issue" class="history-row" @click="useHistory(record)"><div class="history-period"><strong>{{ record.issue }} <small>期</small></strong><span>{{ record.draw_date }}</span></div><div class="history-balls"><span v-for="n in redBalls(record)" :key="n" class="small-ball red">{{ ball(n) }}</span><span class="small-ball blue">{{ ball(record.blue) }}</span></div><ChevronRight :size="16"/></button></div><p v-if="!historyItems.length" class="empty-search">已加载的数据中没有匹配记录。</p><button v-if="historyLimit<filteredHistory.length || records.length<total" class="load-more" @click="showMoreHistory" :disabled="moreLoading"><LoaderCircle v-if="moreLoading" :size="14" class="spinning"/>{{ moreLoading?'加载中…':'加载更多历史' }}<ChevronDown v-if="!moreLoading" :size="14"/></button></div>
          </section>

          <aside class="selection-panel card"><div class="selection-heading"><div><span class="eyebrow">YOUR COMBINATION</span><h2>我的号码</h2></div><span class="selection-icon"><Sparkles :size="19"/></span></div><p class="selection-description">每次选择，都是一次新的观察。</p><div class="selected-numbers"><span v-for="(n,i) in chosenReds" :key="n" class="draw-ball red" @click="toggleRed(n)">{{ ball(n) }}</span><span v-for="n in 6-chosenReds.length" :key="`empty${n}`" class="draw-ball empty">—</span><span class="draw-ball" :class="blue?'blue':'empty blue-empty'" @click="blue=null">{{ blue?ball(blue):'—' }}</span></div><div class="selected-status"><span>{{ chosenReds.length }} 个红球 · {{ blue?1:0 }} 个蓝球</span><span v-if="complete"><Check :size="12"/>已选齐</span></div><div class="selection-actions"><button class="button secondary" @click="randomize"><Shuffle :size="15"/>随机一组</button><button class="reset-button" @click="reset" aria-label="清空已选号码"><RotateCcw :size="16"/></button></div><button class="button analyze-button" @click="analyze" :disabled="!complete || analyzing"><LoaderCircle v-if="analyzing" :size="17" class="spinning"/><BarChart3 v-else :size="17"/>{{ analyzing?'分析中…':'分析号码组合' }}<ArrowUpRight v-if="!analyzing" :size="17"/></button><p v-if="analysisError" class="analysis-error" role="alert">{{ analysisError }} <button @click="analyze">重试</button></p>
            <div class="composition-results" v-if="result"><div class="result-title"><span>组合特征</span><span>已完成分析 <Check :size="12"/></span></div><div class="metrics-grid"><div><span>红球和值</span><b>{{ result.sum }}</b></div><div><span>红球跨度</span><b>{{ result.span }}</b></div><div><span>奇偶比例</span><b>{{ result.odd_count }}<small> : </small>{{ result.even_count }}</b></div><div><span>三区比例</span><b>{{ result.zone_counts.join(' : ') }}</b></div></div><p>三区：01–11 / 12–22 / 23–33</p></div>
            <div v-else class="result-placeholder"><span class="placeholder-chart"><BarChart3 :size="23"/></span><span>选齐号码后，查看和值、跨度<br/>奇偶比例与三区分布。</span></div>
            <div class="selection-footnote"><CircleHelp :size="15"/><p>统计用于观察历史分布，不代表未来开奖概率。随机号码不作中奖预测。</p></div>
          </aside>
        </div>
      </template>
      <footer class="page-footer"><span class="footer-brand">球序 <span>SSQ LAB</span></span><span>观察历史 · 理性选择</span><span>数据来自历史开奖记录</span></footer>
    </main>
    <div v-if="records.length && !loading && !loadError" class="mobile-dock"><button class="dock-selection" @click="tab='picker'" aria-label="打开号码选择"><span>我的号码 <ChevronRight :size="12"/></span><b>{{ chosenReds.length }} / 6 <small>红球</small> <i>·</i> {{ blue?'1 / 1':'0 / 1' }} <small>蓝球</small></b></button><button class="dock-random" @click="randomize" aria-label="随机生成号码"><Shuffle :size="18"/></button><button class="button dock-analyze" @click="analyze" :disabled="!complete || analyzing"><LoaderCircle v-if="analyzing" :size="14" class="spinning"/><BarChart3 v-else :size="14"/>{{ analyzing?'分析中':'分析组合' }}</button></div>
    <Transition name="toast"><div v-if="notice" class="toast-notice" role="status"><Check :size="16"/>{{ notice }}</div></Transition>
  </div>
</template>
