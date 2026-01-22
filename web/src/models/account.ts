/**
 * Frontend TypeScript interfaces for account and branding models
 * Mirrors backend Go models in internal/models/account.go
 */

export interface BrandingSettings {
  // Basic
  logo_url?: string
  favicon_url?: string
  company_name?: string
  
  // Colors
  primary_color?: string
  secondary_color?: string
  accent_color?: string
  background_color?: string
  text_color?: string
  
  // Typography
  font_family?: string
  font_url?: string
  
  // Signing Page
  signing_page_theme?: 'default' | 'minimal' | 'corporate'
  show_powered_by?: boolean
  custom_css?: string
  
  // Email Templates
  email_header_url?: string
  email_footer_text?: string
  email_theme?: 'default' | 'minimal' | 'corporate'
  
  // Custom Domain
  custom_domain?: string
  
  // Legal
  terms_url?: string
  privacy_url?: string
}

export interface BrandingAsset {
  id: string
  account_id: string
  type: 'logo' | 'favicon' | 'email_header' | 'watermark'
  file_path: string
  mime_type: string
  created_at: string
  updated_at: string
}

export interface CustomDomain {
  id: string
  account_id: string
  domain: string
  verified: boolean
  verification_token?: string
  ssl_enabled: boolean
  created_at: string
  verified_at?: string
}
