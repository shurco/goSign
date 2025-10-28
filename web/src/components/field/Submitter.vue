<template>
  <div v-if="mobileView" ref="mobileDropdownRef">
    <div class="flex items-end space-x-2">
      <div
        class="group/contenteditable-container flex w-full items-end justify-between rounded-md border border-[#e7e2df] bg-[#faf7f5] p-2"
      >
        <div class="flex items-center space-x-2">
          <span class="h-3 w-3 flex-shrink-0 rounded-full" :class="getSubmitterColor(selectedSubmitter)" />
          <Contenteditable
            v-model="selectedSubmitter.name"
            class="cursor-text"
            :icon-inline="true"
            :editable="editable"
            :select-on-edit-click="true"
            :icon-width="18"
            @update:model-value="$emit('name-change', selectedSubmitter)"
            @blur="save"
            @click.stop
          />
        </div>
      </div>
      <div class="dropdown dropdown-end dropdown-top">
        <label
          tabindex="0"
          class="flex w-full cursor-pointer justify-center rounded-md border border-[#e7e2df] bg-[#faf7f5] p-2"
          @click.stop="toggleDropdown()"
        >
          <SvgIcon name="chevron-up" width="24" height="24" />
        </label>
        <ul v-if="editable && renderDropdown" tabindex="0" class="mb-2 min-w-max rounded-md" :class="menuClasses">
          <li v-for="submitter in submitters" :key="submitter.id">
            <a
              href="#"
              class="group flex items-center justify-between px-2"
              :class="{ active: submitter === selectedSubmitter }"
              @click.prevent="
                selectSubmitter(submitter);
                closeDropdown();
              "
            >
              <span class="flex items-center py-1">
                <span class="mr-3 ml-1 h-3 w-3 rounded-full" :class="getSubmitterColor(submitter)" />
                <span>
                  {{ submitter.name }}
                </span>
              </span>
              <button v-if="submitters.length > 1 && editable" class="px-2" @click.stop="remove(submitter)">
                <SvgIcon name="trash-x" width="18" height="18" />
              </button>
            </a>
          </li>
          <li v-if="submitters.length < 10 && editable">
            <a
              href="#"
              class="flex px-2"
              @click.prevent="
                addSubmitter();
                closeDropdown();
              "
            >
              <SvgIcon name="user-plus" width="20" height="20" stroke-width="1.6" />
              <span class="py-1"> Add {{ names[submitters.length] }} </span>
            </a>
          </li>
        </ul>
      </div>
    </div>
  </div>
  <div v-else ref="dropdownRef" class="dropdown">
    <label
      v-if="compact"
      tabindex="0"
      :title="selectedSubmitter.name"
      class="flex h-full cursor-pointer items-center justify-center text-[#faf7f5]"
      @click.stop="toggleDropdown()"
    >
      <button class="mx-1 h-3 w-3 rounded-full" :class="getSubmitterColor(selectedSubmitter)" />
    </label>
    <label
      v-else
      ref="label"
      tabindex="0"
      class="group/contenteditable-container hover:border-content group flex w-full justify-between rounded-md border border-[#e7e2df] p-2"
    >
      <div class="flex items-center space-x-2">
        <span class="h-3 w-3 rounded-full" :class="getSubmitterColor(selectedSubmitter)" />
        <Contenteditable
          v-model="selectedSubmitter.name"
          class="cursor-text"
          :icon-inline="true"
          :editable="editable"
          :select-on-edit-click="true"
          :icon-width="18"
          @update:model-value="$emit('name-change', selectedSubmitter)"
          @blur="save"
          @click.stop
        />
      </div>
      <span
        class="flex h-6 w-6 cursor-pointer items-center justify-center rounded border-dashed border-[#291334]/20 transition-all duration-75 group-hover:border"
        @click.stop="toggleDropdown()"
      >
        <SvgIcon name="plus" width="18" height="18" />
      </span>
    </label>
    <ul v-if="(editable || !compact) && renderDropdown" tabindex="0" :class="menuClasses">
      <li v-for="submitter in submitters" :key="submitter.id">
        <a
          href="#"
          class="group flex items-center justify-between px-2"
          :class="{ active: submitter === selectedSubmitter }"
          @click.prevent="
            selectSubmitter(submitter);
            closeDropdown();
          "
        >
          <span class="flex items-center py-1">
            <span class="mr-3 ml-1 h-3 w-3 rounded-full" :class="getSubmitterColor(submitter)" />
            <span>
              {{ submitter.name }}
            </span>
          </span>
          <div v-if="!compact && submitters.length > 1 && editable" class="flex items-center gap-1">
            <div class="invisible flex flex-col gap-1 group-hover:visible group-[.active]:visible">
              <button
                title="Up"
                class="border-base-200 hover:border-base-content hover:bg-base-content hover:text-base-100 group-[.active]:text-base-content flex h-4 w-6 items-center justify-center rounded border bg-white text-[--color-base-content] transition-colors"
                @click.stop="move(submitter, -1)"
              >
                <SvgIcon name="chevron-up" width="12" height="12" />
              </button>
              <button
                title="Down"
                class="border-base-200 hover:border-base-content hover:bg-base-content hover:text-base-100 group-[.active]:text-base-content flex h-4 w-6 items-center justify-center rounded border bg-white text-[--color-base-content] transition-colors"
                @click.stop="move(submitter, 1)"
              >
                <SvgIcon name="chevron-down" width="12" height="12" />
              </button>
            </div>
            <button class="invisible px-2 group-hover:visible group-[.active]:visible" @click.stop="remove(submitter)">
              <SvgIcon name="trash-x" width="18" height="18" />
            </button>
          </div>
        </a>
      </li>
      <li v-if="submitters.length < 10 && editable">
        <a
          href="#"
          class="flex px-2"
          @click.prevent="
            addSubmitter();
            closeDropdown();
          "
        >
          <SvgIcon name="user-plus" class="mr-2" width="20" height="20" stroke-width="1.6" />
          <span class="py-1"> Add {{ subNames[submitters.length] }} </span>
        </a>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, ref } from "vue";
