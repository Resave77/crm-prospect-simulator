<script setup lang="ts">
withDefaults(
  defineProps<{
    badgeLabel?: string;
    disabled?: boolean;
    error?: string | null;
    helper?: string;
    label?: string;
    maxlength?: number;
    modelValue?: string | null | undefined;
    placeholder?: string;
    readonly?: boolean;
    required?: boolean;
    type?: string;
  }>(),
  {
    badgeLabel: "",
    disabled: false,
    error: null,
    helper: "",
    maxlength: undefined,
    modelValue: "",
    placeholder: "",
    readonly: false,
    required: false,
    type: "text",
  }
);

const emit = defineEmits<{
  (event: "update:modelValue", value: string): void;
  (event: "input"): void;
}>();

const onInput = (event: Event) => {
  emit("update:modelValue", (event.target as HTMLInputElement).value);
  emit("input");
};
</script>

<template>
  <label class="flex flex-col gap-[8px]">
    <span
      v-if="label"
      class="flex flex-wrap items-center gap-[6px] text-[12px] font-semibold text-[#334155]"
    >
      <span>{{ label }}</span>
      <span v-if="required" class="text-[#dc2626]">*</span>
      <span
        v-if="badgeLabel"
        class="rounded-full border border-[#e2e8f0] bg-[#f8fafc] px-[8px] py-[2px] text-[11px] font-semibold text-[#64748b]"
      >
        {{ badgeLabel }}
      </span>
    </span>

    <input
      :value="modelValue ?? ''"
      :type="type"
      :maxlength="maxlength"
      :placeholder="placeholder"
      :disabled="disabled"
      :readonly="readonly"
      class="h-[48px] min-w-0 overflow-hidden text-ellipsis whitespace-nowrap rounded-[14px] border bg-white px-[14px] text-[12px] font-medium text-[#0f172a] shadow-[0px_1px_2px_0px_rgba(15,23,42,0.04)] outline-none transition-all placeholder:text-[#94a3b8] focus:ring-2 disabled:cursor-not-allowed disabled:bg-[#f1f5f9] disabled:text-[#64748b]"
      :class="
        error
          ? 'border-red-300 focus:border-red-400 focus:ring-red-100'
          : 'border-[#d9e2ec] focus:border-[#94a3b8] focus:ring-[#dbeafe]'
      "
      @input="onInput"
    />

    <span v-if="error" class="text-[11px] text-red-600">
      {{ error }}
    </span>
    <span v-else-if="helper" class="text-[11px] text-[#64748b]">
      {{ helper }}
    </span>
    <slot v-else />
  </label>
</template>
