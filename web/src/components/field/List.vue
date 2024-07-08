<template>
  <div class="sticky top-0 z-10">
    <FieldSubmitter
      :model-value="selectedSubmitter.id"
      class="w-full rounded-lg bg-[#faf7f5]"
      :submitters="submitters"
      :editable="editable && !defaultSubmitters.length"
      @new-submitter="save"
      @remove="removeSubmitter"
      @name-change="save"
      @update:model-value="
        $emit(
          'change-submitter',
          submitters.find((s: any) => s.id === $event)
        )
      "
    />
  </div>

  <div ref="fields" class="mb-1 mt-2" @dragover.prevent="onFieldDragover" @drop="reorderFields">
    <Field
      v-for="field in submitterFields"
      :key="field.id"
      :data-uuid="field.id"
      :field="field"
      :type-index="fields.filter((f: any) => f.type === field.type).indexOf(field)"
      :editable="editable && (!dragField || dragField !== field)"
      :default-field="defaultFields.find((f: any) => f.name === field.name)"
      :draggable="editable"
      @dragstart="dragField = field"
      @dragend="dragField = null"
      @remove="removeField"
      @scroll-to="$emit('scroll-to-area', $event)"
      @set-draw="$emit('set-draw', $event)"
    />
  </div>

  <div v-if="editable && !onlyDefinedFields" class="grid grid-cols-3 gap-1 pb-2">
    <template v-for="(icon, type) in fieldIcons" :key="type">
      <button
        draggable="true"
        class="group relative flex w-full items-center justify-center rounded border border-dashed border-[#e7e2df] hover:border-[#291334]/20"
        @dragstart="onDragstart({ type: type })"
        @dragend="$emit('drag-end')"
        @click="addField(type)"
      >
        <div class="items-console absolute left-0 flex h-full cursor-grab transition-all group-hover:bg-[#efeae6]/50">
          <SvgIcon name="drag" width="18" height="18" class="my-auto" />
        </div>
        <div class="flex flex-col items-center px-2 py-2">
          <SvgIcon :name="icon" width="20" height="20" />
          <span class="mt-1 text-xs">{{ fieldNames[type] }}</span>
        </div>
      </button>
    </template>
  </div>

  <div v-if="fields.length < 4 && editable" class="rounded border border-[#efeae6] p-2 text-xs">
    <ul class="ml-2 list-outside list-disc pl-2">
      <li>Draw a text field on the page with a mouse</li>
      <li>Drag &amp; drop any other field type on the page</li>
      <li>Click on the field type above to start drawing it</li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, ref } from "vue";
import Field from "@/components/field/Field.vue";
import FieldSubmitter from "@/components/field/Submitter.vue";
import { fieldIcons as fieldIconsConst, fieldNames as fieldNamesConst } from "@/components/field/constants.ts";
import { v4 } from "uuid";

const props = defineProps({
  fields: {
    type: Array,
    required: true
  },
  editable: {
    type: Boolean,
    required: false,
    default: true
  },
  defaultFields: {
    type: Array,
    required: false,
    default: () => []
  },
  onlyDefinedFields: {
    type: Boolean,
    required: false,
    default: true
  },
  defaultSubmitters: {
    type: Array,
    required: false,
    default: () => []
  },
  submitters: {
    type: Array,
    required: true
  },
  selectedSubmitter: {
    type: Object,
    required: true
  }
});

const emit = defineEmits(["set-draw", "set-drag", "drag-end", "scroll-to-area", "change-submitter"]);

const save = inject("save");

const dragField = ref();
const fieldsRef = ref();

const fieldNames = computed(() => fieldNamesConst);
const fieldIcons = computed(() => fieldIconsConst);
const submitterFields = computed(() => props.fields.filter((f: any) => f.submitter_id === props.selectedSubmitter.id));
const submitterDefaultFields = computed(() =>
  props.defaultFields.filter((f: any) => {
    return (
      !submitterFields.value.find((field: any) => field.name === f.name) &&
      (!f.role || f.role === props.selectedSubmitter.name)
    );
  })
);

onMounted(() => {
  fieldsRef.value = [];
});

function onDragstart(field: any): void {
  emit("set-drag", field);
}

function onFieldDragover(e: any): void {
  const targetField = e.target.closest("[data-uuid]");
  const dragFieldElement = dragField.value;

  if (dragFieldElement && targetField && targetField !== dragFieldElement) {
    const fields = Array.from(e.currentTarget.children);
    const currentIndex = fields.indexOf(dragFieldElement);
    const targetIndex = fields.indexOf(targetField);

    if (currentIndex < targetIndex) {
      targetField.after(dragFieldElement);
    } else {
      targetField.before(dragFieldElement);
    }
  }
}

function reorderFields(): void {
  Array.from(fieldsRef.value.children).forEach((el: any, index: number) => {
    if (el.dataset.id !== props.fields[index].id) {
      const field = props.fields.find((f: any) => f.id === el.dataset.id);
      props.fields.splice(props.fields.indexOf(field), 1);
      props.fields.splice(index, 0, field);
    }
  });
  save;
}

function removeSubmitter(submitter: any): void {
  [...props.fields].forEach((field: any) => {
    if (field.submitter_id === submitter.id) {
      removeField(field);
    }
  });
  props.submitters.splice(props.submitters.indexOf(submitter), 1);

  if (props.selectedSubmitter === submitter) {
    emit("change-submitter", props.submitters[0]);
  }
  save;
}

function removeField(field: any): void {
  props.fields.splice(props.fields.indexOf(field), 1);
  save;
}

function addField(type: any, area = null): void {
  const field: any = {
    name: "",
    id: v4(),
    required: type !== "checkbox",
    areas: [],
    submitter_id: props.selectedSubmitter.id,
    type
  };

  if (["select", "multiple", "radio"].includes(type)) {
    field.options = [{ value: "", id: v4() }];
  }

  if (type === "stamp") {
    field.readonly = true;
  }

  if (type === "date") {
    field.preferences = {
      format: Intl.DateTimeFormat().resolvedOptions().locale.endsWith("-US") ? "MM/DD/YYYY" : "DD/MM/YYYY"
    };
  }

  props.fields.push(field);

  if (!["payment", "file"].includes(type)) {
    emit("set-draw", { field });
  }
  save;
}
</script>
