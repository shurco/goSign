# Formula Engine

**Last Updated**: 2026-01-25

## Overview

Formulas allow you to create calculated fields that automatically compute values based on other field values.

## Syntax

Formulas use field IDs as variables (including UUIDs) and support standard arithmetic operations. The formula builder UI shows field **display names** in the expression (e.g. `[[First Number 1]]`) while storing and evaluating using field IDs; validation and evaluation handle UUIDs correctly.

### Basic Operations
- Addition: `field_1 + field_2`
- Subtraction: `field_1 - field_2`
- Multiplication: `field_1 * field_2`
- Division: `field_1 / field_2`
- Parentheses: `(field_1 + field_2) * 0.2`

### Built-in Functions

#### SUM
Sum multiple fields:
```
SUM(field_1, field_2, field_3)
```

#### IF
Conditional calculation:
```
IF(field_1 > 100, field_2, 0)
```

#### MAX
Maximum value:
```
MAX(field_1, field_2, field_3)
```

#### MIN
Minimum value:
```
MIN(field_1, field_2, field_3)
```

#### ROUND
Round to decimal places:
```
ROUND(field_1, 2)
```

## Examples

### Calculate Total
```
field_1 + field_2 + field_3
```

### Calculate Tax (20%)
```
field_1 * 1.2
```

### Conditional Discount
If total > 1000, apply 10% discount:
```
IF(field_1 > 1000, field_1 * 0.9, field_1)
```

### Sum with Tax
```
SUM(field_1, field_2, field_3) * 1.2
```

### Complex Calculation
```
(SUM(field_1, field_2) - field_3) * 0.15 + field_4
```

## Field Types

Formulas work with:
- Number fields
- Payment fields
- Text fields (if numeric)

## Calculation Types

- **number** - Standard number format
- **currency** - Currency format with $ symbol

## Using the Formula Builder

1. Open field editor (gear icon on a number or text field).
2. Click **Formula** in the dropdown.
3. Enter the formula in the text area; the UI shows field names as `[[Field Name]]` (stored as field IDs).
4. Click **Insert Field** buttons to add references (only number/text fields of the active submitter).
5. Click **Functions** to insert SUM, IF, MAX, MIN, ROUND.
6. Use **Examples** to paste sample formulas; preview and validation run as you type.
7. Save to apply the formula to the field.

## Validation

The system validates:
- Formula syntax is correct
- All referenced fields exist
- Field types are compatible
- No division by zero
- Result is a valid number

## API

### Validate Formula
```bash
POST /api/v1/templates/formulas/validate
Content-Type: application/json

{
  "formula": "field_1 + field_2",
  "fields": [...]
}
```

## Technical Notes

- **Field IDs**: Templates may use UUIDs for field IDs. The backend rewrites UUIDs to safe identifiers before compiling/evaluating so that the expression parser does not treat hyphens as minus operators.
- **Validation**: `POST /api/v1/templates/formulas/validate` checks syntax and that all referenced fields exist; non-UUID identifiers (e.g. `field_1`) and full UUIDs are both supported.

## Best Practices

- Use descriptive field names so the formula builder display (`[[Name]]`) is clear.
- Test formulas with sample data.
- Handle edge cases (empty values, zero division).
- Keep formulas readable; use parentheses for clarity.
