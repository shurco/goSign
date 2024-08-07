<template>
  <div
    class="group absolute overflow-visible"
    :style="positionStyle"
    @pointerdown.stop
    @mousedown.stop="startDrag"
    @touchstart="startTouchDrag"
  >
    <div
      v-if="isSelected || isDraw"
      class="border-1.5 pointer-events-none absolute bottom-0 left-0 right-0 top-0 border"
      :class="borderColors[submitterIndex]"
    />
    <div v-if="field.type === 'cells' && (isSelected || isDraw)" class="absolute bottom-0 left-0 right-0 top-0">
      <div
        v-for="(cellW, index) in cells"
        :key="index"
        class="absolute bottom-0 top-0 border-r"
        :class="borderColors[submitterIndex]"
        :style="{ left: (cellW / area.w) * 100 + '%' }"
      >
        <span
          v-if="index === 0 && editable"
          class="absolute -bottom-1 z-10 h-2.5 w-2.5 cursor-ew-resize rounded-full border border-gray-400 bg-white shadow-md"
          style="left: -4px"
          @mousedown.stop="startResizeCell"
        />
      </div>
    </div>

    <div
      v-if="field?.type"
      class="absolute overflow-visible whitespace-nowrap rounded-t border bg-white group-hover:z-10 group-hover:flex"
      :class="{ 'z-10 flex': isNameFocus || isSelected, invisible: !isNameFocus && !isSelected }"
      style="top: -25px; height: 25px"
      @mousedown.stop
      @pointerdown.stop
    >
      <FieldSubmitter
        v-model="field.submitter_id"
        class="border-r"
        :compact="true"
        :editable="editable && !defaultField"
        :menu-classes="'dropdown-content bg-white menu menu-xs p-2 shadow rounded-box w-52 rounded-t-none -left-[1px]'"
        :submitters="template.submitters"
        @update:model-value="save"
        @click="selectedAreaRef = area"
      />
      <FieldType
        v-model="field.type"
        :button-width="27"
        :editable="editable && !defaultField"
        :button-classes="'px-1'"
        :menu-classes="'bg-white rounded-t-none'"
        @update:model-value="[maybeUpdateOptions(), save]"
        @click="selectedAreaRef = area"
      />
      <span
        v-if="field.type !== 'checkbox' || field.name"
        ref="name"
        :contenteditable="editable && !defaultField"
        dir="auto"
        class="block cursor-text pr-1 outline-none"
        style="min-width: 2px"
        @keydown.enter.prevent="onNameEnter"
        @focus="onNameFocus"
        @blur="onNameBlur"
        >{{ optionIndexText }} {{ field.name || defaultName }}</span
      >
      <div v-if="isNameFocus && !['checkbox', 'phone'].includes(field.type)" class="ml-1.5 flex items-center">
        <input
          :id="`required-checkbox-${field.id}`"
          v-model="field.required"
          type="checkbox"
          class="checkbox checkbox-xs no-animation rounded"
          @mousedown.prevent
        />
        <label
          :for="`required-checkbox-${field.id}`"
          class="label text-xs"
          @click.prevent="field.required = !field.required"
          @mousedown.prevent
          >Required</label
        >
      </div>
      <button v-else-if="editable" class="pr-1" title="Remove" @click.prevent="$emit('remove')">
        <SvgIcon name="x" class="h-4 w-4" />
      </button>
    </div>

    <div
      class="flex h-full w-full items-center"
      :class="[bgColors[submitterIndex], field?.default_value ? '' : 'justify-center']"
    >
      <span v-if="field" class="flex h-full items-center justify-center space-x-1">
        <div
          v-if="field?.default_value"
          :class="{ 'text-[1.5vw] lg:text-base': !textOverflowChars, 'text-[1.0vw] lg:text-xs': textOverflowChars }"
        >
          <div ref="textContainer" class="flex items-center px-0.5">
            <span class="whitespace-pre-wrap">{{ field.default_value }}</span>
          </div>
        </div>
        <SvgIcon v-else :name="fieldIcons[field.type]" width="100%" height="100%" class="max-h-10 opacity-50" />
      </span>
    </div>

    <div ref="touchTarget" class="absolute bottom-0 left-0 right-0 top-0 cursor-pointer" />
    <span
      v-if="field?.type && editable"
      class="absolute -bottom-1 -right-1 h-4 w-4 cursor-nwse-resize rounded-full border border-gray-400 bg-white shadow-md md:h-2.5 md:w-2.5"
      @mousedown.stop="startResize"
      @touchstart="startTouchResize"
    />
  </div>
</template>

