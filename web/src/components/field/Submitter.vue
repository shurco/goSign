<template>
  <div v-if="mobileView" @mouseenter="renderDropdown = true" @touchstart="renderDropdown = true">
    <div class="flex space-x-2 items-end">
      <div class="group/contenteditable-container bg-[#faf7f5] rounded-md p-2 border border-[#e7e2df] w-full flex justify-between items-end">
        <div class="flex items-center space-x-2">
          <span class="w-3 h-3 flex-shrink-0 rounded-full" :class="subColors[submitters.indexOf(selectedSubmitter)]" />
          <Contenteditable v-model="selectedSubmitter.name" class="cursor-text" :icon-inline="true" :editable="editable" :select-on-edit-click="true" :icon-width="18"
            @update:model-value="$emit('name-change', selectedSubmitter)" />
        </div>
      </div>
      <div class="dropdown dropdown-top dropdown-end">
        <label tabindex="0" class="bg-[#faf7f5] cursor-pointer rounded-md p-2 border border-[#e7e2df] w-full flex justify-center">
          <SvgIcon name="chevron-up" width="24" height="24" />
        </label>
        <ul v-if="editable && renderDropdown" tabindex="0" class="rounded-md min-w-max mb-2" :class="menuClasses" @click="closeDropdown">
          <li v-for="(submitter, index) in submitters" :key="submitter.id">
            <a href="#" class="flex px-2 group justify-between items-center" :class="{ active: submitter === selectedSubmitter }" @click.prevent="selectSubmitter(submitter)">
              <span class="py-1 flex items-center">
                <span class="rounded-full w-3 h-3 ml-1 mr-3" :class="subColors[index]" />
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
    <label v-if="compact" tabindex="0" :title="selectedSubmitter.name" class="cursor-pointer text-[#faf7f5] flex h-full items-center justify-center">
      <button class="mx-1 w-3 h-3 rounded-full" :class="subColors[submitters.indexOf(selectedSubmitter)]" />
    </label>
    <label v-else ref="label" tabindex="0"
      class="group cursor-pointer group/contenteditable-container rounded-md p-2 border border-[#e7e2df] hover:border-content w-full flex justify-between">
      <div class="flex items-center space-x-2">
        <span class="w-3 h-3 rounded-full" :class="subColors[submitters.indexOf(selectedSubmitter)]" />
        <Contenteditable v-model="selectedSubmitter.name" class="cursor-text" :icon-inline="true" :editable="editable" :select-on-edit-click="true" :icon-width="18"
          @update:model-value="$emit('name-change', selectedSubmitter)" />
      </div>
      <span class="flex items-center transition-all duration-75 group-hover:border border-[#291334]/20 border-dashed w-6 h-6 justify-center rounded">
        <SvgIcon name="plus" width="18" height="18" />
      </span>
    </label>
    <ul v-if="(editable || !compact) && renderDropdown" tabindex="0" :class="menuClasses" @click="closeDropdown">
      <li v-for="(submitter, index) in submitters" :key="submitter.id">
        <a href="#" class="flex px-2 group justify-between items-center" :class="{ active: submitter === selectedSubmitter }" @click.prevent="selectSubmitter(submitter)">
          <span class="py-1 flex items-center">
            <span class="rounded-full w-3 h-3 ml-1 mr-3" :class="subColors[index]" />
            <span>
              {{ submitter.name }}
            </span>
          </span>
          <div v-if="!compact && submitters.length > 1 && editable" class="flex">
            <div class="flex-col pr-1 hidden group-hover:flex -mt-1 h-0">
              <button title="Up" class="relative w-2" style="font-size: 10px; margin-bottom: -4px" @click.stop="[move(submitter, -1), $refs.label.focus()]">
                ▲
              </button>
              <button title="Down" class="relative w-2" style="font-size: 10px; margin-top: -4px" @click.stop="[move(submitter, 1), $refs.label.focus()]">
                ▼
              </button>
            </div>
            <button v-if="!compact && submitters.length > 1 && editable" class="hidden group-hover:block px-2" @click.stop="remove(submitter)">
              <SvgIcon name="trash-x" width="18" height="18" />
            </button>
          </div>
        </a>
      </li>
      <li v-if="submitters.length < 10 && editable">
        <a href="#" class="flex px-2" @click.prevent="addSubmitter">
          <SvgIcon name="user-plus" class="mr-2" width="20" height="20" stroke-width="1.6" />
          <span class="py-1"> Add {{ subNames[submitters.length] }} </span>
        </a>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { ref, inject, computed } from 'vue';
import Contenteditable from "@/components/field/Contenteditable.vue";
import { subColors, subNames } from "@/components/field/constants.ts";
import { v4 } from "uuid";

const props = defineProps({
  submitters: {
    type: Array,
    required: true,
  },
  editable: {
    type: Boolean,
    required: false,
    default: true,
  },
  compact: {
    type: Boolean,
    required: false,
    default: false,
  },
  mobileView: {
    type: Boolean,
    required: false,
    default: false,
  },
  modelValue: {
    type: String,
    required: true,
  },
  menuClasses: {
    type: String,
    required: false,
    default: "dropdown-content menu p-2 shadow bg-[#faf7f5] rounded-box w-full z-10",
  },
})

const emit = defineEmits(["update:model-value", "remove", "new-submitter", "name-change"]);

const save = inject('save');

const renderDropdown: any = ref(false);

const selectedSubmitter: any = computed(() => {
  return props.submitters.find((e: any) => e.id === props.modelValue)
})

function selectSubmitter(submitter: any) {
  emit("update:model-value", submitter.id);
}

function remove(submitter: any) {
  if (window.confirm('Are you sure?')) {
    emit('remove', submitter)
  }
}

function move(submitter: any, direction: any) {
  const currentIndex = props.submitters.indexOf(submitter)
  props.submitters.splice(currentIndex, 1)

  if (currentIndex + direction > props.submitters.length) {
    props.submitters.unshift(submitter)
  } else if (currentIndex + direction < 0) {
    props.submitters.push(submitter)
  } else {
    props.submitters.splice(currentIndex + direction, 0, submitter)
  }

  selectSubmitter(submitter)
  save
}

function addSubmitter() {
  const newSubmitter = {
    name: subNames[props.submitters.length],
    id: v4(),
  }

  props.submitters.push(newSubmitter);
  emit("update:model-value", newSubmitter.id);
  emit("new-submitter", newSubmitter);
}

function closeDropdown() {
  const activeElement = document.activeElement as HTMLElement;
  if (activeElement !== null) {
    activeElement.blur();
  }
}
</script>