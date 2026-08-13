<script setup lang="ts">
import type { CustomerSite } from '../../types/crm'

defineProps<{
  customer: CustomerSite
  operatingStatus?: string
}>()

function date(value: string) {
  return value ? new Intl.DateTimeFormat('id-ID', { day: 'numeric', month: 'short', year: 'numeric' }).format(new Date(value)) : '—'
}
function operating(value?: string) {
  return ({ OPERATIONAL: 'Beroperasi', CLOSED_TEMPORARILY: 'Tutup sementara', CLOSED_PERMANENTLY: 'Tutup permanen' } as Record<string, string>)[value || ''] ?? 'Belum diketahui'
}
</script>

<template>
  <section class="customer-snapshot-card">
    <header><div><p>Customer Snapshot</p><h2>Customer Aktif</h2></div><span>CRM saat ini</span></header>
    <dl>
      <div><dt>Assigned Sales</dt><dd>{{ customer.salesExecutiveName || 'Belum ditetapkan' }}</dd></div>
      <div><dt>Customer Code</dt><dd>{{ customer.customerCode || '—' }}</dd></div>
      <div><dt>Category</dt><dd>{{ customer.category || '—' }}</dd></div>
      <div><dt>Segment</dt><dd>{{ customer.segment || '—' }}</dd></div>
      <div><dt>Operating Status</dt><dd>{{ operating(operatingStatus) }}</dd></div>
      <div><dt>Converted</dt><dd>{{ date(customer.convertedAt) }}</dd></div>
      <div><dt>Last Updated</dt><dd>{{ date(customer.updatedAt) }}</dd></div>
      <div><dt>Status</dt><dd class="active">Customer · Active</dd></div>
    </dl>
  </section>
</template>

<style scoped>
.customer-snapshot-card{display:grid;gap:14px;padding:18px;border:1px solid #dbe8df;border-radius:16px;background:linear-gradient(135deg,#fbfefc,#fff);box-shadow:0 6px 20px rgba(26,91,57,.05)}header{display:flex;align-items:flex-start;justify-content:space-between;gap:12px}header p{margin:0;color:#218653;font-size:11px;font-weight:800;letter-spacing:.08em;text-transform:uppercase}h2{margin:3px 0 0;color:#173f2c;font-size:17px}header>span{padding:5px 8px;border-radius:999px;background:#eaf7ef;color:#176a40;font-size:11px;font-weight:800}dl{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:12px;margin:0}dl>div{display:grid;gap:3px;min-width:0}dt{color:#748078;font-size:10px;font-weight:800;letter-spacing:.04em;text-transform:uppercase}dd{margin:0;color:#27352d;font-size:13px;font-weight:700;overflow-wrap:anywhere}.active{color:#176a40}@media(max-width:700px){dl{grid-template-columns:repeat(2,minmax(0,1fr))}}@media(max-width:390px){dl{grid-template-columns:1fr}}
</style>
