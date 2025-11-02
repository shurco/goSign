<template>
  <div>
    <Page
      v-for="(image, index) in sortedPreviewImages"
      :key="image.id"
      :ref="setPageRefs"
      :number="index"
      :editable="editable"
      :areas="areasIndex[index]"
      :allow-draw="allowDraw"
      :is-drag="isDrag"
      :default-fields="defaultFields"
      :draw-field="drawField"
      :selected-submitter="selectedSubmitter"
      :image="image"
      @drop-field="handleDropField"
      @remove-area="handleRemoveArea"
      @draw="handleDraw"
      @select-submitter="emit('select-submitter', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUpdate, ref } from "vue";
import type { Documents, Submitters } from "@/models/index";
import type { Area, Field } from "@/models/template";
import Page from "@/components/template/Page.vue";

interface Props {
  document: Documents;
  areasIndex?: Record<number, any[]>;
  defaultFields?: Field[];
  allowDraw?: boolean;
  selectedSubmitter: Submitters;
  editable?: boolean;
  drawField?: Field | null;
  isDrag?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  areasIndex: () => ({}),
  defaultFields: () => [],
  allowDraw: true,
  editable: true,
  drawField: null,
  isDrag: false
});

interface Emits {
  "drop-field": [event: any];
  "remove-area": [area: Area];
  draw: [event: any];
  "select-submitter": [submitterId: string];
}

const emit = defineEmits<Emits>();

const pageRefs = ref<any[]>([]);

const numberOfPages = computed(() => {
  if (!props.document) {
    return 0;
  }
  return props.document.metadata?.pdf?.number_of_pages || props.document.preview_images.length;
});

const previewImagesIndex = computed(() => {
  if (!props.document) {
    return {};
  }
  return props.document.preview_images.reduce(
    (acc, e) => {
      e.url = `${props.document.url}/${props.document.id}`;
      acc[parseInt(e.filename, 10)] = e;
      return acc;
    },
    {} as Record<number, any>
  );
});

const sortedPreviewImages = computed(() => {
  if (!props.document || !props.document.preview_images.length) {
    return [];
  }
  const lazyloadMetadata = props.document.preview_images[props.document.preview_images.length - 1].metadata;
  return [...Array(numberOfPages.value).keys()].map((i) => {
    return (
      previewImagesIndex.value[i] || {
        metadata: lazyloadMetadata,
        id: Math.random().toString(),
        url: `${props.document.url}/${props.document.id}`,
        filename: props.document.preview_images[i].filename
      }
    );
  });
});

onBeforeUpdate(() => {
  pageRefs.value = [];
});

function scrollToArea(area: Area): void {
  const pageRef = pageRefs.value[area.page];
  if (!pageRef || !pageRef.areaRefs) {
    return;
  }

  const areaRef = pageRef.areaRefs.find((e: any) => e.area === area);
  if (!areaRef || !areaRef.$el) {
    return;
  }

  areaRef.$el.scrollIntoView({ behavior: "smooth", block: "center" });
}

function setPageRefs(el: any): void {
  if (el) {
    pageRefs.value.push(el);
  }
}

function handleDropField(event: any): void {
  if (!props.document) {
    return;
  }
  emit("drop-field", { ...event, attachment_id: props.document.id });
}

function handleRemoveArea(area: Area): void {
  emit("remove-area", area);
}

function handleDraw(event: any): void {
  if (!props.document) {
    return;
  }
  emit("draw", { ...event, attachment_id: props.document.id });
}

defineExpose({
  scrollToArea,
  get document() {
    return props.document;
  },
  pageRefs
});
</script>
