<script setup lang="ts">import {computed,onMounted,ref} from 'vue'; import Message from 'primevue/message'; import Tag from 'primevue/tag'; import {useAuthStore} from '../../../stores/auth'; import {useCrmStore} from '../../../stores/crm'; const auth=useAuthStore(),crm=useCrmStore(),error=ref(''); const active=computed(()=>crm.myProspects.filter(v=>!['LOST','CONVERTED'].includes(v.status))); const completed=computed(()=>crm.myProspects.filter(v=>['WON','LOST','CONVERTED'].includes(v.status)).length); onMounted(async()=>{try{await Promise.all([crm.loadMyProspects(),crm.loadMyCustomers()])}catch(e){error.value=crm.errorMessage(e)}})</script>
<template>
  <section class="sales-home">
    <Message v-if="error" severity="error">{{ error }}</Message>

    <RouterLink class="ready-card" to="/sales/my-prospects">
      <div>
        <strong>Your day is ready</strong>
        <span>{{ active.length }} visits planned · First at 09:30</span>
      </div>
      <i class="pi pi-arrow-right" />
    </RouterLink>

    <div class="section-title">
      <strong>Quick statistics</strong>
      <RouterLink to="/sales/history">View all</RouterLink>
    </div>

    <div class="quick-stats">
      <RouterLink to="/sales/my-customers">
        <div class="quick-stats-info">
          <span class="stat-icon blue-dot">C</span>
          <small>Today's customers</small>
        </div>
        <strong>{{ crm.myCustomers.length }}</strong>
      </RouterLink>

      <RouterLink to="/sales/my-prospects">
        <div class="quick-stats-info">
          <span class="stat-icon amber-dot">P</span>
          <small>Today's prospects</small>
        </div>
        <strong>{{ active.length }}</strong>
      </RouterLink>

      <RouterLink to="/sales/history">
        <div class="quick-stats-info">
          <span class="stat-icon green-dot"><i class="pi pi-check" /></span>
          <small>Completed visits</small>
        </div>
        <strong>{{ completed }}</strong>
      </RouterLink>

      <RouterLink to="/sales/my-prospects">
        <div class="quick-stats-info">
          <span class="stat-icon red-dot"><i class="pi pi-clock" /></span>
          <small>Pending visits</small>
        </div>
        <strong>{{ active.filter(v => v.status === 'NEGOTIATION' || v.status === 'CONTACTED').length }}</strong>
      </RouterLink>
    </div>

    <div class="section-title">
      <strong>Quick actions</strong>
    </div>

    <div class="quick-actions">
      <RouterLink to="/sales/my-prospects" class="action-primary">
        <span class="action-icon"><i class="pi pi-play" /></span>
        <span>Start visit</span>
      </RouterLink>

      <RouterLink to="/sales/my-prospects">
        <span class="action-icon"><i class="pi pi-map-marker" /></span>
        <span>Open maps</span>
      </RouterLink>

      <RouterLink to="/sales/my-customers">
        <span class="action-icon text-icon">C</span>
        <span>Customer</span>
      </RouterLink>

      <RouterLink to="/sales/my-prospects">
        <span class="action-icon text-icon">P</span>
        <span>Prospect</span>
      </RouterLink>
    </div>

    <div class="section-title">
      <strong>Today's visits</strong>
      <RouterLink to="/sales/my-prospects">See route</RouterLink>
    </div>

    <div class="today-list">
      <RouterLink
        v-for="(item, index) in active.slice(0, 3)"
        :key="item.id"
        :to="`/sales/my-prospects/${item.id}`"
      >
        <time>
          {{ `${9 + index * 2}:30` }}
          <small>Today</small>
        </time>
        <i class="visit-dot" :class="index % 2 === 0 ? 'dot-amber' : 'dot-blue'" />
        <div>
          <strong>{{ item.placeName }}</strong>
          <small>{{ item.placeCategory || item.industryGroup }} · 1.8 km away</small>
        </div>
        <span class="visit-badge">Pending</span>
      </RouterLink>

      <div v-if="!active.length" class="empty-state">
        <strong>No visits queued</strong>
      </div>
    </div>
  </section>
</template>
