<template>
  <div class="home">
    <h1>{{ t('welcome') }}</h1>
    <p>{{ $t('user.greeting') }}</p>
    <button @click="handleClick">{{ t('common.submit') }}</button>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const handleClick = () => {
  console.log(t('user.avatar'))
}
</script>
