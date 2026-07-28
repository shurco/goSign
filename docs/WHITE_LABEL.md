# White-Label Branding

**Last Updated**: 2026-07-28 00:00 UTC

## Overview

Account-level branding configuration for goSign. Branding settings are stored
per account in the `account.settings` JSONB column (under the `branding` key)
and can be managed via the API or the settings UI (**Settings → Branding**).

## Implementation Status

| Capability | Status |
|------------|--------|
| Branding settings storage (get/update via API) | Implemented |
| Company name, logo, colors in settings UI | Implemented (logo is stored as a data URL) |
| Applying branding to the signing page / emails | Planned |
| Asset upload endpoint (logo/favicon files) | Planned |
| Custom domains | Planned |

## Branding Settings

All fields are optional; omitted keys keep their stored values.

- **Basic**: `company_name`, `logo_url`, `favicon_url`
- **Colors**: `primary_color`, `secondary_color`, `accent_color`, `background_color`, `text_color`
- **Typography**: `font_family`, `font_url`
- **Signing page**: `signing_page_theme` (`default`, `minimal`, `corporate`), `show_powered_by`, `custom_css`
- **Email templates**: `email_header_url`, `email_footer_text`, `email_theme`
- **Legal**: `terms_url`, `privacy_url`

## API Endpoints

### Get Branding
```bash
GET /v1/settings/branding
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "success": true,
  "data": {
    "branding": {
      "company_name": "Acme Corp",
      "primary_color": "#4F46E5",
      "show_powered_by": true
    }
  }
}
```

### Update Branding

Provided keys are merged into the stored settings; keys that are not sent
remain unchanged.

```bash
PUT /v1/settings/branding
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "branding": {
    "company_name": "Acme Corp",
    "primary_color": "#4F46E5",
    "logo_url": "https://example.com/logo.png"
  }
}
```

## Best Practices

- Use high-quality logo images (SVG preferred)
- Ensure color contrast meets accessibility standards
- Keep custom CSS minimal and maintainable
- Use web-safe fonts or provide fallbacks
- Optimize image file sizes
