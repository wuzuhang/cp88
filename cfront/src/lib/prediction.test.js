import { test } from 'node:test'
import assert from 'node:assert/strict'
import { predictionTargetIndex, sampleBeforeTarget } from './analysis.js'
import { buildRedAssignmentTable, buildSSQPredictionRules, comparePrediction, consecutiveRedPairs, dirichletMean, fitReferenceModel, generateReferences, normalizedEntropy, retainReferences } from './prediction.js'

test('双色球红球按集合比较，蓝球独立比较，出球顺序和数字交叉不影响命中',()=>{
  const draw={red_1:1,red_2:3,red_3:8,red_4:12,red_5:25,red_6:33,blue:8}
  const row={reds:[33,12,1,2,4,7],blue:3}
  assert.deepEqual(comparePrediction('ssq',row,draw),{matches:[true,true,true,false,false,false],redCount:3,blueMatch:false})
  assert.deepEqual(comparePrediction('ssq',{reds:[1,3,8,12,25,33],blue:8},draw),{matches:Array(6).fill(true),redCount:6,blueMatch:true})
  assert.equal(comparePrediction('ssq',{reds:[2,4,5,6,7,9],blue:8},draw).redCount,0)
})

test('3D按位置比较，支持0和重复数字，同数字错位不命中',()=>{
  const draw={num_1:0,num_2:5,num_3:0}
  assert.deepEqual(comparePrediction('fucai3d',{numbers:[0,0,0]},draw),{matches:[true,false,true],positionCount:2})
  assert.deepEqual(comparePrediction('fucai3d',{numbers:[5,0,5]},draw),{matches:[false,false,false],positionCount:0})
  assert.deepEqual(comparePrediction('fucai3d',{numbers:[0,5,0]},draw),{matches:[true,true,true],positionCount:3})
})

test('未开奖时不生成命中状态，对照过程不修改预测号码或中奖号码',()=>{
  for(const game of ['ssq','fucai3d']){
    const row=game==='ssq'?{reds:[1,2,3,4,5,6],blue:1}:{numbers:[0,5,0]}
    const draw=game==='ssq'?ssqDraw(0):{num_1:0,num_2:5,num_3:0}
    const before=JSON.stringify({row,draw})
    assert.equal(comparePrediction(game,row,undefined),null)
    assert.equal(comparePrediction(game,row,null),null)
    comparePrediction(game,row,draw)
    assert.equal(JSON.stringify({row,draw}),before)
  }
})

function seeded(seed){let state=seed>>>0;return ()=>{state=(1664525*state+1013904223)>>>0;return state/4294967296}}
function ssqDraw(index){
  const sequence=Array.from({length:6},(_,p)=>(index+p)%33+1)
  const reds=[...sequence].sort((a,b)=>a-b)
  return Object.fromEntries([['issue',String(2026000+1000-index)],['blue',index%16+1],...sequence.map((n,p)=>[`seq_red_${p+1}`,n]),...reds.map((n,p)=>[`red_${p+1}`,n])])
}
function digitDraw(index){return {issue:String(2026000+1000-index),num_1:Math.floor(index/100)%10,num_2:Math.floor(index/10)%10,num_3:index%10}}

test('Dirichlet平滑保持权重归一且未出现号码有正权重，熵只描述分布',()=>{
  const probabilities=dirichletMean([20,0,0,0])
  assert.ok(Math.abs(probabilities.reduce((sum,n)=>sum+n,0)-1)<1e-12)
  assert.ok(probabilities.every(n=>n>0&&n<1))
  assert.equal(probabilities[0],(20+50/4)/70)
  assert.equal(normalizedEntropy([.25,.25,.25,.25]),1)
  assert.equal(normalizedEntropy([1,0,0,0]),0)
  assert.ok(normalizedEntropy(probabilities)<1)
})

test('双色球动态规划的均匀模型归一权重等于33P6/33的6次方，位置换序不改变总权重',()=>{
  const positions=Array.from({length:6},()=>Array(33).fill(1/33))
  const weight=buildRedAssignmentTable(positions)[0][63]
  const expected=Array.from({length:6},(_,p)=>(33-p)/33).reduce((product,p)=>product*p,1)
  assert.ok(Math.abs(weight-expected)<1e-12)
  const nonuniform=Array.from({length:6},(_,p)=>dirichletMean(Array.from({length:33},(_,n)=>(p+n)%7)))
  assert.ok(Math.abs(buildRedAssignmentTable(nonuniform)[0][63]-buildRedAssignmentTable([...nonuniform].reverse())[0][63])<1e-12)
})

