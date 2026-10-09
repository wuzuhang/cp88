<script setup>
import { onBeforeUnmount, ref, watch } from 'vue'
import SSQ from './SSQ.vue'
import Fucai3D from './Fucai3D.vue'
const page=ref(location.hash.startsWith('#/fucai3d')?'fucai3d':'ssq')
const updatePage=()=>{page.value=location.hash.startsWith('#/fucai3d')?'fucai3d':'ssq'}
watch(page,game=>{document.title=game==='fucai3d'?'球序 · 福彩 3D 分析':'球序 · 双色球分析';window.scrollTo(0,0)},{immediate:true})
window.addEventListener('hashchange',updatePage)
onBeforeUnmount(()=>window.removeEventListener('hashchange',updatePage))
</script>

<template>
  <nav class="game-navigation" aria-label="玩法切换">
    <span>选择玩法</span>
    <a href="#/ssq" :class="{active:page==='ssq'}" :aria-current="page==='ssq'?'page':undefined">双色球</a>
    <a href="#/fucai3d" :class="{active:page==='fucai3d'}" :aria-current="page==='fucai3d'?'page':undefined">福彩 3D</a>
  </nav>
  <SSQ v-if="page==='ssq'"/>
  <Fucai3D v-else/>
</template>

<style>
.game-navigation{display:flex;align-items:center;gap:8px;padding:10px max(20px,calc((100vw - 1240px)/2));background:#f0f3eb;border-bottom:1px solid #e1e6dc;font-size:12px}
.game-navigation>span{color:#89947e;margin-right:8px;font-size:11px}
.game-navigation a{padding:8px 18px;border-radius:7px;color:#7a8870}
.game-navigation a.active{background:#304d3e;color:#fff;font-weight:600}
</style>
