import type { Fields, Areas, Submitters } from "./index";

/**
 * Extended Area interface with frontend-specific properties
 */
export interface Area extends Omit<Areas, "z"> {
  h: number; // height (replaces z from backend)
  cell_w?: number; // cell width for cells field type
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
 * Extended Field interface with frontend-specific properties
 */
export interface Field extends Omit<Fields, "areas"> {
  readonly?: boolean;
  default_value?: string;
  areas?: Area[];
  options?: FieldOption[];
  preferences?: {
    format?: string; // date format
    price?: number; // payment price
    currency?: string; // payment currency
    [key: string]: any;
  };
}

/**
 * Submitter with extended properties
 */
export type Submitter = Submitters;
