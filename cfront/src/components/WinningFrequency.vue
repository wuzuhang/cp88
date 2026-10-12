<script setup>
import { formatOmission } from '../lib/analysis.js'
defineProps({
  frequency: { type:Object, required:true },
  labels: { type:Array, required:true },
  numbers: { type:Array, required:true },
  sampleLabel: { type:String, required:true },
  omissions: { type:Array, default:()=>[] },
})
</script>

<template>
  <div class="winning-frequency">
    <div class="winning-frequency-heading"><span>中奖号码频次合值</span><strong>{{frequency.total}}<small>次</small></strong></div>
    <div class="winning-frequency-positions"><span v-for="(count,p) in frequency.counts" :key="p" :title="`${labels[p]}号码 ${numbers[p]}，出现 ${count} 次，遗漏 ${formatOmission(omissions[p])} 期`"><small>{{labels[p]}} · {{numbers[p]}}</small><b>{{count}} 次</b><small v-if="omissions[p]">遗漏 {{formatOmission(omissions[p])}} 期</small></span></div>
    <p>{{sampleLabel}} · {{frequency.counts.join(' + ')}} = {{frequency.total}}</p>
    <p v-if="omissions.length">遗漏从目标期的上一期起算，上一期同位置出现为 0；≥ 表示样本内未出现。</p>
  </div>
</template>

<style>
.winning-frequency{margin-top:15px;padding-top:13px;border-top:1px solid #edf0e8}.winning-frequency-heading{display:flex;align-items:center;justify-content:space-between;gap:8px;font-size:11px;color:#859179}.winning-frequency-heading strong{font-size:19px;color:#304d3e;font-variant-numeric:tabular-nums}.winning-frequency-heading strong small{font-size:10px;font-weight:400;margin-left:4px;color:#89977e}.winning-frequency-positions{display:flex;flex-wrap:wrap;gap:10px 16px;margin-top:9px}.winning-frequency-positions>span{display:flex;flex-direction:column;gap:4px}.winning-frequency-positions small{font-size:9px;color:#9aa390}.winning-frequency-positions b{font-size:12px;font-weight:500;color:#677c5b;font-variant-numeric:tabular-nums}.winning-frequency p{margin-top:9px;font-size:9px;color:#9aa390;line-height:1.7;overflow-wrap:anywhere}
</style>
