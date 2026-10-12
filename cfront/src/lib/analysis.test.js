import { test } from 'node:test'
import assert from 'node:assert/strict'
import { buildMatrix, formatOmission, frequencyLevel, positionFrequencyTotal, positionOmissions, randomCombination, sampleBeforeTarget } from './analysis.js'

test('500期样本严格取目标期之前500期，前后切换同步重算频次和遗漏',()=>{
  const records=Array.from({length:506},(_,i)=>({...draw([7,8,9,10,11,12],2),issue:String(10000-i)}))
  for(const i of [0,1,2,502,503])Object.assign(records[i],draw([1,2,3,4,5,6],1))
  const sample=sampleBeforeTarget(records,2,500)
  assert.equal(sample.length,500)
  assert.equal(sample[0].issue,records[3].issue)
  assert.equal(sample.at(-1).issue,records[502].issue)
  const target=buildMatrix(records,2,500,500)
  assert.equal(target.redCount,500)
  assert.equal(target.columns[0].find(item=>item.number===1).count,1)
  assert.equal(target.blue.find(item=>item.number===1).count,1)
  assert.deepEqual(target.omissions[0][0],{periods:499,exact:true})
  assert.deepEqual(target.blueOmissions[0],{periods:499,exact:true})
  const older=buildMatrix(records,3,500,500)
  assert.equal(older.columns[0].find(item=>item.number===1).count,2)
  assert.deepEqual(older.omissions[0][0],{periods:498,exact:true})
  assert.deepEqual(buildMatrix(records,2,500,500),target)
  assert.equal(buildMatrix(records,0,500,500).columns[0].find(item=>item.number===1).count,2)
})

test('遗漏上一期出现为0，未出现使用样本长度下界，空样本不伪造值',()=>{
  const omissions=positionOmissions([{blue:1},{blue:2},{blue:1}],'blue',3)
  assert.deepEqual(omissions,[{periods:0,exact:true},{periods:1,exact:true},{periods:3,exact:false}])
  assert.equal(formatOmission(omissions[0]),'0')
  assert.equal(formatOmission(omissions[2]),'≥3')
  assert.equal(formatOmission(positionOmissions([],'blue',3)[0]),'—')
})

function draw(reds,blue) {
  return Object.fromEntries([...reds.map((n,i) => [`red_${i+1}`,n]),...reds.map((n,i) => [`seq_red_${i+1}`,n]),['blue',blue]])
}

test('双色球中奖频次按出球位置和蓝球对应号码求合值，保留独立样本窗口',()=>{
  const target=draw([1,2,3,4,5,6],1)
  const records=[target,draw([2,1,3,4,5,6],1),draw([1,2,3,4,5,6],2)]
  const matrix=buildMatrix(records,0,2,1)
  const summary=positionFrequencyTotal([...matrix.columns,matrix.blue],[1,2,3,4,5,6,1])
  assert.deepEqual(summary.counts,[1,1,2,2,2,2,1])
  assert.equal(summary.total,11)
  const sorted=positionFrequencyTotal([...matrix.columns.map(column=>[...column].reverse()),[...matrix.blue].reverse()],[1,2,3,4,5,6,1])
  assert.deepEqual(sorted,summary)
  assert.deepEqual(positionFrequencyTotal([],[]),{counts:[],total:0})
})

test('按出球位置计数、排序并按历史偏移排除对照期',()=>{
  const records=[draw([1,2,3,4,5,6],1),draw([7,8,9,10,11,12],2),draw([7,8,9,10,11,12],2)]
  const result=buildMatrix(records,0,2,1)
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
