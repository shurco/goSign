# Web2 Pages And Routes

Short documentation for the new Svelte-based frontend in `web2/`.

## Key Notes

- The application runs as a SvelteKit SPA.
- If a page requires authentication and no token is present, the user is redirected to:
  - `/auth/signin?redirect=<original_path>`
- If the user is already authenticated, these pages automatically redirect to `/dashboard`:
  - `/auth/signin`
  - `/auth/signup`
- Admin-only pages are available only to users with `role = 3`.
- Layout groups:
  - `(main)` - public standard pages
  - `(blank)` - auth, signing, and standalone utility screens
  - `(sidebar)` - main authenticated app with the left navigation
  - `(settings)` - settings area with the main sidebar plus settings sidebar

## How To Reach Sections From The UI

Main left navigation for authenticated users:

- `Dashboard` -> `/dashboard`
- `Submissions` -> `/submissions`
- `Templates` -> `/templates`
- `Settings` -> `/settings` -> redirects to `/settings/general`

Additional admin navigation:

- `Organizations` -> `/admin/organizations`
- `Settings` -> `/admin/settings` -> redirects to `/admin/settings/smtp`

Tabs inside the settings area:

- user settings:
  - `/settings/general`
  - `/settings/email/templates`
  - `/settings/webhooks`
  - `/settings/api-keys`
  - `/settings/branding`
- admin settings:
  - `/admin/settings/smtp`
  - `/admin/settings/sms`
  - `/admin/settings/storage`
  - `/admin/settings/geolocation`

## Public Pages

| URL | Page | Access | How To Reach It |
| --- | --- | --- | --- |
| `/` | Home | Public | Open the site root |
| `/verify` | Verify | Public | Direct URL |
| `/templates/[id]/view` | View template | Public | Direct URL or from the template library |
| `/s/[slug]` | Submitter signing | Public | Public signing link |
| `/auth/signin` | Sign in | Guests only | Direct URL or automatic redirect when opening a protected page |
| `/auth/signup` | Sign up | Guests only | Direct URL |
| `/auth/password/forgot` | Forgot password | Public | From the sign-in page or direct URL |
| `/auth/password/reset` | Reset password | Public | Usually from an email link |
| `/auth/verify-email` | Verify email | Public | Usually from an email link |
| `/[...notfound]` | 404 | Public | Any unknown URL |

## Authenticated User Pages

| URL | Page | Access | How To Reach It |
| --- | --- | --- | --- |
| `/dashboard` | Dashboard | Login required | From the left navigation |
| `/submissions` | Submissions | Login required | From the left navigation |
| `/submissions/[submission_id]/status` | Submission status | Login required | From submissions or dashboard |
| `/templates` | Template library | Login required | From the left navigation |
| `/templates/[id]/edit` | Edit template | Login required | From the template list |
| `/templates/[id]/folder` | Template folder view | Login required | From the template library when opening a folder |
| `/settings` | Settings root | Login required | From the left navigation; redirects to `/settings/general` |
| `/settings/general` | General settings | Login required | From `Settings` |
| `/settings/email/templates` | Organization email templates | Login required | From `Settings` |
| `/settings/webhooks` | Webhooks | Login required | From `Settings` |
| `/settings/api-keys` | API keys | Login required | From `Settings` |
| `/settings/branding` | Branding | Login required | From `Settings` |

## Admin Pages

| URL | Page | Access | How To Reach It |
| --- | --- | --- | --- |
| `/organizations` | Legacy redirect | Admin only | Redirects to `/admin/organizations` |
| `/organizations/[organization_id]/members` | Legacy redirect | Admin only | Redirects to `/admin/organizations/[organization_id]/members` |
| `/admin/organizations` | Organizations | Admin only | From the admin left navigation |
| `/admin/organizations/[organization_id]/members` | Organization members | Admin only | From the organizations list |
| `/admin/settings` | Admin settings root | Admin only | From the admin left navigation; redirects to `/admin/settings/smtp` |
| `/admin/settings/smtp` | SMTP settings | Admin only | From admin settings tabs |
| `/admin/settings/sms` | SMS settings | Admin only | From admin settings tabs |
| `/admin/settings/storage` | Storage settings | Admin only | From admin settings tabs |
| `/admin/settings/geolocation` | Geolocation settings | Admin only | From admin settings tabs |
| `/admin/settings/email/templates` | Compatibility redirect | Admin only | Redirects to `/settings/email/templates` |

## Dynamic URLs

The following routes require real IDs or slugs:

- `/templates/[id]/view`
  - example: `/templates/42/view`
- `/templates/[id]/edit`
  - example: `/templates/42/edit`
- `/templates/[id]/folder`
  - example: `/templates/folder-123/folder`
- `/submissions/[submission_id]/status`
  - example: `/submissions/88/status`
- `/admin/organizations/[organization_id]/members`
  - example: `/admin/organizations/org-1/members`
- `/organizations/[organization_id]/members`
  - redirect-only URL
- `/s/[slug]`
  - example: `/s/2f4a8c-public-sign-link`

## Redirect Behavior

- Unauthenticated user opening any page inside `(sidebar)` or `(settings)`:
  - redirect to `/auth/signin?redirect=<requested_path>`
- Authenticated user opening `/auth/signin` or `/auth/signup`:
  - redirect to `/dashboard`
- `/settings`:
  - redirect to `/settings/general`
- `/admin/settings`:
  - redirect to `/admin/settings/smtp`
- `/organizations`:
  - redirect to `/admin/organizations`
- `/organizations/[organization_id]/members`:
  - redirect to `/admin/organizations/[organization_id]/members`
- `/admin/settings/email/templates`:
  - redirect to `/settings/email/templates`

## Quick Links For Manual Testing

- Public:
  - `/`
  - `/verify`
  - `/auth/signin`
  - `/auth/signup`
  - `/auth/password/forgot`
- After login:
  - `/dashboard`
  - `/submissions`
  - `/templates`
  - `/settings/general`
- As administrator:
  - `/admin/organizations`
  - `/admin/settings/smtp`
  - `/admin/settings/storage`
