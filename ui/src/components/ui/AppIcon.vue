<template>
  <Icon :icon="iconName" :class="classes" v-bind="$attrs" />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Icon } from '@iconify/vue'

interface Props {
  name: 
    // Navigation & UI
    | 'menu' 
    | 'close'
    | 'chevron-up'
    | 'chevron-down'
    | 'chevron-left'
    | 'chevron-right'
    | 'arrow-left'
    | 'arrow-right'
    | 'arrow-external'
    // Actions
    | 'play'
    | 'pause'
    | 'eye'
    | 'info'
    | 'book'
    // Status
    | 'check-circle'
    | 'clock'
    | 'star'
    | 'spinner'
  class?: string
}

const props = withDefaults(defineProps<Props>(), {
  class: ''
})

const iconMap = {
  // Navigation & UI
  menu: 'heroicons:bars-3-20-solid',
  close: 'heroicons:x-mark',
  'chevron-up': 'heroicons:chevron-up',
  'chevron-down': 'heroicons:chevron-down', 
  'chevron-left': 'heroicons:chevron-left',
  'chevron-right': 'heroicons:chevron-right',
  'arrow-left': 'heroicons:arrow-left',
  'arrow-right': 'heroicons:arrow-right',
  'arrow-external': 'heroicons:arrow-top-right-on-square',
  // Actions
  play: 'heroicons:play',
  pause: 'heroicons:pause',
  eye: 'heroicons:eye',
  info: 'heroicons:information-circle',
  book: 'heroicons:book-open',
  // Status
  'check-circle': 'heroicons:check-circle',
  clock: 'heroicons:clock',
  star: 'heroicons:star-solid',
  spinner: 'svg-spinners:180-ring-with-bg',
}

const iconName = computed(() => iconMap[props.name])
const classes = computed(() => props.class || 'w-5 h-5')
</script>