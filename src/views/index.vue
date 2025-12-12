<script lang="ts">
const t = (key: string) => {
  return key;
};
</script>

<template>
  <div class="test">
    <h1>test</h1>
    <span>{{ t("test.logo") }}</span>
    <span>{{ t("test.avatar") }}</span>
    <span>{{ t("test.user") }}</span>
    <span>{{ t("test.name") }}</span>
  </div>
</template>
