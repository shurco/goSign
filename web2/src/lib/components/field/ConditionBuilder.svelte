<script lang="ts">
  import type { Field, FieldConditionGroup } from "@/models/template";
  import Select from "@/components/ui/Select.svelte";
  import Input from "@/components/ui/Input.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";
  import { t } from "@/i18n/index.svelte";

  interface Props {
    field: Field;
    availableFields: Field[];
    onUpdateConditions?: (conditions: FieldConditionGroup[]) => void;
  }

  let { field, availableFields, onUpdateConditions }: Props = $props();

  let conditions = $state<FieldConditionGroup[]>(field.condition_groups || []);

  $effect(() => {
    onUpdateConditions?.(conditions);
  });

  function addGroup(): void {
    conditions.push({
      logic: "AND",
      conditions: [{ field_id: "", operator: "equals", value: "" }],
      action: "show"
    });
  }

  function addCondition(groupIndex: number): void {
    conditions[groupIndex].conditions.push({
      field_id: "",
      operator: "equals",
      value: ""
    });
  }

  function removeCondition(groupIndex: number, condIndex: number): void {
    conditions[groupIndex].conditions.splice(condIndex, 1);
    if (conditions[groupIndex].conditions.length === 0) {
      conditions.splice(groupIndex, 1);
    }
  }

  function removeGroup(groupIndex: number): void {
    conditions.splice(groupIndex, 1);
  }
</script>

<div class="conditions-builder">
  <p class="mb-4 text-sm text-gray-600">
    {t("fields.conditions.description") ||
      "Define when this field is shown, hidden, required or disabled based on other fields."}
  </p>

  {#if !conditions.length}
    <div class="rounded-xl border-2 border-dashed border-gray-200 bg-gray-50/50 py-10 text-center">
      <p class="mb-3 text-sm text-gray-500">{t("fields.conditions.empty") || "No rules yet."}</p>
      <button
        type="button"
        class="inline-flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700"
        onclick={addGroup}
      >
        <SvgIcon name="plus" class="h-4 w-4" />
        {t("fields.conditions.addGroup") || "+ Add rule group"}
      </button>
    </div>
  {:else}
    <div class="space-y-4">
      {#each conditions as group, groupIndex (groupIndex)}
        <div class="rule-group overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm">
          <div class="border-b border-gray-100 bg-gray-50/80 px-4 py-3">
            <div class="flex flex-wrap items-center gap-2">
              <span class="text-xs font-medium tracking-wide text-gray-500 uppercase">
                {t("fields.conditions.rule") || "Rule"}
                {groupIndex + 1}
              </span>
              <span class="text-gray-400">·</span>
              <Select bind:value={group.logic} class="h-9 w-[5.5rem] shrink-0 text-sm">
                <option value="AND">{t("fields.conditions.and")}</option>
                <option value="OR">{t("fields.conditions.or")}</option>
              </Select>
              <span class="text-sm text-gray-500">{t("fields.conditions.then") || "then"}</span>
              <Select bind:value={group.action} class="h-9 min-w-[8.5rem] shrink-0 text-sm">
                <option value="show">{t("fields.conditions.show")}</option>
                <option value="hide">{t("fields.conditions.hide")}</option>
                <option value="require">{t("fields.conditions.require")}</option>
                <option value="disable">{t("fields.conditions.disable")}</option>
              </Select>
              <span class="text-sm text-gray-500">{t("fields.conditions.thisField") || "this field"}</span>
            </div>
          </div>

          <div class="p-4">
            <div class="space-y-3">
              {#each group.conditions as condition, condIndex (condIndex)}
                <div
                  class="grid grid-cols-[minmax(11rem,1fr)_minmax(9rem,max-content)_minmax(7rem,1fr)_auto] items-center gap-3"
                >
                  <Select bind:value={condition.field_id} class="min-w-0 text-sm">
                    <option value="">{t("fields.conditions.selectField")}</option>
                    {#each availableFields as f (f.id)}
                      <option value={f.id}>{f.displayName ?? f.name ?? f.id}</option>
                    {/each}
                  </Select>
                  <Select bind:value={condition.operator} class="min-w-[9rem] shrink-0 text-sm">
                    <option value="equals">{t("fields.conditions.equals")}</option>
                    <option value="not_equals">{t("fields.conditions.notEquals")}</option>
                    <option value="contains">{t("fields.conditions.contains")}</option>
                    <option value="not_contains">{t("fields.conditions.notContains")}</option>
                    <option value="greater_than">{t("fields.conditions.greaterThan")}</option>
                    <option value="less_than">{t("fields.conditions.lessThan")}</option>
                    <option value="is_empty">{t("fields.conditions.isEmpty")}</option>
                    <option value="is_not_empty">{t("fields.conditions.isNotEmpty")}</option>
                  </Select>
                  {#if condition.operator !== "is_empty" && condition.operator !== "is_not_empty"}
                    <Input
                      bind:value={condition.value}
                      placeholder={t("fields.conditions.value")}
                      class="min-w-0 text-sm"
                    />
                  {:else}
                    <div></div>
                  {/if}
                  <button
                    type="button"
                    class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md text-gray-400 hover:bg-red-50 hover:text-red-600"
                    aria-label={t("common.delete")}
                    onclick={() => removeCondition(groupIndex, condIndex)}
                  >
                    <SvgIcon name="trash-x" class="h-4 w-4" />
                  </button>
                </div>
              {/each}
            </div>

            <div class="mt-3 flex flex-wrap items-center justify-between gap-2 border-t border-gray-100 pt-3">
              <button
                type="button"
                class="inline-flex items-center gap-1.5 rounded-md border border-gray-200 bg-white px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-50"
                onclick={() => addCondition(groupIndex)}
              >
                <SvgIcon name="plus" class="h-3.5 w-3.5" />
                {t("fields.conditions.addCondition")}
              </button>
              <button
                type="button"
                class="inline-flex items-center gap-1.5 rounded-md border border-transparent px-3 py-1.5 text-sm text-red-600 hover:bg-red-50"
                onclick={() => removeGroup(groupIndex)}
              >
                <SvgIcon name="trash-x" class="h-3.5 w-3.5" />
                {t("common.delete")}
              </button>
            </div>
          </div>
        </div>
      {/each}

      <button
        type="button"
        class="w-full rounded-xl border-2 border-dashed border-gray-200 py-3 text-sm font-medium text-gray-600 hover:border-indigo-300 hover:bg-indigo-50/50 hover:text-indigo-700"
        onclick={addGroup}
      >
        + {t("fields.conditions.addGroup")}
      </button>
    </div>
  {/if}
</div>
