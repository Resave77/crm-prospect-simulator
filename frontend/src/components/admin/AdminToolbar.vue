<template><div class="admin-toolbar"><div class="admin-toolbar__main"><slot name="search" /><slot name="filters" /></div><div class="admin-toolbar__actions"><slot name="actions" /></div></div></template>
<style scoped>
.admin-toolbar{display:flex;align-items:center;justify-content:space-between;gap:.75rem;flex-wrap:wrap;padding:.75rem;border:1px solid var(--border-light);border-radius:var(--radius-lg);background:var(--surface-card)}
.admin-toolbar__main,.admin-toolbar__actions{display:flex;align-items:center;gap:.5rem;flex-wrap:wrap}.admin-toolbar__main{flex:1;min-width:0}
@media(max-width:640px){.admin-toolbar__main,.admin-toolbar__actions{width:100%}.admin-toolbar__actions :deep(button){flex:1}}
</style>
