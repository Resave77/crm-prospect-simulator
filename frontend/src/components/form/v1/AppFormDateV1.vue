<script setup lang="ts">
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
} from "vue";

const props = withDefaults(
  defineProps<{
    badgeLabel?: string;
    dateFormat?: string;
    disabled?: boolean;
    error?: string | null;
    helper?: string;
    label?: string;
    maxDate?: Date;
    minDate?: Date;
    modelValue: string | Date | null | undefined;
    placeholder?: string;
    readonly?: boolean;
    required?: boolean;
    showIcon?: boolean;
  }>(),
  {
    badgeLabel: "",
    dateFormat: "yy-mm-dd",
    disabled: false,
    error: null,
    helper: "",
    placeholder: "Select Date",
    readonly: false,
    required: false,
    showIcon: true,
  }
);

const emit = defineEmits<{
  (event: "update:modelValue", value: string | Date | null | undefined): void;
  (event: "change"): void;
}>();

const isOpen = ref(false);
// eslint-disable-next-line no-undef
const containerRef = ref<HTMLDivElement | null>(null);
// eslint-disable-next-line no-undef
const panelRef = ref<HTMLDivElement | null>(null);
const panelStyle = ref<Record<string, string>>({});

const monthNames = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December",
];

const normalizeDateString = (value: string): Date | null => {
  if (!value) return null;
  const cleanVal = value.split("T")[0] || "";
  const [year, month, day] = cleanVal.split("-").map(Number);
  if (!year || !month || !day) return null;
  return new Date(year, month - 1, day);
};

