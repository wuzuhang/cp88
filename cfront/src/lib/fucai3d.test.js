import { test } from 'node:test'
import assert from 'node:assert/strict'
import { build3DMatrix,random3D } from './fucai3d.js'
import { frequencyLevel } from './analysis.js'

test('3D 按位置保留 0 与重复数字，历史对照排除目标期',()=>{
  const records=[{num_1:9,num_2:9,num_3:9},{num_1:0,num_2:5,num_3:0},{num_1:0,num_2:0,num_3:0}]
  const matrix=build3DMatrix(records,1,2)
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
