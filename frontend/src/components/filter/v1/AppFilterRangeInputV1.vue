<script setup lang="ts">
import { computed } from "vue";

import {
  filterSizeClasses,
  useFilterSize,
  type FilterSize,
} from "./composables/useFilterSize";

interface Props {
  disabled?: boolean;
  inputClass?: string;
  label?: string;
  max?: number;
  min?: number;
  modelValue: number | string | null | undefined;
  placeholder?: string;
  size?: FilterSize | null;
  step?: number;
  suffix?: string;
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  inputClass: "",
  label: "Range",
  max: 120,
  min: 0,
  placeholder: "Custom max value",
  size: null,
  step: 1,
  suffix: "Days",
});

const emit = defineEmits<{
  (e: "update:modelValue", value: string): void;
}>();

const { actualSize } = useFilterSize({ size: props.size });

const numericValue = computed(() => {
  if (props.modelValue === "" || props.modelValue == null) return null;

  const value = Number(props.modelValue);
  if (Number.isNaN(value)) return null;

  return Math.min(Math.max(value, props.min), props.max);
});

const rangeValue = computed(() => numericValue.value ?? props.max);

const rangeProgress = computed(() => {
  const range = props.max - props.min;
  if (range <= 0) return 0;

  return ((rangeValue.value - props.min) / range) * 100;
});

const rangeStyle = computed(() => ({
  "--filter-range-progress": `${rangeProgress.value}%`,
}));

const summaryText = computed(
  () => `Max ${rangeValue.value} ${props.suffix.toLowerCase()}`
);

const numberInputValue = computed(() =>
  props.modelValue === "" || props.modelValue == null
    ? ""
    : String(props.modelValue)
);

const emitValue = (value: number | string) => {
  if (value === "") {
    emit("update:modelValue", "");
    return;
  }

  const parsedValue = Number(value);
  if (Number.isNaN(parsedValue)) {
    emit("update:modelValue", "");
    return;
  }

  const clampedValue = Math.min(Math.max(parsedValue, props.min), props.max);
  emit("update:modelValue", String(clampedValue));
};

const onRangeInput = (event: Event) => {
  emitValue((event.target as HTMLInputElement).value);
};

const onNumberInput = (event: Event) => {
  emitValue((event.target as HTMLInputElement).value);
};
</script>

<template>
  <div class="space-y-[8px]" :class="inputClass">
    <span class="text-[12px] font-semibold text-[#475569]">
      {{ label }}
    </span>

    <div
      class="rounded-[16px] border border-[#e2e8f0] bg-[#f8fafc] p-[14px] transition"
      :class="{ 'pointer-events-none opacity-60': disabled }"
    >
      <div class="flex items-center justify-between gap-[12px]">
        <span class="text-[12px] font-medium text-[#475569]">
          {{ summaryText }}
        </span>
      </div>

      <input
        type="range"
        :min="min"
        :max="max"
        :step="step"
        :value="rangeValue"
        :disabled="disabled"
        class="filter-range-input mt-[12px] h-[6px] w-full cursor-pointer accent-[#dc2626] disabled:cursor-not-allowed"
        :style="rangeStyle"
        @input="onRangeInput"
      />

      <div class="mt-[12px] flex items-center gap-[10px]">
        <input
          type="number"
          :min="min"
          :max="max"
          :step="step"
          :placeholder="placeholder"
          :value="numberInputValue"
          :disabled="disabled"
          :class="[
            'h-[38px] w-full rounded-[10px] border border-[#dbe4f0] bg-white px-[12px] text-[#0f172a] transition focus:border-[#dc2626] focus:outline-none disabled:cursor-not-allowed disabled:bg-[#f1f5f9]',
            filterSizeClasses[actualSize].input,
          ]"
          @input="onNumberInput"
        />
        <span class="shrink-0 text-[12px] font-semibold text-[#64748b]">
          {{ suffix }}
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.filter-range-input {
  -webkit-appearance: none;
  appearance: none;
  border-radius: 999px;
  background: linear-gradient(
    to right,
    #dc2626 0%,
    #dc2626 var(--filter-range-progress),
    #e2e8f0 var(--filter-range-progress),
    #e2e8f0 100%
  );
  outline: none;
}

.filter-range-input::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 16px;
  height: 16px;
  border: 2px solid #ffffff;
  border-radius: 999px;
  background: #dc2626;
  box-shadow: 0 2px 6px rgba(220, 38, 38, 0.28);
}

.filter-range-input::-moz-range-thumb {
  width: 16px;
  height: 16px;
  border: 2px solid #ffffff;
  border-radius: 999px;
  background: #dc2626;
  box-shadow: 0 2px 6px rgba(220, 38, 38, 0.28);
}

.filter-range-input::-moz-range-track {
  height: 6px;
  border-radius: 999px;
  background: #e2e8f0;
}

.filter-range-input::-moz-range-progress {
  height: 6px;
  border-radius: 999px;
  background: #dc2626;
}
</style>
