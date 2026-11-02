import { test } from 'node:test'
import assert from 'node:assert/strict'
import { predictionTargetIndex, sampleBeforeTarget } from './analysis.js'
import { buildRedAssignmentTable, comparePrediction, dirichletMean, fitReferenceModel, generateReferences, normalizedEntropy } from './prediction.js'

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

test('双色球注间不去重、不改变模型权重，允许10注完全相同',()=>{
  const result=generateReferences('ssq',[ssqDraw(0)],()=>0)
  assert.equal(result.rows.length,10)
  assert.equal(new Set(result.rows.map(row=>row.key)).size,1)
  assert.deepEqual(result.rows[0].reds,[28,29,30,31,32,33])
  assert.equal(result.rows[0].blue,1)
  const upper=generateReferences('ssq',[ssqDraw(0)],()=>1)
  assert.equal(new Set(upper.rows.map(row=>row.key)).size,1)
  assert.deepEqual(upper.rows[0].reds,[1,2,3,4,5,6])
  assert.equal(upper.rows[0].blue,16)
})

test('双色球按固定样本模型独立抽样，不以跨注平衡抑制高权重号码',()=>{
  const sample=Array.from({length:1000},()=>ssqDraw(0))
  const random=seeded(82)
  const rows=Array.from({length:20},()=>generateReferences('ssq',sample,random).rows).flat()
  assert.ok(rows.filter(row=>row.blue===1).length>150)
  assert.ok(rows.filter(row=>row.reds.filter(n=>n<=6).length>=4).length>150)
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
