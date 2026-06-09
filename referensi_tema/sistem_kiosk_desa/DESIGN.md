---
name: Sistem Kiosk Desa
colors:
  surface: '#faf8ff'
  surface-dim: '#dad9e1'
  surface-bright: '#faf8ff'
  surface-container-lowest: '#ffffff'
  surface-container-low: '#f4f3fa'
  surface-container: '#eeedf4'
  surface-container-high: '#e9e7ef'
  surface-container-highest: '#e3e1e9'
  on-surface: '#1a1b21'
  on-surface-variant: '#444651'
  inverse-surface: '#2f3036'
  inverse-on-surface: '#f1f0f7'
  outline: '#757682'
  outline-variant: '#c5c5d3'
  surface-tint: '#4059aa'
  primary: '#00236f'
  on-primary: '#ffffff'
  primary-container: '#1e3a8a'
  on-primary-container: '#90a8ff'
  inverse-primary: '#b6c4ff'
  secondary: '#505f76'
  on-secondary: '#ffffff'
  secondary-container: '#d0e1fb'
  on-secondary-container: '#54647a'
  tertiary: '#00311f'
  on-tertiary: '#ffffff'
  tertiary-container: '#004a31'
  on-tertiary-container: '#27c38a'
  error: '#ba1a1a'
  on-error: '#ffffff'
  error-container: '#ffdad6'
  on-error-container: '#93000a'
  primary-fixed: '#dce1ff'
  primary-fixed-dim: '#b6c4ff'
  on-primary-fixed: '#00164e'
  on-primary-fixed-variant: '#264191'
  secondary-fixed: '#d3e4fe'
  secondary-fixed-dim: '#b7c8e1'
  on-secondary-fixed: '#0b1c30'
  on-secondary-fixed-variant: '#38485d'
  tertiary-fixed: '#6ffbbe'
  tertiary-fixed-dim: '#4edea3'
  on-tertiary-fixed: '#002113'
  on-tertiary-fixed-variant: '#005236'
  background: '#faf8ff'
  on-background: '#1a1b21'
  surface-variant: '#e3e1e9'
typography:
  display-kiosk:
    fontFamily: Inter
    fontSize: 48px
    fontWeight: '700'
    lineHeight: 56px
    letterSpacing: -0.02em
  headline-lg:
    fontFamily: Inter
    fontSize: 32px
    fontWeight: '600'
    lineHeight: 40px
  headline-md:
    fontFamily: Inter
    fontSize: 24px
    fontWeight: '600'
    lineHeight: 32px
  body-lg:
    fontFamily: Inter
    fontSize: 18px
    fontWeight: '400'
    lineHeight: 28px
  body-md:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: '400'
    lineHeight: 24px
  label-lg:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: '600'
    lineHeight: 20px
    letterSpacing: 0.05em
  kiosk-button:
    fontFamily: Inter
    fontSize: 20px
    fontWeight: '600'
    lineHeight: 24px
rounded:
  sm: 0.25rem
  DEFAULT: 0.5rem
  md: 0.75rem
  lg: 1rem
  xl: 1.5rem
  full: 9999px
spacing:
  unit: 8px
  kiosk-gutter: 24px
  admin-gutter: 16px
  touch-target-min: 48px
  container-padding-mobile: 16px
  container-padding-desktop: 32px
---

## Brand & Style
The design system focuses on two distinct but harmonious environments: a high-density administrative dashboard and a low-density, high-accessibility kiosk interface. The brand personality is **Trustworthy, Institutional, and Empowering**, bridging the gap between professional government administration and intuitive public service.

The design style is **Corporate / Modern** with a strong emphasis on **Accessibility**. It utilizes a clean, systematic approach with high-contrast elements and ample touch targets. The administrative side leans toward structural efficiency, while the kiosk side adopts a more tactile, "finger-friendly" layout to ensure residents of all technical backgrounds can navigate the system with confidence.

