<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

const props = withDefaults(
  defineProps<{
    as?: string
    class?: HTMLAttributes['class']
  }>(),
  { as: 'h2' },
)
</script>

<template>
  <component
    :is="props.as"
    data-slot="card-title"
    :class="
      cn('cn-text-title group-data-[size=sm]/card:cn-text-body-sm cn-font-heading', props.class)
    "
  >
    <slot />
  </component>
</template>
