<template>
  <div class="conditions-builder">
    <p class="mb-4 text-sm text-gray-600">
      {{ $t('fields.conditions.description') || 'Define when this field is shown, hidden, required or disabled based on other fields.' }}
    </p>

    <!-- Empty state -->
    <div
      v-if="!conditions.length"
      class="rounded-xl border-2 border-dashed border-gray-200 bg-gray-50/50 py-10 text-center"
    >
      <p class="mb-3 text-sm text-gray-500">{{ $t('fields.conditions.empty') || 'No rules yet.' }}</p>
      <button
        type="button"
        class="inline-flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700"
        @click="addGroup"
      >
        <SvgIcon name="plus" class="h-4 w-4" />
        {{ $t('fields.conditions.addGroup') || '+ Add rule group' }}
      </button>
    </div>

    <!-- Rule groups -->
    <div v-else class="space-y-4">
      <div
        v-for="(group, groupIndex) in conditions"
        :key="groupIndex"
        class="rule-group rounded-xl border border-gray-200 bg-white shadow-sm overflow-hidden"
      >
        <!-- Group header: rule number + "If ... then" line -->
        <div class="border-b border-gray-100 bg-gray-50/80 px-4 py-3">
          <div class="flex flex-wrap items-center gap-2">
            <span class="text-xs font-medium uppercase tracking-wide text-gray-500">
              {{ $t('fields.conditions.rule') || 'Rule' }} {{ groupIndex + 1 }}
            </span>
            <span class="text-gray-400">·</span>
            <Select
              v-model="group.logic"
              class="h-9 w-[5.5rem] shrink-0 text-sm"
            >
              <option value="AND">{{ $t('fields.conditions.and') }}</option>
              <option value="OR">{{ $t('fields.conditions.or') }}</option>
            </Select>
            <span class="text-sm text-gray-500">{{ $t('fields.conditions.then') || 'then' }}</span>
            <Select
              v-model="group.action"
              class="h-9 min-w-[8.5rem] shrink-0 text-sm"
            >
              <option value="show">{{ $t('fields.conditions.show') }}</option>
              <option value="hide">{{ $t('fields.conditions.hide') }}</option>
              <option value="require">{{ $t('fields.conditions.require') }}</option>
              <option value="disable">{{ $t('fields.conditions.disable') }}</option>
            </Select>
            <span class="text-sm text-gray-500">{{ $t('fields.conditions.thisField') || 'this field' }}</span>
          </div>
        </div>

        <!-- Condition rows -->
        <div class="p-4">
          <div class="space-y-3">
            <div
              v-for="(condition, condIndex) in group.conditions"
              :key="condIndex"
              class="grid grid-cols-[minmax(11rem,1fr)_minmax(9rem,max-content)_minmax(7rem,1fr)_auto] items-center gap-3"
            >
              <Select v-model="condition.field_id" class="min-w-0 text-sm">
                <option value="">{{ $t('fields.conditions.selectField') }}</option>
                <option v-for="f in availableFields" :key="f.id" :value="f.id">
                  {{ f.displayName ?? f.name ?? f.id }}
                </option>
              </Select>
              <Select v-model="condition.operator" class="min-w-[9rem] shrink-0 text-sm">
                <option value="equals">{{ $t('fields.conditions.equals') }}</option>
                <option value="not_equals">{{ $t('fields.conditions.notEquals') }}</option>
                <option value="contains">{{ $t('fields.conditions.contains') }}</option>
                <option value="not_contains">{{ $t('fields.conditions.notContains') }}</option>
                <option value="greater_than">{{ $t('fields.conditions.greaterThan') }}</option>
                <option value="less_than">{{ $t('fields.conditions.lessThan') }}</option>
                <option value="is_empty">{{ $t('fields.conditions.isEmpty') }}</option>
                <option value="is_not_empty">{{ $t('fields.conditions.isNotEmpty') }}</option>
              </Select>
              <Input
                v-if="condition.operator !== 'is_empty' && condition.operator !== 'is_not_empty'"
                v-model="condition.value"
                :placeholder="$t('fields.conditions.value')"
                class="min-w-0 text-sm"
              />
              <div v-else />
              <button
                type="button"
                class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md text-gray-400 hover:bg-red-50 hover:text-red-600"
                :aria-label="$t('common.delete')"
                @click="removeCondition(groupIndex, condIndex)"
              >
                <SvgIcon name="trash-x" class="h-4 w-4" />
              </button>
            </div>
          </div>

          <div class="mt-3 flex flex-wrap items-center justify-between gap-2 border-t border-gray-100 pt-3">
            <button
              type="button"
              class="inline-flex items-center gap-1.5 rounded-md border border-gray-200 bg-white px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-50"
              @click="addCondition(groupIndex)"
            >
              <SvgIcon name="plus" class="h-3.5 w-3.5" />
              {{ $t('fields.conditions.addCondition') }}
            </button>
            <button
              type="button"
              class="inline-flex items-center gap-1.5 rounded-md border border-transparent px-3 py-1.5 text-sm text-red-600 hover:bg-red-50"
              @click="removeGroup(groupIndex)"
            >
              <SvgIcon name="trash-x" class="h-3.5 w-3.5" />
              {{ $t('common.delete') }}
            </button>
          </div>
        </div>
      </div>

      <button
        type="button"
        class="w-full rounded-xl border-2 border-dashed border-gray-200 py-3 text-sm font-medium text-gray-600 hover:border-indigo-300 hover:bg-indigo-50/50 hover:text-indigo-700"
        @click="addGroup"
      >
        + {{ $t('fields.conditions.addGroup') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Field, FieldConditionGroup } from '@/models/template'
import Select from '@/components/ui/Select.vue'
import Input from '@/components/ui/Input.vue'
import SvgIcon from '@/components/SvgIcon.vue'

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
