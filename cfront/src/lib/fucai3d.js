export const positions3D = ['百位','十位','个位']
export const digits3D = record => [record.num_1,record.num_2,record.num_3]
export const trialDigits3D = record => [record.test_num_1,record.test_num_2,record.test_num_3]

export function build3DMatrix(records,offset=0,window=100) {
  const sample=records.slice(offset,offset+window)
  const counts=Array.from({length:3},()=>Array(10).fill(0))
  for(const record of sample) digits3D(record).forEach((digit,index)=>{
    if(Number.isInteger(digit)&&digit>=0&&digit<=9) counts[index][digit]++
  })
  const totals=Array.from({length:10},(_,digit)=>counts.reduce((sum,column)=>sum+column[digit],0))
  return {
    count:sample.length,
    columns:counts.map(column=>column.map((count,number)=>({number,count})).sort((a,b)=>a.count-b.count||a.number-b.number)),
    totals,
    hot:totals.map((count,number)=>({number,count})).sort((a,b)=>b.count-a.count||a.number-b.number).slice(0,5),
  }
}

export const random3D = (random=Math.random) => Array.from({length:3},()=>Math.floor(random()*10))
