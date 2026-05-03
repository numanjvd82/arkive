---
name: Arkive Core
colors:
  surface: '#0f1418'
  surface-dim: '#0f1418'
  surface-bright: '#343a3e'
  surface-container-lowest: '#0a0f12'
  surface-container-low: '#171c20'
  surface-container: '#1b2024'
  surface-container-high: '#252b2e'
  surface-container-highest: '#303539'
  on-surface: '#dee3e8'
  on-surface-variant: '#bdc8d1'
  inverse-surface: '#dee3e8'
  inverse-on-surface: '#2c3135'
  outline: '#87929a'
  outline-variant: '#3e484f'
  surface-tint: '#7bd0ff'
  primary: '#8ed5ff'
  on-primary: '#00354a'
  primary-container: '#38bdf8'
  on-primary-container: '#004965'
  inverse-primary: '#00668a'
  secondary: '#b9c8de'
  on-secondary: '#233143'
  secondary-container: '#39485a'
  on-secondary-container: '#a7b6cc'
  tertiary: '#ffc176'
  on-tertiary: '#472a00'
  tertiary-container: '#f1a02b'
  on-tertiary-container: '#613b00'
  error: '#ffb4ab'
  on-error: '#690005'
  error-container: '#93000a'
  on-error-container: '#ffdad6'
  primary-fixed: '#c4e7ff'
  primary-fixed-dim: '#7bd0ff'
  on-primary-fixed: '#001e2c'
  on-primary-fixed-variant: '#004c69'
  secondary-fixed: '#d4e4fa'
  secondary-fixed-dim: '#b9c8de'
  on-secondary-fixed: '#0d1c2d'
  on-secondary-fixed-variant: '#39485a'
  tertiary-fixed: '#ffddb8'
  tertiary-fixed-dim: '#ffb960'
  on-tertiary-fixed: '#2a1700'
  on-tertiary-fixed-variant: '#653e00'
  background: '#0f1418'
  on-background: '#dee3e8'
  surface-variant: '#303539'
typography:
  h1:
    fontFamily: Inter
    fontSize: 24px
    fontWeight: '600'
    lineHeight: 32px
    letterSpacing: -0.02em
  h2:
    fontFamily: Inter
    fontSize: 18px
    fontWeight: '600'
    lineHeight: 24px
    letterSpacing: -0.01em
  body-md:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: '400'
    lineHeight: 20px
  body-sm:
    fontFamily: Inter
    fontSize: 13px
    fontWeight: '400'
    lineHeight: 18px
  code-tabular:
    fontFamily: Inter
    fontSize: 13px
    fontWeight: '400'
    lineHeight: 18px
  label-caps:
    fontFamily: Inter
    fontSize: 11px
    fontWeight: '600'
    lineHeight: 16px
    letterSpacing: 0.05em
rounded:
  sm: 0.125rem
  DEFAULT: 0.25rem
  md: 0.375rem
  lg: 0.5rem
  xl: 0.75rem
  full: 9999px
spacing:
  base: 4px
  xs: 4px
  sm: 8px
  md: 16px
  lg: 24px
  xl: 40px
  container-max: 1440px
  sidebar-width: 240px
---

## Brand & Style

The design system is engineered for the "Sovereign User"—individuals who prioritize data ownership and technical autonomy. The brand personality is clinical, resilient, and precise. It avoids the friendly, rounded aesthetic of consumer-grade cloud storage in favor of a "Utility-First" approach that mirrors a professional IDE or an advanced administrative dashboard.

The visual direction follows a **Technical Minimalism** style. It prioritizes information density and operational speed over decorative elements. By utilizing a dark-first palette and low-contrast borders, the interface recedes into the background, allowing the user's data and file structures to remain the primary focus. The emotional goal is to evoke a sense of impenetrable security and absolute system transparency.

## Colors

The color system is built on a "Deep Night" foundation to reduce eye strain during long periods of technical management. 

- **Primary Canvas:** The background utilizes a very deep charcoal (#0B0E11) rather than pure black to maintain visible depth when UI elements overlap.
- **Surface Tiers:** Nested containers and sidebars use a slightly lighter slate-gray (#161B22) to establish hierarchy without the need for shadows.
- **Accents:** The "Security Blue" (#38BDF8) is used sparingly for active states, primary actions, and encryption indicators. 
- **Status Tones:** Warm slate and muted emeralds are used for system health and transfer status, ensuring the interface remains calm and non-alarmist.

## Typography

This design system utilizes **Inter** for its exceptional legibility and neutral character. The typography strategy is centered on "Information Scanability."

A critical requirement for this system is the use of **Tabular Figures** (`tnum`) for all file sizes, timestamps, and hash strings. This ensures that columns of data align perfectly in list views, allowing users to compare file sizes or dates at a glance without visual staggering. 

Hierarchy is established through weight and color (shifting from Primary Text to Secondary Text) rather than dramatic jumps in font size. Headings remain compact to preserve vertical screen real estate.

## Layout & Spacing

The layout employs a **Fixed-Fluid Hybrid** model typical of administrative tools. A fixed-width sidebar (240px) houses global navigation, while the main content area utilizes a fluid grid that expands to fill the viewport, capped at 1440px for readability.

We use a strict **4px baseline grid**. Standard padding for components is 8px (sm) or 12px (between sm and md) to maintain a compact, "pro" density. High-density data tables should use 8px vertical cell padding. This allows more files to be visible on a single screen without scrolling, emphasizing the "Efficiency" pillar of the system.

## Elevation & Depth

In this design system, depth is communicated through **Tonal Layering** and **Low-Contrast Outlines** rather than physical light and shadow.

- **Level 0 (Main Canvas):** #0B0E11.
- **Level 1 (Navigation/Sidebar):** #161B22.
- **Level 2 (Modals/Floating Menus):** #1C2128 with a 1px border of #30363D.

Shadows are almost entirely avoided. When absolutely necessary for floating elements (like dropdowns), use a sharp, 4px blur with 40% opacity in a pure black tint to provide just enough separation from the background. Surfaces do not "float"; they are "staged" within the grid.

## Shapes

The shape language is disciplined and geometric. A uniform border-radius of **6px (Soft)** is applied to buttons, input fields, and cards. This radius is large enough to feel modern and accessible, but sharp enough to maintain a technical, professional aesthetic.

Iconography should follow a "utilitarian" style: 2px stroke weights, no filled areas unless indicating an active state, and strictly aligned to a 20px or 24px bounding box. Elements like progress bars for file uploads should have 0px or 2px radius to appear more "instrument-like."

## Components

### Buttons
Primary buttons use the Security Blue (#38BDF8) with black text for high contrast. Secondary buttons use a ghost style: 1px border (#30363D) with no fill, shifting to a subtle gray fill on hover.

### File Lists
The core component of the app. Rows should have a subtle hover state (#1C2128) and use tabular Inter for all metadata. Columns must be strictly aligned with fixed widths for "Size" and "Modified" fields.

### Inputs
Search and configuration fields use the Surface color (#161B22) as their background with a 1px border. The focus state is a simple 1px primary blue border—no outer glows or rings.

### Zero-Knowledge Indicators
A specialized component—a small lock icon or "Shield" chip—should appear next to file paths or encryption keys. These use a muted slate-blue to indicate a "protected" state without being visually distracting.

### Progress Bars
Used for file transfers. These should be thin (4px height), using a background of #30363D and a primary blue fill, positioned at the very top or bottom of a file row to minimize space usage.