export const ball = (n) => String(n).padStart(2, '0')
export const redBalls = (record) => Array.from({length:6}, (_,i) => record[`red_${i+1}`])
export const sequenceBalls = (record) => Array.from({length:6}, (_,i) => record[`seq_red_${i+1}`])

// 每个出球位置每期出现一个号码，以实际样本期数计算平均频次。
// 相同出现比例在不同样本规模下保持相同颜色，平均频次落在中间色阶。
export function frequencyLevel(count, sampleSize, numberCount) {
  if (count <= 0 || sampleSize <= 0 || numberCount <= 0) return 0
  const ratio = count / (sampleSize / numberCount)
  if (ratio <= 0.5) return 0
  if (ratio <= 0.85) return 1
  if (ratio <= 1.15) return 2
  if (ratio <= 1.5) return 3
  return 4
}

export function frequencies(records, field, size) {
  const counts = Array(size).fill(0)
  for (const record of records) {
    const value = record[field]
    if (Number.isInteger(value) && value >= 1 && value <= size) counts[value-1]++
  }
  return counts.map((count,i) => ({number:i+1,count})).sort((a,b) => a.count-b.count || a.number-b.number)
}

export function buildMatrix(records, offset=0, redWindow=100, blueWindow=100) {
  const redSample = records.slice(offset,offset+redWindow)
  const blueSample = records.slice(offset,offset+blueWindow)
  const totals = Array(33).fill(0)
  for (const record of redSample) for (const n of redBalls(record)) if (n>=1 && n<=33) totals[n-1]++
  return {
    columns: Array.from({length:6},(_,i) => frequencies(redSample,`seq_red_${i+1}`,33)),
    blue: frequencies(blueSample,'blue',16),
    totals, redCount:redSample.length, blueCount:blueSample.length,
    hot: totals.map((count,i) => ({number:i+1,count})).sort((a,b) => b.count-a.count || a.number-b.number).slice(0,5),
  }
}

export function randomCombination(random=Math.random) {
  const pool = Array.from({length:33},(_,i) => i+1)
  for (let i=pool.length-1; i>0; i--) {
    const j=Math.floor(random()*(i+1))
    ;[pool[i],pool[j]]=[pool[j],pool[i]]
  }
  return {reds:pool.slice(0,6),blue:Math.floor(random()*16)+1}
}
