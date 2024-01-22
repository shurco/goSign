<template>
  <component :is="route.meta.layoutComponent">
    <RouterView />
  </component>
</template>

<script setup lang="ts">
import { RouterView, useRoute } from "vue-router";
const route = useRoute();
</script>