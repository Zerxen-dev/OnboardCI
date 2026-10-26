# OnboardCI Brand Identity Guidelines

<p align="center">
  <img src="logo-wordmark.svg" width="380" alt="OnboardCI Logo Wordmark" />
</p>

This document specifies the visual identity, architectural geometry, color palette, typography, clear-space rules, and usage policies for the OnboardCI open-source project.

---

## 1. Concept: The Abstract Gate

The visual identity of OnboardCI moves away from literal to-do list checkboxes and generic clipboard clipart toward an **abstract architectural gateway**.

* **Metaphor & Reasoning**: "Moving from a literal checkbox to a dynamic gateway."
  * Real onboarding is not just checking boxes; it is crossing a threshold into an open-source codebase.
  * The mark constructs an open geometric monogram **O** (for *Onboard*) that simultaneously frames an architectural portal.
* **The Foundation Pillar (Left)**: A confident, solid vertical monolith representing structural stability and container isolation (Docker sandbox boundary).
* **The Chevron Canopy & Verification Threshold (Right)**: An angled entrance chevron (`>`) that sweeps downward into the gateway, paired with an ascending sharp checkmark (`✓`) clearing the contributor forward.
* **Pure Solid Geometry**: No outlines, no strokes, no gradients. Heavy, confident solid shapes engineered for crisp visual impact at any scale.

### Grid & Dimensions
* **Canvas Grid**: 512 × 512 px
* **Corner Radius**: `rx="110"` on app icon container
* **Safe Margin**: 64 px perimeter safe zone
* **Pillar Coordinates**: `x=112, y=128, w=64, h=256, rx=20`
* **Optical Center**: Centered at `(256, 256)`

---

## 2. Color Palette

The official color palette draws inspiration from high-contrast developer environments (e.g., Vercel, Linear, Supabase):

| Token | Hex Value | Role | Usage |
|---|---|---|---|
| **Deep Ink** | `#0B1220` | Canvas Base | Primary dark background, icon container, terminal contrast base |
| **Signal Green** | `#22C55E` | Verification / Pass | Gateway mark, verified step indicators, primary brand accents |
| **Paper White** | `#F8FAFC` | Typographic Lead | "Onboard" wordmark prefix, primary text, high-contrast monochrome mark |
| **Slate** | `#64748B` | Secondary Detail | Subtitles, structural borders, secondary descriptions |

---

## 3. Typography Stack

For the wordmark and official collateral, OnboardCI pairs geometric typography with deliberate font-weight contrast:

* **Wordmark Typography**:
  - `font-family`: `Inter, system-ui, -apple-system, sans-serif`
  - **"Onboard"**: `font-weight: 700` (Bold), Color: `#F8FAFC` (Paper White)
  - **"CI"**: `font-weight: 400` (Regular/Normal), Color: `#22C55E` (Signal Green)
  - Letter-spacing: `-1.5px` (Tight, modern tracking)

* **Code & CLI Diagnostics**:
  ```css
  font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace;
  font-weight: 500;
  ```

---

## 4. Official Asset Catalog

| Asset | File | Dimensions | Role |
|---|---|---|---|
| **Primary Icon Mark** | [`assets/logo.svg`](logo.svg) • [`PNG`](logo.png) | 512 × 512 px | App icon, profile avatars, release badges |
| **Primary Wordmark** | [`assets/logo-wordmark.svg`](logo-wordmark.svg) • [`PNG`](logo-wordmark.png) | 640 × 160 px | README header, hero banners, site headers |
| **Monochrome Mark** | [`assets/logo-mono.svg`](logo-mono.svg) • [`PNG`](logo-mono.png) | 512 × 512 px | Single-color white for dark backgrounds |
| **Favicon** | [`assets/favicon.svg`](favicon.svg) • [`PNG`](favicon.png) | 32 × 32 px | Browser tabs, web app shortcut icons |
| **Social Card Preview** | [`assets/social-preview.svg`](social-preview.svg) • [`PNG`](social-preview.png) | 1280 × 640 px | OpenGraph metadata, Twitter / Slack cards |

---

## 5. Usage Rules: Do's and Don'ts

### Do:
- Display `assets/logo-wordmark.svg` centered at 380–420px width in top-level documentation.
- Maintain the two-tone wordmark split (Paper White `Onboard` + Signal Green `CI`).
- Preserve the exact `#0B1220` container with `#22C55E` geometry.
- Use `assets/logo-mono.svg` when placing the icon on custom branded backgrounds.

### Don't:
- Do not revert to literal checkboxes, clipboards, or outlined square glyphs.
- Do not apply drop shadows, rainbow gradients, or glowing neon filters.
- Do not alter the proportions or skew the gateway angles.
