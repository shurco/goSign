<template>
  <div class="rounded-2xl border border-[var(--color-base-300)] bg-[var(--color-base-100)] shadow-xl">
    <div v-if="$slots.header" class="border-b border-[var(--color-base-300)] px-6 py-5">
      <slot name="header" />
    </div>
    <div class="px-6 py-5">
      <slot />
    </div>
    <div v-if="$slots.footer" class="border-t border-[var(--color-base-300)] bg-[var(--color-base-200)]/30 px-6 py-4">
      <slot name="footer" />
    </div>
  </div>
</template>

<script setup lang="ts">
// Simple card component with header, body, and footer slots
</script>
