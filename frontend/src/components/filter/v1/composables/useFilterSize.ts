import { computed } from "vue";

export type FilterSize = "small" | "medium" | "large";

export interface UseFilterSizeOptions {
  size?: FilterSize | null;
}

export const useFilterSize = (options: UseFilterSizeOptions = {}) => {
  const actualSize = computed<FilterSize>(() => {
    if (options.size !== undefined && options.size !== null) {
      return options.size;
    }
    return "small";
  });

  return { actualSize };
};

export const filterSizeClasses = {
  large: {
    icon: "text-[14px]",
    input: "text-[14px]",
    wrapper: "h-[38px] text-[14px]",
  },
  medium: {
    icon: "text-[13px]",
    input: "text-[13px]",
    wrapper: "h-[38px] text-[13px]",
  },
  small: {
    icon: "text-[12px]",
    input: "text-[12px]",
    wrapper: "h-[33px] text-[12px]",
  },
};