const formatDateLocal = (value: Date | null): string | null => {
  if (!value) return null;
  const year = value.getFullYear();
  const month = String(value.getMonth() + 1).padStart(2, "0");
  const day = String(value.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
};

const selectedDateObj = computed<Date | null>(() => {
  if (!props.modelValue) return null;
  if (typeof props.modelValue === "string") {
    return normalizeDateString(props.modelValue);
  }
  return props.modelValue;
});

const displayText = computed(() => {
  if (!selectedDateObj.value) return "";
  return formatDateLocal(selectedDateObj.value) || "";
});

const now = new Date();
const viewYear = ref(selectedDateObj.value?.getFullYear() ?? now.getFullYear());
const viewMonth = ref(selectedDateObj.value?.getMonth() ?? now.getMonth());

watch(
  selectedDateObj,
  (newDate) => {
    if (newDate) {
      viewYear.value = newDate.getFullYear();
      viewMonth.value = newDate.getMonth();
    }
  },
  { immediate: true }
);

const currentMonthYearLabel = computed(() => {
  const mName = monthNames[viewMonth.value] || "";
  return `${mName} ${viewYear.value}`;
});

const calendarDays = computed(() => {
  const year = viewYear.value;
  const month = viewMonth.value;

  const firstDayOfWeek = new Date(year, month, 1).getDay(); // 0 = Sunday
  const daysInMonth = new Date(year, month + 1, 0).getDate();
  const daysInPrevMonth = new Date(year, month, 0).getDate();

  const days: Array<{
    date: Date;
    dayNum: number;
    isCurrentMonth: boolean;
    key: string;
  }> = [];

  // Trailing days from previous month
  for (let i = firstDayOfWeek - 1; i >= 0; i -= 1) {
    const dayNum = daysInPrevMonth - i;
    const dateObj = new Date(year, month - 1, dayNum);
    days.push({
      date: dateObj,
      dayNum,
      isCurrentMonth: false,
      key: `prev-${dayNum}`,
    });
  }

  // Days of current month
  for (let d = 1; d <= daysInMonth; d += 1) {
    const dateObj = new Date(year, month, d);
    days.push({
      date: dateObj,
      dayNum: d,
      isCurrentMonth: true,
      key: `curr-${d}`,
    });
  }

  // Leading days of next month
  const targetCells = days.length > 35 ? 42 : 35;
  const needed = targetCells - days.length;
  for (let n = 1; n <= needed; n += 1) {
    const dateObj = new Date(year, month + 1, n);
    days.push({
      date: dateObj,
      dayNum: n,
      isCurrentMonth: false,
      key: `next-${n}`,
    });
  }

  return days;
});

const isSelectedDate = (dateObj: Date) => {
  if (!selectedDateObj.value) return false;
  return (
    selectedDateObj.value.getFullYear() === dateObj.getFullYear() &&
    selectedDateObj.value.getMonth() === dateObj.getMonth() &&
    selectedDateObj.value.getDate() === dateObj.getDate()
  );
};

const isDisabledDate = (dateObj: Date) => {
  if (props.minDate) {
    const min = new Date(
      props.minDate.getFullYear(),
      props.minDate.getMonth(),
      props.minDate.getDate()
    );
    if (dateObj < min) return true;
  }

  if (props.maxDate) {
    const max = new Date(
      props.maxDate.getFullYear(),
      props.maxDate.getMonth(),
      props.maxDate.getDate()
    );
    if (dateObj > max) return true;
  }

  return false;
};

const prevMonth = () => {
  if (viewMonth.value === 0) {
    viewMonth.value = 11;
    viewYear.value -= 1;
  } else {
    viewMonth.value -= 1;
  }
};

const nextMonth = () => {
  if (viewMonth.value === 11) {
    viewMonth.value = 0;
    viewYear.value += 1;
  } else {
    viewMonth.value += 1;
  }
};

const updatePosition = () => {
  if (!containerRef.value) return;
  const rect = containerRef.value.getBoundingClientRect();
  const panelWidth = 245;

  panelStyle.value = {
    left: `${Math.min(window.innerWidth - panelWidth - 8, Math.max(8, rect.left))}px`,
    position: "fixed",
    top: `${rect.bottom + 4}px`,
    width: `${panelWidth}px`,
    zIndex: "9999",
  };
};

const togglePicker = async () => {
  if (props.disabled || props.readonly) return;
  if (!isOpen.value) {
    updatePosition();
    isOpen.value = true;
    await nextTick();
  } else {
    isOpen.value = false;
  }
};

const selectDate = (dateObj: Date) => {
  if (isDisabledDate(dateObj)) return;

  if (typeof props.modelValue === "string") {
    const str = formatDateLocal(dateObj) || "";
    emit("update:modelValue", str);
  } else {
    emit("update:modelValue", dateObj);
  }
  emit("change");
  isOpen.value = false;
};

const handleClickOutside = (event: MouseEvent) => {
  // eslint-disable-next-line no-undef
  const target = event.target as Node;
  if (
    containerRef.value &&
    !containerRef.value.contains(target) &&
    panelRef.value &&
    !panelRef.value.contains(target)
  ) {
    isOpen.value = false;
  }
};

const handleScrollOrResize = () => {
  if (isOpen.value) {
    updatePosition();
  }
};

watch(isOpen, (open) => {
  if (open) {
    window.addEventListener("scroll", handleScrollOrResize, true);
    window.addEventListener("resize", handleScrollOrResize);
  } else {
    window.removeEventListener("scroll", handleScrollOrResize, true);
    window.removeEventListener("resize", handleScrollOrResize);
  }
});

onMounted(() => {
  window.addEventListener("click", handleClickOutside);
});

onBeforeUnmount(() => {
  window.removeEventListener("click", handleClickOutside);
  window.removeEventListener("scroll", handleScrollOrResize, true);
  window.removeEventListener("resize", handleScrollOrResize);
});
</script>

<template>
  <div class="flex flex-col gap-[8px]">
    <label
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
    </label>

    <div
      ref="containerRef"
      class="flex h-[48px] w-full min-w-0 cursor-pointer items-center justify-between rounded-[14px] border bg-white px-3.5 shadow-[0_1px_2px_0_rgba(15,23,42,0.04)] transition-all focus-within:border-[#94a3b8] focus-within:ring-2 focus-within:ring-[#dbeafe]"
      :class="[
        error ? 'border-red-300' : 'border-[#d9e2ec]',
        disabled || readonly
          ? 'pointer-events-none bg-[#f1f5f9] opacity-60'
          : 'hover:border-[#cbd5e1]',
        isOpen ? 'border-[#94a3b8] ring-2 ring-[#dbeafe]' : '',
      ]"
      @click.stop="togglePicker"
    >
      <span
        class="truncate text-[12px] font-medium"
        :class="displayText ? 'text-[#0f172a]' : 'text-[#94a3b8]'"
      >
        {{ displayText || placeholder }}
      </span>

      <svg
        v-if="showIcon"
        xmlns="http://www.w3.org/2000/svg"
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        class="lucide lucide-calendar ml-2 size-4 shrink-0 text-[#64748b]"
      >
        <path d="M8 2v4" />
        <path d="M16 2v4" />
        <rect width="18" height="18" x="3" y="4" rx="2" />
        <path d="M3 10h18" />
      </svg>
    </div>

    <Teleport to="body">
      <div
        v-if="isOpen"
        ref="panelRef"
        :style="panelStyle"
        class="w-[245px] overflow-hidden rounded-[4px] border border-[#dbe3ef] bg-white p-[0px] shadow-[0_2px_8px_rgba(15,23,42,0.16)]"
      >
        <div
          class="flex h-[49px] items-center justify-between border-b border-[#e2e8f0] px-[16px]"
        >
          <button
            type="button"
            aria-label="Previous month"
            class="flex items-center justify-center p-1 text-[#52647a] transition-colors hover:text-[#0f172a]"
            @click.stop="prevMonth"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="lucide lucide-chevron-left size-[16px] text-[#52647a]"
            >
              <path d="m15 18-6-6 6-6" />
            </svg>
          </button>
          <span class="font-['Inter'] text-[12px] font-medium text-[#17243a]">
            {{ currentMonthYearLabel }}
          </span>
          <button
            type="button"
            aria-label="Next month"
            class="flex items-center justify-center p-1 text-[#52647a] transition-colors hover:text-[#0f172a]"
            @click.stop="nextMonth"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="lucide lucide-chevron-right size-[16px] text-[#52647a]"
            >
              <path d="m9 18 6-6-6-6" />
            </svg>
          </button>
        </div>

        <div
          class="grid grid-cols-7 px-[9px] pb-[9px] pt-[7px] text-center font-['Inter'] text-[12px] text-[#17243a]"
        >
          <span class="py-[5px]">Su</span>
          <span class="py-[5px]">Mo</span>
          <span class="py-[5px]">Tu</span>
          <span class="py-[5px]">We</span>
          <span class="py-[5px]">Th</span>
          <span class="py-[5px]">Fr</span>
          <span class="py-[5px]">Sa</span>

          <button
            v-for="item in calendarDays"
            :key="item.key"
            type="button"
            class="mx-auto flex size-[26px] items-center justify-center rounded-full text-[12px] transition-colors"
            :class="[
              isSelectedDate(item.date)
                ? 'bg-[#f04444] font-bold text-white'
                : item.isCurrentMonth
                  ? 'text-[#27354a] hover:bg-[#f1f5f9]'
                  : 'text-[#8a96a6] hover:bg-[#f8fafc]',
              isDisabledDate(item.date) ? 'cursor-not-allowed opacity-35' : '',
            ]"
            :disabled="isDisabledDate(item.date)"
            @click.stop="selectDate(item.date)"
          >
            {{ item.dayNum }}
          </button>
        </div>
      </div>
    </Teleport>

    <span v-if="error" class="text-[11px] text-red-600">
      {{ error }}
    </span>
    <span v-else-if="helper" class="text-[11px] text-[#64748b]">
      {{ helper }}
    </span>
    <slot v-else />
  </div>
</template>
