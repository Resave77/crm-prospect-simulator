<script setup lang="ts">
import { useRouter } from 'vue-router'
import Button from 'primevue/button'

const router = useRouter()
</script>

<template>
  <section class="create-page">
    <header class="topbar">
      <Button label="Back to Customer Site" icon="pi pi-arrow-left" text class="back-button" @click="router.push('/admin/customers')" />
      <div class="topbar-title">
        <strong>Create New Customer</strong>
        <span>Customer Management &gt; Create</span>
      </div>
      <Button label="Cancel" severity="secondary" outlined size="small" @click="router.push('/admin/customers')" />
    </header>

    <main class="content">
      <div class="intro">
        <span class="eyebrow">Customer Management</span>
        <h1>Create New Customer</h1>
        <p>Choose the customer record you want to create. Customer sites can be linked to a parent company.</p>
      </div>

      <section class="choice-card">
        <div class="choice-header">
          <div>
            <h2>What would you like to create?</h2>
            <p>Select one option to continue with the appropriate ERP form.</p>
          </div>
          <span class="scope-badge">Customer Management</span>
        </div>

        <div class="choices">
          <button type="button" class="entity-card site-card" @click="router.push('/admin/customers/add/site')">
            <span class="entity-icon"><i class="pi pi-map-marker" /></span>
            <span class="entity-copy">
              <strong>Customer Site</strong>
              <small>Create an outlet, branch, store, or other operational customer location.</small>
              <span class="entity-link">Create Customer Site <i class="pi pi-arrow-right" /></span>
            </span>
          </button>
          <button type="button" class="entity-card company-card" @click="router.push('/admin/customers/add/company')">
            <span class="entity-icon"><i class="pi pi-building" /></span>
            <span class="entity-copy">
              <strong>Parent Company</strong>
              <small>Create the corporate entity that owns or groups one or more customer sites.</small>
              <span class="entity-link">Create Parent Company <i class="pi pi-arrow-right" /></span>
            </span>
          </button>
        </div>
      </section>
    </main>
  </section>
</template>

<style scoped>
.create-page{min-height:100vh;background:#f7f9fc;color:#172033}.topbar{display:flex;align-items:center;gap:14px;min-height:58px;padding:0 32px;border-bottom:1px solid #dfe6ef;background:#fff}.back-button{padding:0 14px 0 0;border-right:1px solid #dfe6ef;border-radius:0;color:#536681;font-size:12px}.topbar-title{display:grid;gap:2px;flex:1}.topbar-title strong{font-size:17px}.topbar-title span{color:#8191a9;font-size:10px;letter-spacing:.04em}.content{max-width:1130px;margin:0 auto;padding:38px 32px}.intro{margin-bottom:22px}.eyebrow{color:#e63946;font-size:11px;font-weight:800;letter-spacing:.12em;text-transform:uppercase}.intro h1{margin:7px 0 6px;font-size:28px;letter-spacing:-.03em}.intro p,.choice-header p{margin:0;color:#60728e;font-size:13px}.choice-card{border:1px solid #dce5f0;border-radius:18px;background:#fff;box-shadow:0 12px 32px rgba(28,50,84,.06)}.choice-header{display:flex;align-items:flex-start;justify-content:space-between;gap:16px;padding:24px 26px;border-bottom:1px solid #edf1f6}.choice-header h2{margin:0 0 5px;font-size:17px}.scope-badge{padding:6px 10px;border:1px solid #f4b3ba;border-radius:999px;background:#fff5f5;color:#d62839;font-size:10px;font-weight:700;white-space:nowrap}.choices{display:grid;grid-template-columns:1fr 1fr;gap:14px;padding:26px}.entity-card{display:flex;align-items:flex-start;gap:16px;padding:22px;text-align:left;border:1px solid #dce5f0;border-radius:14px;background:#fff;cursor:pointer;transition:.18s}.entity-card:hover{transform:translateY(-2px);box-shadow:0 8px 22px rgba(28,50,84,.1)}.site-card:hover{border-color:#8bbcff;background:#f8fbff}.company-card:hover{border-color:#f3a1aa;background:#fffafa}.entity-icon{display:grid;place-items:center;width:48px;height:48px;flex:none;border-radius:13px;background:#eff6ff;color:#2563eb;font-size:19px}.company-card .entity-icon{background:#fff0f1;color:#e63946}.entity-copy{display:grid;gap:7px}.entity-copy strong{font-size:16px}.entity-copy small{min-height:38px;color:#667995;font-size:12px;line-height:1.55}.entity-link{color:#2563eb;font-size:12px;font-weight:800}.company-card .entity-link{color:#d62839}.entity-link i{margin-left:4px;font-size:10px}@media(max-width:700px){.topbar{padding:0 16px}.content{padding:26px 16px}.choices{grid-template-columns:1fr;padding:16px}.choice-header{padding:20px;flex-direction:column}.topbar-title span{display:none}}
</style>