<script>
import FieldSubmitter from "@/components/field/Submitter.vue";
import FieldType from "@/components/field/Type.vue";
import { bgColors, borderColors, fieldIcons, fieldNames } from "@/components/field/constants.ts";
import { v4 } from "uuid";

export default {
  name: "FieldArea",
  components: {
    FieldType,
    FieldSubmitter
  },

  inject: ["template", "selectedAreaRef", "save"],

  props: {
    area: {
      type: Object,
      required: true
    },
    isDraw: {
      type: Boolean,
      required: false,
      default: false
    },
    defaultField: {
      type: Object,
      required: false,
      default: null
    },
    editable: {
      type: Boolean,
      required: false,
      default: true
    },
    field: {
      type: Object,
      required: false,
      default: null
    }
  },

  emits: ["start-resize", "stop-resize", "start-drag", "stop-drag", "remove"],

  data() {
    return {
      isResize: false,
      isDragged: false,
      isNameFocus: false,
      textOverflowChars: 0,
      dragFrom: { x: 0, y: 0 }
    };
  },
  computed: {
    defaultName() {
      return "name";
    },
    borderColors() {
      return borderColors;
    },
    fieldNames() {
      return fieldNames;
    },
    fieldIcons() {
      return fieldIcons;
    },
    bgColors() {
      return bgColors;
    },

    optionIndexText() {
      if (this.area.option_id && this.field.options) {
        return `${this.field.options.findIndex((o) => o.id === this.area.option_id) + 1}.`;
      } else {
        return "";
      }
    },
    cells() {
      const cells = [];
      let currentWidth = 0;
      while (currentWidth + (this.area.cell_w + this.area.cell_w / 4) < this.area.w) {
        currentWidth += this.area.cell_w || 9999999;
        cells.push(currentWidth);
      }
      return cells;
    },
    submitter() {
      return this.template.submitters.find((s) => s.id === this.field.submitter_id);
    },
    submitterIndex() {
      return this.template.submitters.indexOf(this.submitter);
    },

    isSelected() {
      return this.selectedAreaRef === this.area;
    },

    positionStyle() {
      const { x, y, w, h } = this.area;

      return {
        top: y * 100 + "%",
        left: x * 100 + "%",
        width: w * 100 + "%",
        height: h * 100 + "%"
      };
    }
  },
  watch: {
    "field.default_value"() {
      if (
        this.field.type === "text" &&
        this.field.default_value &&
        this.$refs.textContainer &&
        (this.textOverflowChars === 0 || this.textOverflowChars - 4 > this.field.default_value.length)
      ) {
        this.textOverflowChars =
          this.$el.clientHeight < this.$refs.textContainer.clientHeight ? this.field.default_value.length : 0;
      }
    }
  },
  mounted() {
    if (
      this.field.type === "text" &&
      this.field.default_value &&
      this.$refs.textContainer &&
      (this.textOverflowChars === 0 || this.textOverflowChars - 4 > this.field.default_value)
    ) {
      this.$nextTick(() => {
        this.textOverflowChars =
          this.$el.clientHeight < this.$refs.textContainer.clientHeight ? this.field.default_value.length : 0;
      });
    }
  },
  methods: {
    onNameFocus(e) {
      this.selectedAreaRef = this.area;
      this.isNameFocus = true;
      this.$refs.name.style.minWidth = this.$refs.name.clientWidth + "px";

      if (!this.field.name) {
        setTimeout(() => {
          this.$refs.name.innerText = " ";
        }, 1);
      }
    },
    startResizeCell(e) {
      this.$el.getRootNode().addEventListener("mousemove", this.onResizeCell);
      this.$el.getRootNode().addEventListener("mouseup", this.stopResizeCell);
      this.$emit("start-resize", "ew");
    },
    stopResizeCell(e) {
      this.$el.getRootNode().removeEventListener("mousemove", this.onResizeCell);
      this.$el.getRootNode().removeEventListener("mouseup", this.stopResizeCell);
      this.$emit("stop-resize");
      this.save;
    },
    onResizeCell(e) {
      if (e.target.id === "mask") {
        const positionX = e.layerX / (e.target.clientWidth - 1);

        if (positionX > this.area.x) {
          this.area.cell_w = positionX - this.area.x;
        }
      }
    },
    maybeUpdateOptions() {
      delete this.field.default_value;

      if (!["radio", "multiple", "select"].includes(this.field.type)) {
        delete this.field.options;
      }

      if (["select", "multiple", "radio"].includes(this.field.type)) {
        this.field.options ||= [{ value: "", id: v4() }];
      }

      (this.field.areas || []).forEach((area) => {
        if (this.field.type === "cells") {
          area.cell_w = (area.w * 2) / Math.floor(area.w / area.h);
        } else {
          delete area.cell_w;
        }
      });
    },
    onNameBlur(e) {
      const text = this.$refs.name.innerText.trim();
      this.isNameFocus = false;
      this.$refs.name.style.minWidth = "";
      if (text) {
        this.field.name = text;
      } else {
        this.field.name = "";
        this.$refs.name.innerText = this.defaultName;
      }

      this.save;
    },
    onNameEnter(e) {
      this.$refs.name.blur();
    },
    resize(e) {
      if (e.target.id === "mask") {
        this.area.w = e.layerX / e.target.clientWidth - this.area.x;
        this.area.h = e.layerY / e.target.clientHeight - this.area.y;
      }
    },
    drag(e) {
      if (e.target.id === "mask") {
        this.isDragged = true;
        this.area.x = (e.layerX - this.dragFrom.x) / e.target.clientWidth;
        this.area.y = (e.layerY - this.dragFrom.y) / e.target.clientHeight;
      }
    },
    startDrag(e) {
      this.selectedAreaRef = this.area;

      if (!this.editable) {
        return;
      }
      const rect = e.target.getBoundingClientRect();
      this.dragFrom = { x: e.clientX - rect.left, y: e.clientY - rect.top };
      this.$el.getRootNode().addEventListener("mousemove", this.drag);
      this.$el.getRootNode().addEventListener("mouseup", this.stopDrag);
      this.$emit("start-drag");
    },
    startTouchDrag(e) {
      if (e.target !== this.$refs.touchTarget) {
        return;
      }
      this.$refs?.name?.blur();
      e.preventDefault();
      this.isDragged = true;
      const rect = e.target.getBoundingClientRect();
      this.selectedAreaRef = this.area;
      this.dragFrom = { x: rect.left - e.touches[0].clientX, y: rect.top - e.touches[0].clientY };
      this.$el.getRootNode().addEventListener("touchmove", this.touchDrag);
      this.$el.getRootNode().addEventListener("touchend", this.stopTouchDrag);
      this.$emit("start-drag");
    },
    touchDrag(e) {
      const page = this.$parent.$refs.mask.previousSibling;
      const rect = page.getBoundingClientRect();
      this.area.x = (this.dragFrom.x + e.touches[0].clientX - rect.left) / rect.width;
      this.area.y = (this.dragFrom.y + e.touches[0].clientY - rect.top) / rect.height;
    },
    stopTouchDrag() {
      this.$el.getRootNode().removeEventListener("touchmove", this.touchDrag);
      this.$el.getRootNode().removeEventListener("touchend", this.stopTouchDrag);
      if (this.isDragged) {
        this.save;
      }
      this.isDragged = false;
      this.$emit("stop-drag");
    },
    stopDrag() {
      this.$el.getRootNode().removeEventListener("mousemove", this.drag);
      this.$el.getRootNode().removeEventListener("mouseup", this.stopDrag);
      if (this.isDragged) {
        this.save;
      }
      this.isDragged = false;
      this.$emit("stop-drag");
    },
    startResize() {
      this.selectedAreaRef = this.area;
      this.$el.getRootNode().addEventListener("mousemove", this.resize);
      this.$el.getRootNode().addEventListener("mouseup", this.stopResize);
      this.$emit("start-resize", "nwse");
    },
    stopResize() {
      this.$el.getRootNode().removeEventListener("mousemove", this.resize);
      this.$el.getRootNode().removeEventListener("mouseup", this.stopResize);
      this.$emit("stop-resize");
      this.save;
    },
    startTouchResize(e) {
      this.selectedAreaRef = this.area;
      this.$refs?.name?.blur();
      e.preventDefault();
      this.$el.getRootNode().addEventListener("touchmove", this.touchResize);
      this.$el.getRootNode().addEventListener("touchend", this.stopTouchResize);
      this.$emit("start-resize", "nwse");
    },
    touchResize(e) {
      const page = this.$parent.$refs.mask.previousSibling;
      const rect = page.getBoundingClientRect();
      this.area.w = (e.touches[0].clientX - rect.left) / rect.width - this.area.x;
      this.area.h = (e.touches[0].clientY - rect.top) / rect.height - this.area.y;
    },
    stopTouchResize() {
      this.$el.getRootNode().removeEventListener("touchmove", this.touchResize);
      this.$el.getRootNode().removeEventListener("touchend", this.stopTouchResize);
      this.$emit("stop-resize");
      this.save;
    }
  }
};
</script>
