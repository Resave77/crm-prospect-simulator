<script setup lang="ts"></script>
<template>
  <div class="flex flex-col items-center justify-center px-6 text-center">
    <div class="relative mb-6">
      <div
        class="absolute inset-0 bg-red-100 rounded-full scale-150 blur-xl opacity-50"
      ></div>
      <i
        class="pi pi-lock text-[#DD0033] text-7xl relative z-10 animate-bounce-slow"
      ></i>
    </div>

    <h1 class="text-2xl font-bold text-gray-900 mb-2">Akses Dibatasi</h1>
    <p class="text-gray-500 max-w-sm mb-8">
      Maaf, akun kamu tidak memiliki izin untuk mengakses halaman ini. Silakan
      hubungi Admin jika menurutmu ini adalah kesalahan.
    </p>

    <!-- <div class="flex flex-col sm:flex-row gap-3 w-full max-w-xs">
      <Button 
        label="Kembali ke Dashboard" 
        icon="pi pi-home"
        class="p-button-danger w-full !bg-[#DD0033] !border-none h-12"
        @click="router.push('/dashboard')"
      />
      <Button 
        label="Hubungi Support" 
        icon="pi pi-envelope"
        class="p-button-text w-full !text-[#DD0033] h-12"
        @click="contactSupport"
      />
    </div> -->
  </div>
</template>

<style scoped>
.animate-bounce-slow {
  animation: bounce 3s infinite;
}
@keyframes bounce {
  0%,
  100% {
    transform: translateY(-5%);
  }
  50% {
    transform: translateY(0);
  }
}
</style>
