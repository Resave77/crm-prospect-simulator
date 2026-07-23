<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { getMyCustomer } from '../../../api/crm'
import type { CustomerDetail } from '../../../types/crm'

const route = useRoute(); const detail = ref<CustomerDetail | null>(null); const error = ref(''); const loading = ref(true)
const mapElement = ref<HTMLElement | null>(null); let map: L.Map | null = null
function renderMap() { const item = detail.value?.customer; if (!mapElement.value || item?.address.latitude == null || item.address.longitude == null) return; map?.remove(); map = L.map(mapElement.value).setView([item.address.latitude, item.address.longitude], 16); L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', { attribution: '&copy; OpenStreetMap contributors' }).addTo(map); L.marker([item.address.latitude, item.address.longitude]).addTo(map).bindPopup(item.name).openPopup() }
onMounted(async () => { try { detail.value = await getMyCustomer(String(route.params.id)); await nextTick(); renderMap() } catch (caught) { error.value = (caught as { response?: { data?: { error?: { message?: string } } } }).response?.data?.error?.message ?? 'Unable to load customer.' } finally { loading.value = false } })
onBeforeUnmount(() => map?.remove())
</script>

<template><section class="mobile-page"><RouterLink class="back-link" to="/sales/my-customers"><i class="pi pi-arrow-left" /> My Customer</RouterLink><Message v-if="error" severity="error">{{ error }}</Message><div v-if="loading" class="empty-state"><i class="pi pi-spin pi-spinner" /></div><template v-else-if="detail"><div class="mobile-title"><div><p class="eyebrow">Customer Existing</p><h1>{{ detail.customer.name }}</h1></div><Tag value="ACTIVE" severity="success" /></div><div class="mobile-section customer-code-grid"><div><span>Customer code</span><strong>{{ detail.customer.customerCode }}</strong></div><div><span>Parent code</span><strong>{{ detail.customer.parentCode }}</strong></div></div><div class="mobile-section snapshot-detail"><h2>Site and company</h2><p><i class="pi pi-building" /> {{ detail.customer.parentCompanyName }}</p><p><i class="pi pi-map-marker" /> {{ detail.customer.address.previewAddress }}</p><p><i class="pi pi-user" /> {{ detail.customer.salesExecutiveName }}</p><div class="tag-row"><Tag :value="detail.customer.segment" /><Tag :value="detail.customer.category" severity="secondary" /></div><div v-if="detail.customer.address.latitude != null && detail.customer.address.longitude != null" ref="mapElement" class="snapshot-map" aria-label="Saved customer location map" /><Message v-else severity="warn" :closable="false">No saved coordinates are available for this customer.</Message></div><div class="mobile-section"><h2>Conversion source</h2><p><strong>Prospect:</strong> {{ detail.sourceProspectName }}</p><p><strong>Source ID:</strong> {{ detail.customer.sourceProspectId }}</p><p><strong>Converted:</strong> {{ new Date(detail.customer.convertedAt).toLocaleString() }}</p></div><div class="mobile-section"><h2>Site contacts</h2><div v-if="detail.customer.contacts.length" class="contact-list"><div v-for="contact in detail.customer.contacts" :key="`${contact.name}-${contact.phone}`"><strong>{{ contact.name }}</strong><span>{{ contact.position }}</span><a v-if="contact.phone" :href="`tel:${contact.phone}`">{{ contact.phone }}</a><a v-if="contact.email" :href="`mailto:${contact.email}`">{{ contact.email }}</a></div></div><p v-else class="muted">No contacts saved.</p></div><div class="mobile-section"><h2>Parent company contacts</h2><div v-if="detail.parentCompany.contacts.length" class="contact-list"><div v-for="contact in detail.parentCompany.contacts" :key="`${contact.name}-${contact.email}`"><strong>{{ contact.name }}</strong><span>{{ contact.position }}</span><a v-if="contact.phone" :href="`tel:${contact.phone}`">{{ contact.phone }}</a><a v-if="contact.email" :href="`mailto:${contact.email}`">{{ contact.email }}</a></div></div><p v-else class="muted">No parent company contacts saved.</p></div><Message severity="info" :closable="false">Existing Customer attendance is deferred to a future scope.</Message></template></section></template>

<style scoped>
.customer-code-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.65rem; }
.customer-code-grid > div { padding: 0.65rem; display: grid; gap: 0.2rem; background: var(--brand-blue-50); border-radius: var(--radius-sm); }
.customer-code-grid span { color: var(--text-muted); font-size: 0.52rem; text-transform: uppercase; }
.customer-code-grid strong { font-size: 0.72rem; }
.snapshot-detail a { color: var(--brand-blue); text-decoration: none; }
.snapshot-map { height: 245px; margin-top: 0.7rem; overflow: hidden; border: 1px solid var(--border-light); border-radius: var(--radius-md); }
.contact-list { display: grid; gap: 0.5rem; }
.contact-list > div { padding: 0.6rem; display: grid; gap: 0.1rem; background: var(--surface-subtle); border-radius: var(--radius-sm); font-size: 0.6rem; }
.contact-list span { color: var(--text-muted); }
.contact-list a { color: var(--brand-blue); text-decoration: none; }
</style>
