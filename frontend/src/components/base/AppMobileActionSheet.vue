<script setup lang="ts">
import Button from "primevue/button";
import Dialog from "primevue/dialog";
import { computed } from "vue";

export type ActionSheetOption = {
  icon?: string;
  iconBgClass?: string;
  iconColorClass?: string;
  id: string;
  subtitle?: string;
  title: string;
};

const props = withDefaults(
  defineProps<{
    baseZIndex?: number;
    modelValue: boolean;
    options?: ActionSheetOption[];
    subtitle?: string;
    title?: string;
  }>(),
  {
    baseZIndex: 10050,
    options: () => [
      {
        icon: "pi-qrcode",
        iconBgClass: "bg-emerald-100",
        iconColorClass: "text-emerald-700",
        id: "scan",
        subtitle: "Pemindaian otomatis dengan pemotong sudut rapi",
        title: "Scan Dokumen (Presisi)",
      },
      {
        icon: "pi-camera",
        iconBgClass: "bg-blue-100",
        iconColorClass: "text-blue-700",
        id: "photo",
        subtitle: "Mengambil gambar standar menggunakan kamera HP",
        title: "Ambil Foto Biasa",
      },
      {
        icon: "pi-image",
        iconBgClass: "bg-purple-100",
        iconColorClass: "text-purple-700",
        id: "gallery",
        subtitle: "Mengunggah file gambar atau PDF dari galeri HP",
        title: "Pilih dari Galeri",
      },
    ],
    subtitle: "Silakan pilih cara untuk mengunggah berkas dokumen Anda",
    title: "Pilih Mode Pengambilan",
  }
);

const emit = defineEmits<{
  (e: "update:modelValue", value: boolean): void;
  (e: "select", id: any): void;
}>();

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit("update:modelValue", val),
});

const selectOption = (id: string) => {
  emit("select", id);
  visible.value = false;
};
</script>

<template>
  <Dialog
    v-model:visible="visible"
    :modal="true"
    :base-z-index="props.baseZIndex"
    position="bottom"
    class="w-[100vw] md:w-[480px] m-0 rounded-t-3xl overflow-hidden bg-white border-t border-slate-100 shadow-2xl"
    :show-header="false"
    :dismissable-mask="true"
  >
    <div class="px-5 py-6 flex flex-col gap-4">
      <div class="w-12 h-1.5 bg-slate-200 rounded-full mx-auto mb-2"></div>
      <h3 class="text-base font-bold text-slate-800 text-center">
        {{ title }}
      </h3>
      <p v-if="subtitle" class="text-xs text-slate-500 text-center -mt-2 mb-2">
        {{ subtitle }}
      </p>

      <div class="flex flex-col gap-2.5">
        <button
          v-for="opt in options"
          :key="opt.id"
          type="button"
          class="flex items-center gap-3.5 w-full p-4 rounded-xl border border-slate-100 bg-slate-50 text-left text-slate-700 hover:bg-slate-100 active:bg-slate-200 transition-colors font-semibold text-sm"
          @click="selectOption(opt.id)"
        >
          <div
            v-if="opt.icon"
            class="flex h-9 w-9 items-center justify-center rounded-lg shrink-0"
            :class="opt.iconBgClass || 'bg-slate-100'"
          >
            <span
              class="pi text-lg"
              :class="[opt.icon, opt.iconColorClass || 'text-slate-700']"
            />
          </div>
          <div>
            <div class="text-slate-800 text-sm font-bold">{{ opt.title }}</div>
            <div
              v-if="opt.subtitle"
              class="text-slate-500 text-xs font-normal mt-0.5"
            >
              {{ opt.subtitle }}
            </div>
          </div>
        </button>
      </div>

      <Button
        label="Batal"
        severity="danger"
        text
        class="w-full mt-2 py-3 font-semibold text-sm"
        @click="visible = false"
      />
    </div>
  </Dialog>
</template>
