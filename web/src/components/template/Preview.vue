<template>
  <div>
    <div class="relative">
      <img
        :src="`${previewImage.url}/p/${previewImage.filename}`"
        :width="previewImage.metadata.width"
        :height="previewImage.metadata.height"
        class="rounded border border-[#e7e2df]"
        loading="lazy"
      />
      <div
        class="group absolute top-0 right-0 bottom-0 left-0 flex cursor-pointer justify-end p-1"
        @click="$emit('scroll-to', item)"
      >
        <div v-if="editable" class="flex w-full justify-between">
          <div style="width: 26px" />
          <div class="flex flex-col justify-between opacity-0 group-hover:opacity-100">
            <div>
              <button
                class="btn btn-xs border-base-200 hover:border-base-content hover:bg-base-content hover:text-base-100 w-full rounded bg-white text-[--color-base-content] transition-colors"
                style="width: 24px; height: 24px"
                @click.stop="$emit('remove', item)"
              >
                &times;
              </button>
            </div>
            <div v-if="withArrows" class="flex flex-col space-y-1">
              <button
                class="btn btn-xs border-base-200 hover:border-base-content hover:bg-base-content hover:text-base-100 w-full rounded bg-white text-[--color-base-content] transition-colors"
                style="width: 24px; height: 24px"
                @click.stop="$emit('up', item)"
              >
                &uarr;
              </button>
              <button
                class="btn btn-xs border-base-200 hover:border-base-content hover:bg-base-content hover:text-base-100 w-full rounded bg-white text-[--color-base-content] transition-colors"
                style="width: 24px; height: 24px"
                @click.stop="$emit('down', item)"
              >
                &darr;
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="flex pt-1.5 pb-2">
      <Contenteditable
        :model-value="item.name"
        :icon-width="16"
        :editable="editable"
        style="max-width: 95%"
        class="mx-auto"
        @update:model-value="onUpdateName"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import Contenteditable from "@/components/field/Contenteditable.vue";

const props = defineProps({
  item: {
    type: Object,
    required: true
  },
  template: {
    type: Object,
    required: true
  },
  document: {
    type: Object,
    required: true
  },
  editable: {
    type: Boolean,
    required: false,
    default: true
  },
  acceptFileTypes: {
    type: String,
    required: false,
    default: "image/*, application/pdf"
  },
  withArrows: {
    type: Boolean,
    required: false,
    default: true
  }
});

const emit = defineEmits(["scroll-to", "change", "remove", "up", "down", "replace"]);

const previewImage = computed(() => {
  return [...props.document.preview_images].sort((a, b) => parseInt(a.filename, 10) - parseInt(b.filename, 10))[0];
});

const onUpdateName = (value: any) => {
  props.item.name = value;
  emit("change");
};
</script>
