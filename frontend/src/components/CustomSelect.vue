<template>
  <div class="custom-select" @blur="open = false" tabindex="0">
    <div class="selected" @click="open = !open">
      {{ selectedLabel }}
      <span class="arrow" :class="{ up: open }">▼</span>
    </div>
    <div class="items" v-show="open">
      <div 
        v-for="option in options" 
        :key="option.value"
        class="item"
        :class="{ active: option.value === modelValue }"
        @click.stop="select(option)"
      >
        {{ option.label }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

const props = defineProps<{
  modelValue: string;
  options: { value: string; label: string }[];
}>();

const emit = defineEmits(['update:modelValue', 'change']);

const open = ref(false);

const selectedLabel = computed(() => {
  const opt = props.options.find(o => o.value === props.modelValue);
  return opt ? opt.label : props.modelValue;
});

const select = (option: { value: string; label: string }) => {
  emit('update:modelValue', option.value);
  emit('change', option.value);
  open.value = false;
};
</script>

<style scoped>
.custom-select {
  position: relative;
  width: 100%;
  outline: none;
  font-size: 13px;
  font-family: 'Inter', sans-serif;
  user-select: none;
  --wails-draggable: no-drag;
}

.selected {
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  color: var(--text-primary);
  padding: 8px 12px;
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
  transition: var(--transition);
}

.custom-select:focus .selected {
  border-color: var(--accent);
  box-shadow: 0 0 0 2px var(--accent-light);
}

.arrow {
  font-size: 10px;
  transition: transform 0.2s;
  color: var(--text-secondary);
}

.arrow.up {
  transform: rotate(180deg);
}

.items {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: #2a2a2a;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
  margin-top: 4px;
  z-index: 1000;
  box-shadow: 0 4px 12px rgba(0,0,0,0.5);
}

.item {
  padding: 8px 12px;
  cursor: pointer;
  color: var(--text-primary);
  transition: background 0.2s;
}

.item:hover {
  background: rgba(255, 255, 255, 0.1);
}

.item.active {
  background: var(--accent);
  color: white;
}
</style>
