import type { InjectionKey, Ref } from "vue";

import type { FilterBarChipGroupV1 } from "@/components/filter/v1/AppFilterBarChipsV1.vue";

export type AppPageHeaderV4FilterRegistration = {
  chipGroups: Ref<FilterBarChipGroupV1[]>;
  onClearAll: () => void;
  onRemove: (payload: { key: string; value: number | string }) => void;
};

export type AppPageHeaderV4FilterContext = {
  register: (registration: AppPageHeaderV4FilterRegistration) => () => void;
};

export const appPageHeaderV4FilterContextKey: InjectionKey<AppPageHeaderV4FilterContext> =
  Symbol("app-page-header-v4-filter-context");
