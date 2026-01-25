# Conditional Fields

**Last Updated**: 2026-01-25

## Overview

Conditional fields allow you to show, hide, require, or disable fields based on values of other fields using AND/OR logic.

## Supported Operators

- **equals** - Field value equals specified value
- **not_equals** - Field value does not equal specified value
- **contains** - Field value contains specified text (for text fields)
- **not_contains** - Field value does not contain specified text
- **greater_than** - Field value is greater than specified number
- **less_than** - Field value is less than specified number
- **is_empty** - Field is empty or not filled
- **is_not_empty** - Field has a value

## Supported Actions

- **show** - Show field when condition is met
- **hide** - Hide field when condition is not met
- **require** - Make field required when condition is met
- **disable** - Disable field when condition is met

## Logic Operators

- **AND** - All conditions in group must be true
- **OR** - At least one condition in group must be true

## Examples

### Simple Show/Hide
Show "Company Name" field only when "Account Type" equals "Business":

```json
{
  "field_id": "company_name",
  "condition_groups": [{
    "logic": "AND",
    "conditions": [{
      "field_id": "account_type",
      "operator": "equals",
      "value": "Business"
    }],
    "action": "show"
  }]
}
```

### Multiple Conditions (AND)
Require "Tax ID" when account type is "Business" AND country is "US":

```json
{
  "field_id": "tax_id",
  "condition_groups": [{
    "logic": "AND",
    "conditions": [
      {
        "field_id": "account_type",
        "operator": "equals",
        "value": "Business"
      },
      {
        "field_id": "country",
        "operator": "equals",
        "value": "US"
      }
    ],
    "action": "require"
  }]
}
```

### Multiple Conditions (OR)
Show "Discount Code" when total is greater than 1000 OR user type is "Premium":

```json
{
  "field_id": "discount_code",
  "condition_groups": [{
    "logic": "OR",
    "conditions": [
      {
        "field_id": "total",
        "operator": "greater_than",
        "value": 1000
      },
      {
        "field_id": "user_type",
        "operator": "equals",
        "value": "Premium"
      }
    ],
    "action": "show"
  }]
}
```

## Using the Condition Builder

1. Open field editor (gear icon on a field).
2. Click **Conditional Logic** in the dropdown.
3. Add a rule group; set logic (AND/OR), then action (show/hide/require/disable) for "this field".
4. In each row: select field (only fields of the **active submitter** are listed), operator, and value.
5. Use **Add Condition** for more rows; use **Delete** at the bottom to remove the group.
6. Field selectors show display names (e.g. "First Number 1"); condition groups use a compact grid layout.
7. UI supports multiple locales (description, empty state, rule labels).

## Validation

The system validates:
- Referenced fields exist
- No circular dependencies
- Operators are valid for field types
- Conditions are logically sound

## API

### Validate Conditions
```bash
POST /api/v1/templates/:id/conditions/validate
Content-Type: application/json

{
  "fields": [...]
}
```

## Best Practices

- Keep conditions simple and clear
- Test all condition combinations
- Avoid deep nesting of conditions
- Use descriptive field names
- Document complex conditional logic
