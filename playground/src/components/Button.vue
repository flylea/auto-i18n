<template>
  <button class="btn" @click="$t('button.click')">
    {{ $t('button.submit') }}
  </button>
</template>

<script setup>
// Script section also uses t()
const msg = t('button.loading')
</script>