import Contenteditable from "@/components/field/Contenteditable.vue";
import { subColors, subNames } from "@/components/field/constants.ts";
import { useDropdown } from "@/composables/useDropdown";
import { v4 } from "uuid";

const props = defineProps({
  submitters: {
    type: Array,
    required: true
  },
  editable: {
    type: Boolean,
    required: false,
    default: true
  },
  compact: {
    type: Boolean,
    required: false,
    default: false
  },
  mobileView: {
    type: Boolean,
    required: false,
    default: false
  },
  modelValue: {
    type: String,
    required: true
  },
  menuClasses: {
    type: String,
    required: false,
    default: "dropdown-content menu p-2 shadow bg-[#faf7f5] rounded-box w-full z-10"
  }
});

const emit = defineEmits(["update:model-value", "remove", "new-submitter", "name-change"]);

const save = inject("save");

const dropdownRef = ref<HTMLElement | null>(null);
const mobileDropdownRef = ref<HTMLElement | null>(null);

const {
  isOpen: renderDropdown,
  close: closeDropdown,
  toggle: toggleDropdown
} = useDropdown(computed(() => (props.mobileView ? mobileDropdownRef.value : dropdownRef.value)));

const selectedSubmitter: any = computed(() => {
  return props.submitters.find((e: any) => e.id === props.modelValue);
});

// Store original colors by submitter ID to preserve colors when moving
const submitterColors = computed(() => {
  const colors: Record<string, string> = {};
  props.submitters.forEach((submitter: any, index: number) => {
    if (!submitter.colorIndex && submitter.colorIndex !== 0) {
      submitter.colorIndex = index;
    }
    colors[submitter.id] = subColors[submitter.colorIndex % subColors.length];
  });
  return colors;
});

function getSubmitterColor(submitter: any): string {
  return submitterColors.value[submitter.id];
}

function selectSubmitter(submitter: any): any {
  emit("update:model-value", submitter.id);
}

function remove(submitter: any): any {
  if (window.confirm("Are you sure?")) {
    emit("remove", submitter);
  }
}

function move(submitter: any, direction: any): void {
  const currentIndex = props.submitters.indexOf(submitter);

  // Check bounds
  const newIndex = currentIndex + direction;
  if (newIndex < 0 || newIndex >= props.submitters.length) {
    return;
  }

  // Remember current selection
  const wasSelected = submitter.id === props.modelValue;

  // Remove from current position
  props.submitters.splice(currentIndex, 1);

  // Insert at new position
  props.submitters.splice(newIndex, 0, submitter);

  // Restore selection if this submitter was selected
  if (wasSelected) {
    selectSubmitter(submitter);
  }

  if (save) {
    save();
  }
}

function addSubmitter(): any {
  const newSubmitter = {
    name: subNames[props.submitters.length],
    id: v4(),
    colorIndex: props.submitters.length
  };

  props.submitters.push(newSubmitter);
  emit("update:model-value", newSubmitter.id);
  emit("new-submitter", newSubmitter);
}
</script>
