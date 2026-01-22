# White-Label Branding

**Last Updated**: 2026-01-21 00:00 UTC

## Overview

Customize the appearance of goSign with your company branding, including logos, colors, fonts, and custom domains.

## Branding Settings

### Basic Branding
- **Company Name** - Display name throughout the platform
- **Logo** - Company logo (PNG, JPEG, SVG)
- **Favicon** - Browser tab icon

### Colors
- **Primary Color** - Main brand color (#4F46E5)
- **Secondary Color** - Secondary brand color (#6366F1)
- **Accent Color** - Accent color for highlights (#10B981)
- **Background Color** - Page background (#FFFFFF)
- **Text Color** - Primary text color (#111827)

### Typography
- **Font Family** - Custom font (e.g., 'Inter', 'Roboto')
- **Font URL** - Google Fonts or custom font URL

### Signing Page Themes
- **default** - Standard signing page layout
- **minimal** - Minimal design with less visual elements
- **corporate** - Corporate theme with prominent branding

### Email Templates
- **Email Header** - Custom header image for emails
- **Email Footer** - Custom footer text
- **Email Theme** - Email template style

### Advanced
- **Custom CSS** - Advanced styling with custom CSS
- **Hide "Powered by"** - Remove goSign branding (Pro feature)
- **Custom Domain** - Use your own domain (Enterprise feature)

## API Endpoints

### Get Branding
```bash
GET /api/v1/branding
```

### Update Branding
```bash
PUT /api/v1/branding
Content-Type: application/json

{
  "branding": {
    "company_name": "Acme Corp",
    "primary_color": "#4F46E5",
    "logo_url": "https://example.com/logo.png",
    "signing_page_theme": "corporate",
    "show_powered_by": false
  }
}
```

### Upload Asset
```bash
POST /api/v1/branding/assets
Content-Type: multipart/form-data

type=logo
file=<image file>
```

## Custom Domain

### Setup Process
1. Add custom domain via API
2. Receive verification token
3. Add DNS TXT record with token
4. System verifies domain
5. SSL certificate automatically provisioned

### API
```bash
POST /api/v1/branding/domain
Content-Type: application/json

{
  "domain": "sign.example.com"
}
```

## CSS Variables

Branding colors are applied as CSS variables:
- `--color-primary`
- `--color-secondary`
- `--color-accent`
- `--color-background`
- `--color-text`
- `--font-family`

## Best Practices

- Use high-quality logo images (SVG preferred)
- Ensure color contrast meets accessibility standards
- Test branding on different devices
- Keep custom CSS minimal and maintainable
- Use web-safe fonts or provide fallbacks
- Optimize image file sizes
