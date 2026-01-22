# Formula Engine

**Last Updated**: 2026-01-21 00:00 UTC

## Overview

Formulas allow you to create calculated fields that automatically compute values based on other field values.

## Syntax

Formulas use field IDs as variables and support standard arithmetic operations.

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

1. Open field editor
2. Click "Formula" tab
3. Enter formula or use builder
4. Click field names to insert
5. Click functions to insert syntax
6. Preview result in real-time
7. Validate before saving

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

## Best Practices

- Use descriptive field IDs
- Test formulas with sample data
- Handle edge cases (empty values, zero division)
- Keep formulas readable
- Document complex calculations
- Use parentheses for clarity
