import type { Areas, Field as BaseField, Submitter } from "./index";

/**
 * Extended Area interface with frontend-specific properties
 */
export interface Area extends Omit<Areas, "z"> {
  h: number; // height (replaces z from backend)
  cell_w?: number; // cell width for cells field type
  cell_count?: number; // number of cells (persisted, used on signing)
  option_id?: string; // option id for radio/multiple field types
  initialX?: number; // initial X position for drawing
  initialY?: number; // initial Y position for drawing
}

/**
 * Field option for select, radio, and multiple field types
 */
export interface FieldOption {
  id: string;
  value: string;
}

/**
 * Condition operator types
 */
export type ConditionOperator = 
  | 'equals'
  | 'not_equals'
  | 'contains'
  | 'not_contains'
  | 'greater_than'
  | 'less_than'
  | 'is_empty'
  | 'is_not_empty'

/**
 * Condition action types
 */
export type ConditionAction = 'show' | 'hide' | 'require' | 'disable'

/**
 * Logic operator for combining conditions
 */
export type LogicOperator = 'AND' | 'OR'

/**
 * Single field condition
 */
export interface FieldCondition {
  field_id: string
  operator: ConditionOperator
  value: any
}

/**
 * Field condition group with AND/OR logic
 */
export interface FieldConditionGroup {
  logic: LogicOperator
  conditions: FieldCondition[]
  action: ConditionAction
}

/**
 * Extended Field interface with frontend-specific properties
 */
export type Field = Omit<BaseField, "areas"> & {
  readonly?: boolean;
  default_value?: string;
  label?: string;
  translations?: Record<string, string>;
  condition_groups?: FieldConditionGroup[];
  areas?: Area[];
  options?: FieldOption[];
  validation?: FieldValidation;
  preferences?: FieldPreferences;
};

export interface FieldPreferences {
  format?: string;
  align?: string;
  font?: string;
  font_type?: string;
  font_size?: number;
  valign?: string;
  color?: string;
  price?: number;
  currency?: string;
  formula?: string;
  with_logo?: boolean;
  with_signature_id?: boolean;
  reasons?: string[];
  reason_field_uuid?: string;
  method?: string;
}

export interface FieldValidation {
  pattern?: string;
  message?: string;
  min?: number;
  max?: number;
  step?: string;
}

/**
 * Submitter with extended properties
 */
export type { Submitter };
