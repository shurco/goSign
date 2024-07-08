<template>
  <div v-if="mobileView" @mouseenter="renderDropdown = true" @touchstart="renderDropdown = true">
    <div class="flex items-end space-x-2">
      <div
        class="group/contenteditable-container flex w-full items-end justify-between rounded-md border border-[#e7e2df] bg-[#faf7f5] p-2"
      >
        <div class="flex items-center space-x-2">
          <span class="h-3 w-3 flex-shrink-0 rounded-full" :class="subColors[submitters.indexOf(selectedSubmitter)]" />
          <Contenteditable
            v-model="selectedSubmitter.name"
            class="cursor-text"
            :icon-inline="true"
            :editable="editable"
            :select-on-edit-click="true"
            :icon-width="18"
            @update:model-value="$emit('name-change', selectedSubmitter)"
          />
        </div>
      </div>
      <div class="dropdown dropdown-top dropdown-end">
        <label
          tabindex="0"
          class="flex w-full cursor-pointer justify-center rounded-md border border-[#e7e2df] bg-[#faf7f5] p-2"
        >
          <SvgIcon name="chevron-up" width="24" height="24" />
        </label>
        <ul
          v-if="editable && renderDropdown"
          tabindex="0"
          class="mb-2 min-w-max rounded-md"
          :class="menuClasses"
          @click="closeDropdown()"
        >
          <li v-for="(submitter, index) in submitters" :key="submitter.id">
            <a
              href="#"
              class="group flex items-center justify-between px-2"
              :class="{ active: submitter === selectedSubmitter }"
              @click.prevent="selectSubmitter(submitter)"
            >
              <span class="flex items-center py-1">
                <span class="ml-1 mr-3 h-3 w-3 rounded-full" :class="subColors[index]" />
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
            <a href="#" class="flex px-2" @click.prevent="addSubmitter">
              <SvgIcon name="user-plus" width="20" height="20" stroke-width="1.6" />
              <span class="py-1"> Add {{ names[submitters.length] }} </span>
            </a>
          </li>
        </ul>
      </div>
    </div>
  </div>
  <div v-else class="dropdown" @mouseenter="renderDropdown = true" @touchstart="renderDropdown = true">
    <label
      v-if="compact"
      tabindex="0"
      :title="selectedSubmitter.name"
      class="flex h-full cursor-pointer items-center justify-center text-[#faf7f5]"
    >
      <button class="mx-1 h-3 w-3 rounded-full" :class="subColors[submitters.indexOf(selectedSubmitter)]" />
    </label>
    <label
      v-else
      ref="label"
      tabindex="0"
      class="group/contenteditable-container hover:border-content group flex w-full cursor-pointer justify-between rounded-md border border-[#e7e2df] p-2"
    >
      <div class="flex items-center space-x-2">
        <span class="h-3 w-3 rounded-full" :class="subColors[submitters.indexOf(selectedSubmitter)]" />
        <Contenteditable
          v-model="selectedSubmitter.name"
          class="cursor-text"
          :icon-inline="true"
          :editable="editable"
          :select-on-edit-click="true"
          :icon-width="18"
          @update:model-value="$emit('name-change', selectedSubmitter)"
        />
      </div>
      <span
        class="flex h-6 w-6 items-center justify-center rounded border-dashed border-[#291334]/20 transition-all duration-75 group-hover:border"
      >
        <SvgIcon name="plus" width="18" height="18" />
      </span>
    </label>
    <ul v-if="(editable || !compact) && renderDropdown" tabindex="0" :class="menuClasses" @click="closeDropdown()">
      <li v-for="(submitter, index) in submitters" :key="submitter.id">
        <a
          href="#"
          class="group flex items-center justify-between px-2"
          :class="{ active: submitter === selectedSubmitter }"
          @click.prevent="selectSubmitter(submitter)"
        >
          <span class="flex items-center py-1">
            <span class="ml-1 mr-3 h-3 w-3 rounded-full" :class="subColors[index]" />
            <span>
              {{ submitter.name }}
            </span>
          </span>
          <div v-if="!compact && submitters.length > 1 && editable" class="flex">
            <div class="-mt-1 hidden h-0 flex-col pr-1 group-hover:flex">
              <button
                title="Up"
                class="relative w-2"
                style="font-size: 10px; margin-bottom: -4px"
                @click.stop="[move(submitter, -1), $refs.label.focus()]"
              >
                ▲
              </button>
              <button
                title="Down"
                class="relative w-2"
                style="font-size: 10px; margin-top: -4px"
                @click.stop="[move(submitter, 1), $refs.label.focus()]"
              >
                ▼
              </button>
            </div>
            <button
              v-if="!compact && submitters.length > 1 && editable"
              class="hidden px-2 group-hover:block"
              @click.stop="remove(submitter)"
            >
              <SvgIcon name="trash-x" width="18" height="18" />
            </button>
          </div>
        </a>
      </li>
      <li v-if="submitters.length < 10 && editable">
        <a href="#" class="flex px-2" @click.prevent="addSubmitter()">
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

const renderDropdown: any = ref(false);

const selectedSubmitter: any = computed(() => {
  return props.submitters.find((e: any) => e.id === props.modelValue);
});

function selectSubmitter(submitter: any): any {
  emit("update:model-value", submitter.id);
}

function remove(submitter: any): any {
  if (window.confirm("Are you sure?")) {
    emit("remove", submitter);
  }
}

function move(submitter: any, direction: any): any {
  const currentIndex = props.submitters.indexOf(submitter);
  props.submitters.splice(currentIndex, 1);

  if (currentIndex + direction > props.submitters.length) {
    props.submitters.unshift(submitter);
  } else if (currentIndex + direction < 0) {
    props.submitters.push(submitter);
  } else {
    props.submitters.splice(currentIndex + direction, 0, submitter);
  }

  selectSubmitter(submitter);
  save;
}

function addSubmitter(): any {
  const newSubmitter = {
    name: subNames[props.submitters.length],
    id: v4()
  };

  props.submitters.push(newSubmitter);
  emit("update:model-value", newSubmitter.id);
  emit("new-submitter", newSubmitter);
}

function closeDropdown(): any {
  renderDropdown.value = false;
  const activeElement = document.activeElement as HTMLElement;
  if (activeElement !== null) {
    activeElement.blur();
  }
}
</script>