test('双色球独立生成10注，注内红球合法且不重复，不修改所选样本',()=>{
  const sample=Array.from({length:500},(_,i)=>ssqDraw(i)),before=JSON.stringify(sample)
  const result=generateReferences('ssq',sample,seeded(23))
  assert.equal(result.rows.length,10)
  assert.equal(result.sampleCount,500)
  assert.ok(result.entropy>=0&&result.entropy<=1)
  for(const row of result.rows){
    assert.equal(row.reds.length,6)
    assert.equal(new Set(row.reds).size,6)
    assert.ok(row.reds.every(n=>Number.isInteger(n)&&n>=1&&n<=33))
    assert.deepEqual(row.reds,[...row.reds].sort((a,b)=>a-b))
    assert.ok(Number.isInteger(row.blue)&&row.blue>=1&&row.blue<=16)
    assert.equal(row.sum,row.reds.reduce((sum,n)=>sum+n,0))
    assert.equal('logWeight' in row,false)
    assertSSQConditions(row,result.rules)
  }
  assert.equal(JSON.stringify(sample),before)
})

test('3D完整枚举1000种有序结果，保留0和重复数字',()=>{
  const sample=Array.from({length:1000},(_,i)=>digitDraw(i))
  const model=fitReferenceModel('fucai3d',sample)
  assert.ok(Math.abs(model.entropy-1)<1e-12)
  const result=generateReferences('fucai3d',sample,()=>0)
  assert.equal(result.candidateCount,1000)
  assert.equal(result.rows.length,10)
  assert.deepEqual(result.rows[0].numbers,[0,0,0])
  assert.equal(result.rows[0].pattern,'豹子')
  assert.equal(new Set(result.rows.map(row=>row.key)).size,10)
  assert.ok(result.rows.every(row=>row.numbers.length===3&&row.numbers.every(n=>Number.isInteger(n)&&n>=0&&n<=9)))
})

function assertSSQConditions(row,rules) {
  const repeated=row.reds.filter(number=>rules.latestReds.includes(number))
  assert.ok(repeated.length>=1&&repeated.length<=2)
  assert.equal(row.repeatCount,repeated.length)
  assert.ok(row.reds.some((number,index)=>index>0&&number===row.reds[index-1]+1))
  assert.equal(row.consecutivePairs.length,1)
  assert.equal(row.reds.filter((number,index)=>index>0&&number===row.reds[index-1]+1).length,1)
  assert.equal(new Set(row.sequence).size,6)
  assert.deepEqual([...row.sequence].sort((a,b)=>a-b),row.reds)
  row.sequence.forEach((number,position)=>assert.ok(!rules.redExcluded[position].includes(number)))
  assert.notEqual(row.blue,rules.latestBlue)
  assert.ok(!rules.blueExcluded.includes(row.blue))
}

test('双色球五条条件在每注同时成立，并按各自出球位置排除遗漏极值',()=>{
  const sample=Array.from({length:500},(_,i)=>ssqDraw(i))
  for(const seed of [1,23,88]){
    const result=generateReferences('ssq',sample,seeded(seed))
    assert.equal(result.rows.length,10)
    assert.equal(result.rules.latestIssue,sample[0].issue)
    for(const row of result.rows)assertSSQConditions(row,result.rules)
  }
  const rules=buildSSQPredictionRules(sample)
  assert.ok(rules.redExcluded[0].includes(1))
  assert.equal(rules.redExcluded[2].includes(1),false)
  assert.ok(rules.blueExcluded.includes(sample[0].blue))
})

test('遗漏极值并列全部排除，样本内未出现按样本期数下界处理',()=>{
  const rules=buildSSQPredictionRules([ssqDraw(0),ssqDraw(1)])
  assert.deepEqual(rules.redExtrema[0].minimumNumbers,[1])
  assert.equal(rules.redExtrema[0].maximum,2)
  assert.equal(rules.redExtrema[0].maximumNumbers.length,31)
  assert.equal(rules.redExcluded[0].length,32)
  assert.equal(rules.redExcluded[0].includes(2),false)
  assert.equal(rules.blueExcluded.length,15)
  assert.equal(rules.blueExcluded.includes(2),false)
})

test('双色球无解时明确报错，不放宽条件、不生成不足或不合规的号码',()=>{
  assert.throws(()=>generateReferences('ssq',[ssqDraw(0)]),/没有可用蓝球/)
  const noRedSupport=Array.from({length:500},(_,i)=>({...ssqDraw(0),blue:i%16+1}))
  assert.throws(()=>generateReferences('ssq',noRedSupport),/无法同时满足红球/)
  assert.throws(()=>generateReferences('ssq',[ssqDraw(0),ssqDraw(1)]),/无法同时满足红球/)
})

