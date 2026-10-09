import { test } from 'node:test'
import assert from 'node:assert/strict'
import { buildMatrix, frequencyLevel, randomCombination } from './analysis.js'

function draw(reds,blue) {
  return Object.fromEntries([...reds.map((n,i) => [`red_${i+1}`,n]),...reds.map((n,i) => [`seq_red_${i+1}`,n]),['blue',blue]])
}

test('按出球位置计数、排序并按历史偏移排除对照期',()=>{
  const records=[draw([1,2,3,4,5,6],1),draw([7,8,9,10,11,12],2),draw([7,8,9,10,11,12],2)]
  const result=buildMatrix(records,1,2,1)
  assert.equal(result.redCount,2)
  assert.equal(result.blueCount,1)
  assert.equal(result.totals[0],0)
  assert.equal(result.totals[6],2)
  assert.deepEqual(result.columns[0].at(-1),{number:7,count:2})
  assert.deepEqual(result.blue.at(-1),{number:2,count:1})
  assert.equal(result.hot[0].number,7)
})

test('随机组合包含六个不同红球和合法蓝球',()=>{
  for(let i=0;i<100;i++) {
    const {reds,blue}=randomCombination()
    assert.equal(new Set(reds).size,6)
    assert.ok(reds.every(n=>n>=1&&n<=33))
    assert.ok(blue>=1&&blue<=16)
  }
})

test('空历史不会生成虚构频次',()=>{
  const result=buildMatrix([])
  assert.equal(result.redCount,0)
  assert.equal(result.blueCount,0)
  assert.ok(result.totals.every(n=>n===0))
})

test('色阶随样本期数归一化，100 与 1000 期的相同出现比例颜色一致',()=>{
  for (const numberCount of [33,16]) {
    for (const count of [0,1,2,3,4,5,6,10]) {
      assert.equal(frequencyLevel(count,100,numberCount),frequencyLevel(count*10,1000,numberCount))
    }
  }
})

test('大样本的红蓝球都保留从浅到深的五个色阶',()=>{
  assert.deepEqual([15,20,30,40,50].map(count=>frequencyLevel(count,1000,33)),[0,1,2,3,4])
  assert.deepEqual([30,45,62,90,110].map(count=>frequencyLevel(count,1000,16)),[0,1,2,3,4])
  assert.equal(frequencyLevel(5,0,33),0)
  assert.equal(frequencyLevel(0,1000,16),0)
  assert.equal(frequencyLevel(50,800,16),2)
})
