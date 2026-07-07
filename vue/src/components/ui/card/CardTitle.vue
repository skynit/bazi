<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { cn } from "@/lib/utils"

const props = defineProps<{
  class?: HTMLAttributes["class"]
}>()
</script>

<template>
  <div
    data-slot="card-title"
    :class="cn('cn-text-title group-data-[size=sm]/card:cn-text-body-sm cn-font-heading', props.class)"
  >
    <slot />
  </div>
</template>
