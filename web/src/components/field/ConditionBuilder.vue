<template>
  <div class="space-y-4">
    <h3 class="mb-4 text-lg font-semibold">{{ $t('fields.conditions.title') }}</h3>
    
    <div v-for="(group, groupIndex) in conditions" :key="groupIndex" class="mb-4 rounded-lg border border-gray-200 p-4">
      <div class="mb-3 flex items-center gap-2">
        <!-- Logic selector (AND/OR) -->
        <Select v-model="group.logic" class="w-24">
          <option value="AND">{{ $t('fields.conditions.and') }}</option>
          <option value="OR">{{ $t('fields.conditions.or') }}</option>
        </Select>
        
        <!-- Action selector -->
        <Select v-model="group.action" class="flex-1">
          <option value="show">{{ $t('fields.conditions.show') }}</option>
          <option value="hide">{{ $t('fields.conditions.hide') }}</option>
          <option value="require">{{ $t('fields.conditions.require') }}</option>
          <option value="disable">{{ $t('fields.conditions.disable') }}</option>
        </Select>
        
        <button
          class="rounded-md border border-gray-300 bg-white px-2 py-1 text-xs hover:bg-gray-50"
          @click="removeGroup(groupIndex)"
        >
          {{ $t('common.delete') }}
        </button>
      </div>
      
      <!-- Conditions in group -->
      <div v-for="(condition, condIndex) in group.conditions" :key="condIndex" class="mb-2 flex gap-2">
        <!-- Field selector -->
        <Select v-model="condition.field_id" class="flex-1">
          <option value="">{{ $t('fields.conditions.selectField') }}</option>
          <option v-for="field in availableFields" :key="field.id" :value="field.id">
            {{ field.name }}
          </option>
        </Select>
        
        <!-- Operator selector -->
        <Select v-model="condition.operator" class="w-40">
          <option value="equals">{{ $t('fields.conditions.equals') }}</option>
          <option value="not_equals">{{ $t('fields.conditions.notEquals') }}</option>
          <option value="contains">{{ $t('fields.conditions.contains') }}</option>
          <option value="not_contains">{{ $t('fields.conditions.notContains') }}</option>
          <option value="greater_than">{{ $t('fields.conditions.greaterThan') }}</option>
          <option value="less_than">{{ $t('fields.conditions.lessThan') }}</option>
          <option value="is_empty">{{ $t('fields.conditions.isEmpty') }}</option>
          <option value="is_not_empty">{{ $t('fields.conditions.isNotEmpty') }}</option>
        </Select>
        
        <!-- Value input (hidden for is_empty/is_not_empty) -->
        <Input
          v-if="condition.operator !== 'is_empty' && condition.operator !== 'is_not_empty'"
          v-model="condition.value"
          :placeholder="$t('fields.conditions.value')"
          class="flex-1"
        />
        
        <button
          class="rounded-md border border-gray-300 bg-white px-2 py-1 text-xs hover:bg-gray-50"
          @click="removeCondition(groupIndex, condIndex)"
        >
          ×
        </button>
      </div>
      
      <button
        class="rounded-md border border-gray-300 bg-white px-2 py-1 text-xs hover:bg-gray-50"
        @click="addCondition(groupIndex)"
      >
        + {{ $t('fields.conditions.addCondition') }}
      </button>
    </div>
    
    <button
      class="rounded-md bg-indigo-600 px-3 py-1.5 text-sm text-white hover:bg-indigo-700"
      @click="addGroup"
    >
      + {{ $t('fields.conditions.addGroup') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Field, FieldConditionGroup } from '@/models/template'
import Select from '@/components/ui/Select.vue'
import Input from '@/components/ui/Input.vue'

interface Props {
  field: Field
  availableFields: Field[]
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:conditions': [conditions: FieldConditionGroup[]]
}>()

const conditions = ref<FieldConditionGroup[]>(props.field.condition_groups || [])

// Watch for changes and emit updates
watch(conditions, (newConditions) => {
  emit('update:conditions', newConditions)
}, { deep: true })

function addGroup() {
  conditions.value.push({
    logic: 'AND',
    conditions: [{ field_id: '', operator: 'equals', value: '' }],
    action: 'show'
  })
}

function addCondition(groupIndex: number) {
  conditions.value[groupIndex].conditions.push({
    field_id: '',
    operator: 'equals',
    value: ''
  })
}

function removeCondition(groupIndex: number, condIndex: number) {
  conditions.value[groupIndex].conditions.splice(condIndex, 1)
  if (conditions.value[groupIndex].conditions.length === 0) {
    conditions.value.splice(groupIndex, 1)
  }
}

function removeGroup(groupIndex: number) {
  conditions.value.splice(groupIndex, 1)
}
</script>

<style scoped>
/* Styles moved to template classes */
</style>
