import { test } from 'node:test'
import assert from 'node:assert/strict'
import { build3DMatrix,random3D } from './fucai3d.js'
import { frequencyLevel, positionFrequencyTotal, predictionTargetIndex } from './analysis.js'

test('3D下一期预测包含本期已开奖，历史中奖号码与本期之前的样本分别对应',()=>{
  const records=[{num_1:0,num_2:0,num_3:0},{num_1:1,num_2:2,num_3:3},{num_1:1,num_2:2,num_3:3}]
  const initial=build3DMatrix(records,predictionTargetIndex(0),2)
  assert.equal(records[predictionTargetIndex(0)],undefined)
  assert.deepEqual(positionFrequencyTotal(initial.columns,[0,0,0]),{counts:[1,1,1],total:3})
  assert.deepEqual(initial.omissions.map(column=>column[0]),Array(3).fill({periods:0,exact:true}))
  const historical=build3DMatrix(records,predictionTargetIndex(1),2)
  assert.equal(records[predictionTargetIndex(1)],records[0])
  assert.deepEqual(positionFrequencyTotal(historical.columns,[0,0,0]),{counts:[0,0,0],total:0})
  assert.deepEqual(historical.omissions.map(column=>column[0]),Array(3).fill({periods:2,exact:false}))
  assert.deepEqual(build3DMatrix(records,predictionTargetIndex(0),2),initial)
})

test('3D在500期真实基数下排除本期及更新开奖，零和重复号码按位置重算',()=>{
  const records=Array.from({length:506},()=>({num_1:1,num_2:2,num_3:3}))
  for(const i of [0,1,2,502,503])records[i]={num_1:0,num_2:0,num_3:0}
  const target=build3DMatrix(records,2,500)
  assert.equal(target.count,500)
  assert.deepEqual(positionFrequencyTotal(target.columns,[0,0,0]),{counts:[1,1,1],total:3})
  assert.deepEqual(target.omissions.map(column=>column[0]),Array(3).fill({periods:499,exact:true}))
  const older=build3DMatrix(records,3,500)
  assert.deepEqual(positionFrequencyTotal(older.columns,[0,0,0]),{counts:[2,2,2],total:6})
  assert.deepEqual(older.omissions.map(column=>column[0]),Array(3).fill({periods:498,exact:true}))
  assert.deepEqual(build3DMatrix(records,2,500),target)
  assert.equal(build3DMatrix(records,0,500).columns[0].find(item=>item.number===0).count,2)
})

test('3D 中奖频次合值保留零与重复数字，仅按对应位置计数',()=>{
  const records=[{num_1:0,num_2:0,num_3:0},{num_1:0,num_2:5,num_3:0},{num_1:0,num_2:0,num_3:9}]
  const matrix=build3DMatrix(records,0,2)
  assert.deepEqual(positionFrequencyTotal(matrix.columns,[0,0,0]),{counts:[2,1,1],total:4})
  assert.deepEqual(positionFrequencyTotal(matrix.columns,[0,5,0]),{counts:[2,1,1],total:4})
  assert.deepEqual(positionFrequencyTotal(build3DMatrix([]).columns,[0,0,0]),{counts:[0,0,0],total:0})
})

test('3D 按位置保留 0 与重复数字，历史对照排除目标期',()=>{
  const records=[{num_1:9,num_2:9,num_3:9},{num_1:0,num_2:5,num_3:0},{num_1:0,num_2:0,num_3:0}]
  const matrix=build3DMatrix(records,0,2)
  assert.equal(matrix.count,2)
  assert.equal(matrix.totals[0],5)
  assert.equal(matrix.totals[5],1)
  assert.equal(matrix.totals[9],0)
  assert.deepEqual(matrix.columns[0].at(-1),{number:0,count:2})
  assert.deepEqual(matrix.columns[1].at(-2),{number:0,count:1})
  assert.equal(matrix.hot[0].number,0)
})
test('3D 100 和 1000 期同频率色阶一致',()=>{
  assert.deepEqual([4,7,10,14,17].map(n=>frequencyLevel(n,100,10)),[0,1,2,3,4])
  for(const count of [0,4,7,10,14,17])assert.equal(frequencyLevel(count,100,10),frequencyLevel(count*10,1000,10))
})
test('3D 随机号码允许全零、重复且保留三位',()=>{
  assert.deepEqual(random3D(()=>0),[0,0,0])
  assert.deepEqual(random3D(()=>.999),[9,9,9])
  for(let i=0;i<100;i++)assert.ok(random3D().every(n=>Number.isInteger(n)&&n>=0&&n<=9))
  assert.ok(build3DMatrix([]).totals.every(count=>count===0))
})
