<script setup lang="ts">import { computed, onMounted, ref } from 'vue'; import Message from 'primevue/message'; import Tag from 'primevue/tag'; import { useCrmStore } from '../../stores/crm'; const crm=useCrmStore(); const error=ref(''); const outcomes=computed(()=>crm.myProspects.filter((v)=>v.status==='WON'||v.status==='LOST')); onMounted(async()=>{try{await crm.loadMyProspects()}catch(e){error.value=crm.errorMessage(e)}})</script>
<template><section class="mobile-page"><div class="mobile-title"><div><p class="eyebrow">Outcome history</p><h1>History</h1></div></div><Message v-if="error" severity="error">{{ error }}</Message><div class="mobile-card-list"><RouterLink v-for="item in outcomes" :key="item.id" class="history-card" :to="`/sales/my-prospects/${item.id}`"><div><strong>{{ item.placeName }}</strong><span>{{ item.industryGroup }} · {{ new Date(item.updatedAt).toLocaleDateString() }}</span></div><Tag :value="item.status" :severity="item.status==='WON'?'success':'danger'" /></RouterLink><div v-if="!outcomes.length&&!crm.loading" class="empty-state"><i class="pi pi-history" /><strong>No completed outcomes yet</strong></div></div></section></template>

<style scoped>
.history-card {
  display: flex; justify-content: space-between; align-items: center;
  padding: 1rem 1.25rem; background: var(--surface-card); border: 1px solid var(--border-light);
  border-radius: var(--radius-xl); text-decoration: none; color: var(--text-primary);
  box-shadow: var(--shadow-sm); transition: all 0.2s ease;
}
.history-card:hover { border-color: var(--border-default); box-shadow: var(--shadow-md); transform: translateY(-2px); }
.history-card div { display: grid; gap: 0.15rem; }
.history-card strong { font-size: 0.95rem; font-weight: 700; }
.history-card span { color: var(--text-muted); font-size: 0.8rem; }
</style>
