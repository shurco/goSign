<template>
  <div class="relative cursor-crosshair select-none" :style="drawField ? 'touch-action: none' : ''">
    <img
      ref="image"
      loading="lazy"
      :src="`${image.url}/${image.filename}`"
      :width="width"
      :height="height"
      class="mb-4 rounded border"
      @load="onImageLoad"
    />
    <div class="absolute bottom-0 left-0 right-0 top-0" @pointerdown="onStartDraw">
      <FieldArea
        v-for="(item, i) in areas"
        :key="i"
        :ref="setAreaRefs"
        :area="item.area"
        :field="item.field"
        :editable="editable"
        :default-field="defaultFields.find((f) => f.name === item.field.name)"
        @start-resize="resizeDirection = $event"
        @stop-resize="resizeDirection = null"
        @start-drag="isMove = true"
        @stop-drag="isMove = false"
        @remove="$emit('remove-area', item.area)"
      />
      <FieldArea
        v-if="newArea"
        :is-draw="true"
        :field="{ submitter_id: selectedSubmitter.id, type: drawField?.type || defaultFieldType }"
        :area="newArea"
      />
    </div>
    <div
      v-show="resizeDirection || isMove || isDrag || showMask || drawField"
      id="mask"
      ref="mask"
      class="absolute bottom-0 left-0 right-0 top-0 z-10"
      :class="{
        'cursor-grab': isDrag || isMove,
        'cursor-nwse-resize': drawField,
        [resizeDirectionClasses[resizeDirection]]: !!resizeDirectionClasses
      }"
      @pointermove="onPointermove"
      @pointerdown="onStartDraw"
      @dragover.prevent
      @drop="onDrop"
      @pointerup="onPointerup"
    />
  </div>
</template>

<script>
import FieldArea from "@/components/template/Area.vue";

export default {
  components: {
    FieldArea
  },
  props: {
    image: {
      type: Object,
      required: true
    },
    areas: {
      type: Array,
      required: false,
      default: () => []
    },
    defaultFields: {
      type: Array,
      required: false,
      default: () => []
    },
    allowDraw: {
      type: Boolean,
      required: false,
      default: true
    },
    selectedSubmitter: {
      type: Object,
      required: true
    },
    drawField: {
      type: Object,
      required: false,
      default: null
    },
    editable: {
      type: Boolean,
      required: false,
      default: true
    },
    isDrag: {
      type: Boolean,
      required: false,
      default: false
    },
    number: {
      type: Number,
      required: true
    }
  },
  emits: ["draw", "drop-field", "remove-area"],
  data() {
    return {
      areaRefs: [],
      showMask: false,
      isMove: false,
      resizeDirection: null,
      newArea: null
    };
  },
  computed: {
    defaultFieldType() {
      return "text";
    },
    resizeDirectionClasses() {
      return {
        nwse: "cursor-nwse-resize",
        ew: "cursor-ew-resize"
      };
    },
    width() {
      return this.image.metadata.width;
    },
    height() {
      return this.image.metadata.height;
    }
  },
  beforeUpdate() {
    this.areaRefs = [];
  },
  methods: {
    onImageLoad(e) {
      e.target.setAttribute("width", e.target.naturalWidth);
      e.target.setAttribute("height", e.target.naturalHeight);
    },
    setAreaRefs(el) {
      if (el) {
        this.areaRefs.push(el);
      }
    },
    onDrop(e) {
      this.$emit("drop-field", {
        x: e.layerX,
        y: e.layerY,
        maskW: this.$refs.mask.clientWidth,
        maskH: this.$refs.mask.clientHeight,
        page: this.number
      });
    },
    onStartDraw(e) {
      if (!this.allowDraw) {
        return;
      }

      if (!this.drawField) {
        return;
      }

      if (!this.editable) {
        return;
      }

      this.showMask = true;

      this.$nextTick(() => {
        this.newArea = {
          initialX: e.layerX / this.$refs.mask.clientWidth,
          initialY: e.layerY / this.$refs.mask.clientHeight,
          x: e.layerX / this.$refs.mask.clientWidth,
          y: e.layerY / this.$refs.mask.clientHeight,
          w: 0,
          h: 0
        };
      });
    },
    onPointermove(e) {
      if (this.newArea) {
        const dx = e.layerX / this.$refs.mask.clientWidth - this.newArea.initialX;
        const dy = e.layerY / this.$refs.mask.clientHeight - this.newArea.initialY;

        if (dx > 0) {
          this.newArea.x = this.newArea.initialX;
        } else {
          this.newArea.x = e.layerX / this.$refs.mask.clientWidth;
        }

        if (dy > 0) {
          this.newArea.y = this.newArea.initialY;
        } else {
          this.newArea.y = e.layerY / this.$refs.mask.clientHeight;
        }

        if (this.drawField?.type === "cells") {
          this.newArea.cell_w = this.newArea.h * (this.$refs.mask.clientHeight / this.$refs.mask.clientWidth);
        }

        this.newArea.w = Math.abs(dx);
        this.newArea.h = Math.abs(dy);
      }
    },
    onPointerup(e) {
      if (this.newArea) {
        const area = {
          x: this.newArea.x,
          y: this.newArea.y,
          w: this.newArea.w,
          h: this.newArea.h,
          page: this.number
        };

        if ("cell_w" in this.newArea) {
          area.cell_w = this.newArea.cell_w;
        }

        this.$emit("draw", area);
      }

      this.showMask = false;
      this.newArea = null;
    }
  }
};
</script>
