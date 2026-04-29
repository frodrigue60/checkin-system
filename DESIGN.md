# Design System Specification: The Precision Bureau

## 1. Overview & Creative North Star

The objective of this design system is to transform a standard utility—attendance management—into a high-end, authoritative experience. We are moving away from the "standard SaaS dashboard" aesthetic toward a concept we call **"The Precision Bureau."**

This North Star focuses on the intersection of architectural structure and editorial clarity. We achieve this through:

- **Intentional Asymmetry:** Avoid perfectly centered grids. Use wide, asymmetric gutters to create a sense of bespoke "stationery."

- **Tonal Authority:** Rather than relying on heavy lines, we use the sophisticated Material palette to define space through subtle shifts in atmospheric depth.

- **The Editorial Scale:** We use extreme contrast between our display typography (Public Sans) and our functional utility type (Inter) to guide the eye toward critical data points like "Time In" and "Status."

## 2. Colors & Surface Philosophy

The palette is rooted in deep corporate blues and sophisticated grays, but its application must feel "glass-like" rather than "flat."

### The "No-Line" Rule

Under no circumstances are 1px solid borders to be used for sectioning or containment. Boundaries must be defined solely through:

1.  **Background Color Shifts:** A component using `surface_container_low` (#f2f4f5) sitting on the base `surface` (#f8fafb).

2.  **Tonal Transitions:** Using `surface_container_lowest` (#ffffff) to make critical interaction areas pop against a darker `surface_dim` (#d8dadb).

### Surface Hierarchy & Nesting

Treat the UI as a series of physical layers.

- **Base:** `background` (#f8fafb).

- **Nesting:** Place `surface_container` (#eceeef) cards on the background. Inside those cards, use `surface_container_lowest` (#ffffff) for active input fields or primary data modules. This creates "nested depth" that feels premium and tactile.

### The "Glass & Gradient" Rule

To avoid a generic appearance, use Glassmorphism for floating elements (like a "Check-in" sticky bar). Apply a semi-transparent `surface` color with a 20px backdrop-blur.

- **Signature Gradients:** For high-impact CTAs, use a subtle linear gradient from `primary` (#002c60) to `primary_container` (#1b437c) at a 135-degree angle. This adds "soul" and dimension to an otherwise flat interface.

## 3. Typography

We use a dual-font strategy to balance character with extreme legibility.

- **Display & Headlines (Public Sans):** This is our "Editorial" voice. Use `display-lg` (3.5rem) for hero stats like "Total Hours This Week." Public Sans provides an authoritative, Swiss-inspired precision.

- **Body & Labels (Inter):** This is our "Functional" voice. Inter’s tall x-height ensures that even `label-sm` (0.6875rem) data—like "Timestamp: 08:59:01"—remains perfectly legible during high-speed interactions.

- **Hierarchy Tip:** Never use `body-lg` where a `title-md` is required. Use the `title` scale for component headers to ensure the weight feels "premium."

## 4. Elevation & Depth

In this system, depth is a function of light and tone, not structural scaffolding.

- **The Layering Principle:** Avoid shadows for static components. Elevate a card by shifting it from `surface_container_low` to `surface_container_lowest`.

- **Ambient Shadows:** When an element must float (e.g., a modal or a floating action button), use an extra-diffused shadow.
  - _Spec:_ `0px 12px 32px rgba(25, 28, 29, 0.06)`. This uses the `on_surface` color as the shadow base, making it feel like a natural light obstruction rather than a dark grey "drop shadow."

- **The "Ghost Border" Fallback:** If a border is required for accessibility, use the `outline_variant` (#c3c6d1) at **20% opacity**. It should be felt, not seen.

- **Glassmorphism:** Use `surface_tint` (#3a5e99) at 5% opacity over a blurred background for a "frosted glass" overlay during check-in/check-out confirmation states.

## 5. Components

### Buttons

- **Primary:** Solid `primary` (#002c60) background with `on_primary` (#ffffff) text. Use `lg` (0.5rem) roundedness.

- **Secondary:** `secondary_container` (#cbe7f5) background. No border.

- **States:** On hover, shift the background color to the next tier (e.g., Primary moves to a slightly more vibrant gradient).

### Tonal Chips (Status Indicators)

Instead of standard badges, use tonal chips for status:

- **Check-In:** `tertiary_container` (#004f11) background with `on_tertiary_container` (#72c26e) text.

- **Lunch:** Use a custom Amber-tone shift (derived from tertiary variant) to signal attention.

- **Check-Out/Absence:** `error_container` (#ffdad6) with `on_error_container` (#93000a).

### Input Fields

- **Surface:** `surface_container_lowest` (#ffffff).

- **Indicator:** Instead of a full-box border, use a 2px bottom-bar of `outline` (#737781) that transforms into `primary` (#002c60) on focus.

- **Typography:** Use `body-md` for user input and `label-md` for floating labels.

### Attendance Cards & Lists

- **Rule:** **Forbid dividers.**

- **Implementation:** Separate list items using 12px of vertical white space. If a visual break is needed, use a full-width background strip of `surface_container_low` behind every second item to create a "Zebra" rhythm that feels intentional and architectural.

### The "Time-Stamp" Hero

A unique component for this app. Use `display-md` for the current time, anchored to the top-right of the screen with a wide-margin `surface_container_highest` (#e1e3e4) vertical bar. This creates an asymmetric "anchor" for the user’s eye.

### Evidence Capture Modal

For field operations, use a centered `surface_container_lowest` dialog with a persistent `backdrop-blur`. The camera preview should use a `16:9` aspect ratio with a subtle `outline` (#737781) to frame the subject, following the "Editorial Scale" by using large labels for instructions.

## 6. Do's and Don'ts

### Do

- **Do** use the `lg` (0.5rem) corner radius for cards and `full` (9999px) for status chips.

- **Do** use `on_surface_variant` (#434750) for secondary metadata to create a clear visual hierarchy.

- **Do** embrace negative space. If a screen feels "empty," increase the `display` type size rather than adding borders or lines.

### Don't

- **Don't** use 100% black (#000000). Always use `on_surface` (#191c1d).

- **Don't** use standard "Material Design Blue." Stick to our specific `primary` (#002c60) to maintain the premium, "Bureau" feel.

- **Don't** stack more than three levels of surface nesting. It breaks the "clean" architectural feel and begins to look cluttered.
