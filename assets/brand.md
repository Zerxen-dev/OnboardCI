# OnboardCI Brand Identity Guidelines

This document specifies the visual identity, design geometry, color palette, typography, clear-space rules, and usage policies for the OnboardCI open-source project.

---

## 1. Concept & Geometric Construction

The OnboardCI mark embodies a **checkpoint gate**:
* **Container Walls**: Two sturdy vertical framing bars symbolizing the isolated, sandboxed environment (Docker container boundaries).
* **The Threshold / Checkmark**: A gateway entrance threshold that seamlessly curves into an ascending verification checkmark.
* **Core Metaphor**: "Verify at the gate" — every doc command and environment variable is validated before code enters production or developer setups.

### Grid & Dimensions
* **Canvas Grid**: 512 × 512 px
* **Safe Margin**: 64 px perimeter padding (artwork remains within 384 × 384 px active area)
* **Corner Radius**: 8–12 px rounded corners at 512 px scale (4 px equivalent relative scale)
* **Geometry**: Pure geometric vectors constructed with rounded rectangles, circles, and straight lines. No gradients, no dropshadows, no raster dependencies.

---

## 2. Color Palette

The official color palette is defined below. These hex codes must be preserved across all implementations.

| Token | Hex Value | Role | Usage |
|---|---|---|---|
| **Deep Ink** | `#0B1220` | Primary Dark Canvas | Backgrounds, dark mode canvases, terminal contrast base |
| **Signal Green** | `#22C55E` | Verification / Pass | Checkpoint gate threshold, verified checks, primary accents |
| **Warning Amber** | `#F59E0B` | Warning / Fail | Failed doc step alerts, missing env flag warnings |
| **Paper White** | `#F8FAFC` | Light Canvas | Light mode canvas, documentation backgrounds, mono mark fill |
| **Slate** | `#64748B` | Secondary Detail | Structural borders, container walls, secondary text |

---

## 3. Typography Stack

For wordmarks, promotional materials, and documentation:

* **Primary Wordmark & Headings**:
  ```css
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  font-weight: 700;
  letter-spacing: -0.03em;
  ```
* **Code & CLI Diagnostics**:
  ```css
  font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace;
  font-weight: 500;
  ```

---

## 4. Clear Space & Scaling Rules

* **Minimum Clear Space**: Maintain a buffer equal to the width of the vertical framing bar (64 px at 512 scale, or 12.5% of mark width) on all sides.
* **Minimum Display Size**:
  * Scaled Mark (`logo.svg`): Minimum 32 × 32 px.
  * Favicon (`favicon.svg`): Optimized specifically for 32 × 32 px and 16 × 16 px rendering.
* **High-Contrast Monochromatic Mark (`logo-mono.svg`)**:
  * Designed to maintain high contrast on both `#0B1220` (Dark) and `#F8FAFC` (Light) backgrounds without color degradation.

---

## 5. Usage Rules: Do's and Don'ts

### Do:
- Use `assets/logo.svg` on dark backgrounds or centered within white documentation pages.
- Use `assets/logo-mono.svg` for single-color print, monochrome badges, or embedding in dark/light terminals.
- Maintain the original aspect ratio and geometry.
- Keep the 2-tone simplicity.

### Don't:
- Do not apply drop shadows, 3D extrusions, or multi-stop gradients.
- Do not rotate the checkpoint mark or skew the check angle.
- Do not add text or extra icons inside the primary 512×512 mark.
- Do not alter the green (`#22C55E`) or deep ink (`#0B1220`) shades.

---

## 6. GitHub Action Branding

GitHub Actions `action.yml` restricts `branding.icon` to a predefined set of GitHub Octicons (custom SVG paths cannot be rendered in GitHub Marketplace badges directly).
* **Action Icon**: `check-circle`
* **Action Color**: `green`
* The official SVG mark remains `assets/logo.svg` across all repositories and documentation headers.
