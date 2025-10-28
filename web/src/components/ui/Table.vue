<template>
  <div class="w-full overflow-x-auto">
    <table class="w-full border-collapse">
      <thead v-if="$slots.header" class="border-b border-gray-200 bg-gray-50">
        <slot name="header" />
      </thead>
      <tbody :class="zebra ? 'divide-y divide-gray-200' : ''">
        <slot />
      </tbody>
      <tfoot v-if="$slots.footer" class="border-t border-gray-200 bg-gray-50">
        <slot name="footer" />
      </tfoot>
    </table>
  </div>
</template>

<script setup lang="ts">
interface Props {
  zebra?: boolean;
}

withDefaults(defineProps<Props>(), {
  zebra: false
});
</script>

<style scoped>
tbody.divide-y > tr:nth-child(even) {
  background-color: #f9fafb;
}

tbody > tr:hover {
  background-color: #f3f4f6;
}
</style>