## Colors
The palette is rooted in stability and clarity.
- **Primary (Deep Blue):** Used for headers, primary navigation, and institutional branding to convey authority and trust.
- **Secondary (Slate):** Utilized for supporting text, icons, and secondary actions to maintain a neutral, professional tone.
- **Success (Green):** Specifically reserved for primary "Call to Action" buttons on the Kiosk to guide users toward completion.
- **Warning (Yellow):** Used for alerts and status indicators requiring attention without causing alarm.

Color contrast ratios must adhere to WCAG 2.1 AA standards, particularly for the Kiosk interface where readability in varied lighting conditions is critical.

## Typography
This design system employs **Inter** for its exceptional legibility and systematic weight distribution. 

### Scale Strategy
- **Admin Dashboard:** Uses `body-md` and `label-lg` for data-heavy tables and forms to maximize information density without sacrificing clarity.
- **Kiosk Interface:** Exclusively uses `display-kiosk`, `headline-lg`, and `kiosk-button`. Text must be large enough to be read from a standing distance.
- **Hierarchy:** Use semi-bold (600) for interactive labels and bold (700) for page titles to ensure a clear path of navigation.

## Layout & Spacing
The system uses an 8px grid-based spacing model.

### Kiosk Layout
- **Model:** Fixed-center or Fluid with heavy margins.
- **Philosophy:** High-padding, single-column or simple 2-column layouts to prevent cognitive overload.
- **Touch-Targets:** All interactive elements must maintain a minimum height of 48px, with 12px-16px of clearance between buttons to prevent accidental taps.

### Admin Dashboard
- **Model:** 12-column Fluid Grid with a persistent left sidebar.
- **Philosophy:** Efficient use of horizontal space for data tables and multi-input forms. 
- **Breakpoints:** 
  - Mobile (< 640px): Full-width cards, hidden sidebar.
  - Tablet (640px - 1024px): Condensed sidebar, 2-column grid.
  - Desktop (> 1024px): Expanded sidebar, full 12-column flexibility.

## Elevation & Depth
The system utilizes **Tonal Layers** combined with **Low-Contrast Outlines** to maintain a professional, government-standard appearance.

- **Cards:** Use a 1px border (`#E2E8F0`) with a very soft, ambient shadow (Y: 2px, Blur: 4px, Opacity: 0.05) to separate content from the light gray background.
- **Kiosk Buttons:** Slightly higher elevation (Y: 4px, Blur: 8px) to provide a "pressable" tactile feel, indicating clear interactivity.
- **Modals:** Use a heavy backdrop blur (12px) and a centralized elevation to focus the user's attention entirely on the task at hand, especially critical in public kiosk environments.

## Shapes
The design system uses a **Rounded** (0.5rem) corner radius as the standard. This strikes a balance between the serious nature of government work and the approachability required for a public kiosk. 

- **Standard Elements:** 8px (0.5rem) for input fields, small buttons, and cards.
- **Kiosk Action Buttons:** 16px (1rem) to emphasize their friendly, touch-ready nature.
- **Input Fields:** Consistent 8px rounding to match the professional tone of the admin dashboard.

## Components
### Buttons
- **Kiosk Primary:** Large, height 60px+, Green background, White text.
- **Admin Primary:** Height 40px, Deep Blue background, White text.
- **States:** Hover should darken the background by 10%; Active/Pressed should scale slightly (0.98) to provide tactile feedback on the kiosk.

### Cards
- **Admin:** Minimal padding (16px), flat borders, used for data containers.
- **Kiosk:** Large padding (32px), higher contrast, used for navigation "tiles" with large icons.

### Input Fields
- **Administrative:** Standard text inputs with clear 1px borders and `label-lg` titles.
- **Kiosk Numpad:** A custom component consisting of 80x80px grid buttons for numeric entry (NIK, Phone Number), ensuring high hit-success rates for elderly or non-technical users.

### Lists
- **Data Tables (Admin):** Alternating row colors (Zebra striping) using `#F8FAFC` for better horizontal scanning of resident data.
- **Option Lists (Kiosk):** Large list items (height 72px+) with chevron icons to indicate "drill-down" navigation.