<template>
  <!-- Backdrop -->
  <div class="fixed inset-0 z-40" @click="$emit('close')"></div>

  <!-- Menu -->
  <div
    ref="menuRef"
    class="ring-opacity-5 absolute z-50 mt-2 w-48 rounded-md border border-gray-200 bg-white ring-1 ring-black transition-colors hover:border-gray-300"
    :style="{ top: position.top + 'px', left: position.left + 'px' }"
  >
    <div class="py-1" role="menu" aria-orientation="vertical">
      <!-- Edit Organization -->
      <button
        class="flex w-full items-center px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900"
        role="menuitem"
        @click="$emit('edit')"
      >
        <PencilIcon class="mr-3 h-4 w-4" />
        Edit Organization
      </button>

      <!-- Manage Members -->
      <button
        class="flex w-full items-center px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900"
        role="menuitem"
        @click="$emit('manage-members')"
      >
        <UsersIcon class="mr-3 h-4 w-4" />
        Manage Members
      </button>

      <!-- Divider -->
      <div class="my-1 border-t border-gray-100"></div>

      <!-- Delete Organization -->
      <button
        v-if="canDelete"
        class="flex w-full items-center px-4 py-2 text-left text-sm text-red-700 hover:bg-red-50 hover:text-red-900"
        role="menuitem"
        @click="$emit('delete')"
      >
        <TrashIcon class="mr-3 h-4 w-4" />
        Delete Organization
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { PencilIcon, TrashIcon, UsersIcon } from "@heroicons/vue/24/outline";

interface Props {
  organization: {
    id: string;
    name: string;
    owner_id: string;
  };
}

const props = defineProps<Props>();

interface Emits {
  (e: "close"): void;
  (e: "edit"): void;
  (e: "manage-members"): void;
  (e: "delete"): void;
}

const emit = defineEmits<Emits>();

const menuRef = ref<HTMLElement>();
const position = ref({ top: 0, left: 0 });

// TODO: Get current user ID from store/auth
const currentUserId = ref("user-id");

const canDelete = computed(() => {
  return props.organization.owner_id === currentUserId.value;
});

const updatePosition = (event: MouseEvent) => {
  const rect = (event.target as HTMLElement).getBoundingClientRect();
  position.value = {
    top: rect.bottom + window.scrollY,
    left: rect.left + window.scrollX
  };
};

const handleClickOutside = (event: MouseEvent) => {
  if (menuRef.value && !menuRef.value.contains(event.target as Node)) {
    emit("close");
  }
};

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
});

// Expose updatePosition method for parent component
defineExpose({
  updatePosition
});
</script>
