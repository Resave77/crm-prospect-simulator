<template><div class="admin-pagination-strip"><span class="admin-pagination-strip__meta"><slot name="meta" /></span><div class="admin-pagination-strip__controls"><slot /></div></div></template>
<style scoped>
.admin-pagination-strip{display:flex;align-items:center;justify-content:space-between;gap:.75rem;min-height:2.5rem;color:var(--text-muted);font-size:.78rem}.admin-pagination-strip__controls{display:flex;align-items:center;gap:.35rem}@media(max-width:480px){.admin-pagination-strip{align-items:flex-start;flex-direction:column}.admin-pagination-strip__controls{width:100%}}
</style>
