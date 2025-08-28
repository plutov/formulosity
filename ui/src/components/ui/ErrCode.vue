<template>
  <div class="mb-4 flex items-center p-4 text-sm text-orange-800 border border-orange-300 rounded-lg bg-orange-50">
    <Icon icon="heroicons:information-circle" class="flex-shrink-0 inline w-4 h-4 me-3" />
    <span class="sr-only">Info</span>
    <div>
      <p>{{ message }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'

interface Props {
  message?: string
}

withDefaults(defineProps<Props>(), {
  message: 'An error occurred'
})
</script>