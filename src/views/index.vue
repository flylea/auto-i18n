<script lang="ts">
const t = (key: string) => {
  return key;
};
const $t = (key: string) => {
  return key;
};
</script>

<template>
  <h1>test</h1>
  <span>{{ t("test.avatar") }}</span>
  <span>{{ $t("test.user") }}</span>
  <span>{{ t("test.name") }}</span>
</template>