test('有条件的双色球注间仍独立，允许10注完全相同且随机源退化也不会死循环',()=>{
  const sample=Array.from({length:500},(_,i)=>ssqDraw(i))
  for(const value of [0,1]){
    const result=generateReferences('ssq',sample,()=>value)
    assert.equal(result.rows.length,10)
    assert.equal(new Set(result.rows.map(row=>row.key)).size,1)
    for(const row of result.rows)assertSSQConditions(row,result.rules)
  }
})

test('在可用蓝球中继续按样本权重独立抽样，不施加跨注平衡',()=>{
  const sample=Array.from({length:500},(_,i)=>({...ssqDraw(i),blue:i===0?1:(i>=10&&i<24?i-7:2)}))
  const random=seeded(82)
  const rows=Array.from({length:20},()=>generateReferences('ssq',sample,random).rows).flat()
  assert.ok(rows.filter(row=>row.blue===2).length>150)
  assert.ok(rows.every(row=>row.blue!==1))
})

test('历史对照的模型只读取目标期之前的所选样本',()=>{
  for(const game of ['ssq','fucai3d']){
    const draw=game==='ssq'?ssqDraw:digitDraw
    const records=Array.from({length:510},(_,i)=>draw(i))
    const target=predictionTargetIndex(4)
    const before=generateReferences(game,sampleBeforeTarget(records,target,500),seeded(34))
    for(let i=0;i<=target;i++)records[i]=draw(i+100)
    const after=generateReferences(game,sampleBeforeTarget(records,target,500),seeded(34))
    assert.deepEqual(after,before)
    assert.equal(before.sampleCount,500)
  }
})

test('空样本、非法号码和未知玩法提供错误而非伪造统计',()=>{
  assert.throws(()=>generateReferences('ssq',[]),/没有可分析/)
  assert.throws(()=>generateReferences('unknown',[ssqDraw(0)]),/不支持/)
  assert.throws(()=>generateReferences('fucai3d',[{num_1:10,num_2:0,num_3:0}]),/无效/)
  assert.throws(()=>generateReferences('ssq',[{...ssqDraw(0),seq_red_2:1}]),/无效/)
})


test('红球恰好一组两连号：禁止三连号、多组连号和无连号',()=>{
  assert.deepEqual(consecutiveRedPairs([1,3,5,10,11,20]),[[10,11]])
  assert.deepEqual(consecutiveRedPairs([11,1,20,3,10,5]),[[10,11]])
  assert.equal(consecutiveRedPairs([5,6,7,12,18,25]).length,2)
  assert.equal(consecutiveRedPairs([1,2,8,9,20,30]).length,2)
  assert.equal(consecutiveRedPairs([5,6,7,8,18,25]).length,3)
  assert.equal(consecutiveRedPairs([1,5,10,17,25,33]).length,0)
})

test('多次独立预测的每注红球只有一个相邻数对，其他条件仍同时满足',()=>{
  const sample=Array.from({length:500},(_,i)=>ssqDraw(i))
  for(let seed=1;seed<=12;seed++){
    const result=generateReferences('ssq',sample,seeded(seed))
    assert.equal(result.rows.length,10)
    for(const row of result.rows)assertSSQConditions(row,result.rules)
  }
})


test('双色球完整保留全部10注，顺序、号码及原始序号不变',()=>{
  const sample=Array.from({length:500},(_,i)=>ssqDraw(i))
  const generated=generateReferences('ssq',sample,seeded(23))
  const before=JSON.stringify(generated)
  const retained=retainReferences('ssq',generated)
  assert.equal(retained.generatedCount,10)
  assert.equal(retained.rows.length,10)
  assert.deepEqual(retained.rows.map(row=>row.ordinal),Array.from({length:10},(_,index)=>index+1))
  for(const [index,ordinal] of Array.from({length:10},(_,index)=>index+1).entries()){
    assert.deepEqual(retained.rows[index],{...generated.rows[ordinal-1],ordinal})
    assertSSQConditions(retained.rows[index],retained.rules)
  }
  assert.equal(JSON.stringify(generated),before)
})

test('3D仍保留全部10注及原序号',()=>{
  const generated=generateReferences('fucai3d',Array.from({length:1000},(_,i)=>digitDraw(i)),seeded(23))
  const retained=retainReferences('fucai3d',generated)
  assert.equal(retained.rows.length,10)
  assert.deepEqual(retained.rows.map(row=>row.ordinal),Array.from({length:10},(_,i)=>i+1))
  assert.deepEqual(retained.rows.map(row=>row.key),generated.rows.map(row=>row.key))
})
