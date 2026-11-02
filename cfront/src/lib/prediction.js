import { redBalls, sequenceBalls } from './analysis.js'
import { digits3D } from './fucai3d.js'

export const PRIOR_STRENGTH = 50
const DIVERSITY_PENALTY = 2

// 开奖结果仅用于生成后的对照，不参与模型拟合或选号。
export function comparePrediction(game,row,winningRecord) {
  if(!winningRecord)return null
  if(game==='ssq') {
    const actual=new Set(redBalls(winningRecord))
    const matches=row.reds.map(number=>actual.has(number))
    return {matches,redCount:matches.filter(Boolean).length,blueMatch:row.blue===winningRecord.blue}
  }
  const actual=digits3D(winningRecord)
  const matches=row.numbers.map((number,position)=>number===actual[position])
  return {matches,positionCount:matches.filter(Boolean).length}
}

// 对称 Dirichlet 先验，总伪计数为 50；未出现的号码仍保留正权重。
export function dirichletMean(counts, strength=PRIOR_STRENGTH) {
  if(!counts.length || !Number.isFinite(strength) || strength<=0 || counts.some(n=>!Number.isFinite(n)||n<0)) throw new Error('无效的频次或先验参数')
  const total=counts.reduce((sum,n)=>sum+n,0)+strength
  return counts.map(count=>(count+strength/counts.length)/total)
}

export function normalizedEntropy(probabilities) {
  if(probabilities.length<2)return 0
  return Math.max(0,Math.min(1,-probabilities.reduce((sum,p)=>sum+(p>0?p*Math.log(p):0),0)/Math.log(probabilities.length)))
}

const unitRandom = random => Math.max(0,Math.min(1-Number.EPSILON,Number(random())||0))

function weightedIndex(weights,random) {
  let remaining=unitRandom(random)*weights.reduce((sum,n)=>sum+n,0)
  for(let i=0;i<weights.length;i++) { remaining-=weights[i];if(remaining<0)return i }
  for(let i=weights.length-1;i>=0;i--)if(weights[i]>0)return i
  throw new Error('模型权重无效')
}

// table[n][mask]：仅使用 n..32 号数字，填满 mask 中红球位置的总模型权重。
// 每个数字只能跳过或分配给一个位置，因此无需穷举全部组合即可严格约束注内不重复。
export function buildRedAssignmentTable(positions) {
  const table=Array.from({length:34},()=>new Float64Array(64))
  table[33][0]=1
  for(let number=32;number>=0;number--)for(let mask=0;mask<64;mask++) {
    let weight=table[number+1][mask]
    for(let position=0;position<6;position++)if(mask&(1<<position))weight+=positions[position][number]*table[number+1][mask^(1<<position)]
    table[number][mask]=weight
  }
  return table
}

function independentSSQ(model,random) {
  const table=buildRedAssignmentTable(model.positions)
  return Array.from({length:10},()=>{
    let mask=63
    const reds=[]
    for(let number=0;number<33&&mask;number++) {
      const weights=[table[number+1][mask],...model.positions.slice(0,6).map((column,p)=>mask&(1<<p)?column[number]*table[number+1][mask^(1<<p)]:0)]
      const choice=weightedIndex(weights,random)
      if(choice>0){reds.push(number+1);mask^=1<<(choice-1)}
    }
    if(mask)throw new Error('无法生成合法的红球组合')
    const blue=weightedIndex(model.positions[6],random)+1
    // 每注重新初始化 mask；不参考前面的注，不去重，也不施加跨注惩罚。
    return {key:`${reds.join(',')}|${blue}`,reds,blue,...combinationFeatures(reds)}
  })
}

function similarity3D(a,b) {
  return a.numbers.filter((n,p)=>n===b.numbers[p]).length/3
}

function selectDiverse3D(candidates,random) {
  const remaining=[...candidates],selected=[]
  // 从完整候选中不放回抽样。相似度惩罚只用于分散组合，不宣称提升中奖率。
  const maximum=Math.max(...remaining.map(row=>row.logWeight))
  while(selected.length<10 && remaining.length) {
    const weights=remaining.map(row=>Math.exp(row.logWeight-maximum-DIVERSITY_PENALTY*(selected.length?Math.max(...selected.map(other=>similarity3D(row,other))):0)))
    selected.push(remaining.splice(weightedIndex(weights,random),1)[0])
  }
  return selected.map(({logWeight,...row})=>row)
}

function combinationFeatures(numbers) {
  return {
    sum:numbers.reduce((sum,n)=>sum+n,0),
    span:Math.max(...numbers)-Math.min(...numbers),
    odd:numbers.filter(n=>n%2!==0).length,
    pattern:['','豹子','组三','组六'][new Set(numbers).size] || '',
  }
}

export function fitReferenceModel(game,sample) {
  if(game!=='ssq'&&game!=='fucai3d')throw new Error('不支持的玩法')
  if(!sample.length)throw new Error('当前窗口没有可分析的历史数据')
  const sizes=game==='ssq'?[33,33,33,33,33,33,16]:[10,10,10]
  const minimum=game==='ssq'?1:0
  const counts=sizes.map(size=>Array(size).fill(0))
  const sums=[],spans=[]
  for(const record of sample) {
    const numbers=game==='ssq'?[...sequenceBalls(record),record.blue]:digits3D(record)
    if(numbers.some((n,p)=>!Number.isInteger(n)||n<minimum||n>=minimum+sizes[p]) || (game==='ssq'&&new Set(numbers.slice(0,6)).size!==6))throw new Error('样本包含无效开奖号码，请先更新数据')
    numbers.forEach((number,p)=>counts[p][number-minimum]++)
    const values=game==='ssq'?numbers.slice(0,6):numbers
    const features=combinationFeatures(values)
    sums.push(features.sum);spans.push(features.span)
  }
  const positions=counts.map(column=>dirichletMean(column))
  const meanSum=sums.reduce((sum,n)=>sum+n,0)/sample.length
  return {
    positions,
    sampleCount:sample.length,
    firstIssue:sample.at(-1).issue || '—',
    lastIssue:sample[0].issue || '—',
    entropy:positions.reduce((sum,column)=>sum+normalizedEntropy(column),0)/positions.length,
    meanSum,
    sdSum:Math.sqrt(sums.reduce((sum,n)=>sum+(n-meanSum)**2,0)/sample.length),
    meanSpan:spans.reduce((sum,n)=>sum+n,0)/sample.length,
  }
}

export function generateReferences(game,sample,random=Math.random) {
  const model=fitReferenceModel(game,sample)
  const {positions,...summary}=model
  if(game==='ssq')return {...summary,rows:independentSSQ(model,random)}
  const candidates=new Map()
  // 3D 保留既有的完整枚举、分散和不放回抽样规则。
  for(let value=0;value<1000;value++) {
    const numbers=[Math.floor(value/100),Math.floor(value/10)%10,value%10]
    const key=numbers.join('')
    candidates.set(key,{key,numbers,...combinationFeatures(numbers),logWeight:numbers.reduce((sum,n,p)=>sum+Math.log(model.positions[p][n]),0)})
  }
  return { ...summary, candidateCount:candidates.size, rows:selectDiverse3D([...candidates.values()],random) }
}
