---
name: flashx-uiux
description: FlashX-specific UI/UX design and frontend implementation guidance for Admin, Landing, Rider, Driver, and future web/app surfaces. Use whenever creating, redesigning, reviewing, or polishing FlashX interfaces.
user-invocable: false
---

# FlashX UI/UX Skill

This is the canonical UI/UX skill for FlashX. It synthesizes proven ideas from frontend-design, shadcn/ui guidance, UI/UX Pro Max, and Superdesign, but is intentionally adapted to FlashX and does not depend on third-party CLIs, external design services, or executable scripts.

Before changing UI, read:
1. `design-system/flashx/MASTER.md`
2. The current page/component source
3. Existing shared UI primitives before creating new ones

## Product truth

FlashX is: **“Tài xế của bạn, khi bạn cần.”**

MVP services:
- Lái hộ ô tô (`designated_driver_car`)
- Lái hộ xe máy (`designated_driver_bike`)
- Đăng kiểm hộ (`vehicle_inspection_assist`)

The customer owns the vehicle. Do not visually or verbally imply that FlashX is a standard Grab/Uber passenger fleet product.

## Design north star

Create a product that feels **premium, calm, modern, spatial, and native-like**. The visual inspiration is contemporary iOS quality: precise spacing, clean white surfaces, controlled translucency, clear hierarchy, high-quality motion, and excellent touch behavior. Do not clone Apple screens, proprietary assets, or branding.

FlashX brand direction:
- White is the dominant surface.
- Green is the action/brand color, not wallpaper.
- Dark text provides structure.
- Depth is subtle and layered.
- Motion is restrained and meaningful.
- Data-heavy Admin screens remain easy to scan.

### Signature element: FlashX Flowline

Use one memorable visual idea consistently: a subtle green route/timeline line with a soft active pulse that represents a job moving through its lifecycle. It may appear in hero visuals, trip/job timelines, inspection progress, and live-status surfaces. Do not scatter glowing effects everywhere.

## Non-negotiable anti-patterns

Avoid:
- purple/blue SaaS gradients as a default visual language
- dark enterprise-dashboard styling as the main theme
- oversized generic KPI cards everywhere
- decorative glassmorphism on every surface
- emoji as structural/navigation icons
- random raw hex/Tailwind colors inside feature components
- excessive borders, shadows, pills, gradients, or animations
- buttons that look enabled but do nothing
- placeholder navigation that cannot be clicked
- fake claims, fake metrics, fake testimonials, fake guarantees
- a page that looks like unmodified default shadcn/ui
- copying another brand's layout pixel-for-pixel

## Workflow

### 1. Understand the screen before styling

Identify:
- who is using it
- the screen's single primary job
- the most important decision/action
- critical states: loading, empty, error, success, disabled, offline/reconnect where relevant
- device context: desktop ops, mobile customer, mobile driver, marketing visitor

For an existing page, preserve working business logic unless the task explicitly changes it.

### 2. Inspect before inventing

Before creating UI primitives:
- check existing components
- check `components.json`
- reuse Button/Card/Badge/Table/Sheet/Dialog/etc. where suitable
- compose components rather than making styled one-off divs
- use existing project icon library; FlashX currently prefers Lucide on web

Do not install or execute third-party UI registries/scripts automatically. If external code is considered, inspect the source and license first.

### 3. Create a compact page plan

Before implementation, establish:
- hierarchy
- layout concept
- primary CTA
- information density
- responsive transformation
- one signature moment, if the page needs one

Do not spend creativity on decoration that does not improve comprehension.

### 4. Build with semantic tokens

All visual decisions should come from `design-system/flashx/MASTER.md`.

Prefer semantic names such as:
- `background`
- `surface`
- `primary`
- `primary-foreground`
- `muted`
- `muted-foreground`
- `border`
- `success`
- `warning`
- `destructive`
- `info`

Feature components should not introduce arbitrary brand colors.

### 5. Component rules

For web Admin/Landing:
- Next.js + React
- Tailwind CSS
- shadcn-style primitives
- TanStack Table for complex tabular data
- React Hook Form + Zod for non-trivial forms
- Lucide icons unless project configuration changes

Use component variants before custom styling.
Use `gap-*` layouts rather than `space-x-*` / `space-y-*`.
Use `size-*` when width and height are equal.
Use `cn()` for conditional class merging.
Use `Skeleton` for loading, `Alert` for callouts/errors, `Badge` for statuses, and `Dialog`/`Sheet`/`Drawer` for overlays when available.

Every dialog/sheet/drawer must have an accessible title.
Every icon-only button must have an accessible name.
Do not use color as the only status indicator.

### 6. iOS-level interaction quality

Target controls:
- minimum interactive hit area: 44×44 CSS px where practical
- clear pressed/active feedback
- no layout-jumping hover/press transforms
- visible keyboard focus
- desktop hover must never be the only way to discover an action

Motion:
- quick control feedback: ~140–180ms
- normal transitions: ~200–260ms
- larger sheet/page spatial transitions: ~280–360ms
- favor opacity/transform over width/height animation
- respect `prefers-reduced-motion`

Animation must explain continuity, state, or priority. If motion is purely decorative, remove it unless it is the page's deliberate signature moment.

### 7. Typography and copy

Use the platform/system type stack defined in the master system. FlashX values legibility and native familiarity over decorative font novelty.

Copy rules:
- write from the user's point of view
- use plain Vietnamese for user-facing Vietnamese UI
- buttons describe the result: “Lưu thay đổi”, “Duyệt tài xế”, “Đặt dịch vụ”
- keep the same verb through action → loading → success feedback
- errors say what happened and what the user can do next
- empty states point to the next action
- do not expose internal implementation vocabulary unnecessarily

### 8. Admin-specific UX

The Admin home screen should answer in seconds:
1. What is happening now?
2. What needs attention?
3. What can I act on immediately?

Prioritize actionable queues over decorative analytics.

For job operations:
- make lifecycle/timeline prominent
- show service type, customer, driver/provider, vehicle, schedule, price, incident state
- expose the next valid operational action clearly
- avoid normal cancel/reassignment after `VEHICLE_RECEIVED` unless backend rules explicitly permit it
- make inspection workflow visually distinct enough to reflect its longer lifecycle

Dense desktop tables should become simplified cards/list rows on small screens rather than forcing horizontal scrolling whenever feasible.

### 9. Landing-specific UX

Landing should communicate immediately:
- customer uses their own vehicle
- FlashX finds the driver/person to perform the service
- exactly three MVP services
- booking can be immediate or scheduled where supported

Hero should feel premium and product-specific, not like a generic startup template.
Use authentic product flows and vehicle/service visuals rather than abstract SaaS charts.

Do not claim insurance, nationwide coverage, 24/7 availability, app-store status, regulatory approval, or other facts unless they are verified in current product context.

### 10. Responsive behavior

Review at minimum:
- ~375px mobile
- ~768px tablet/narrow desktop
- ~1280px desktop
- wide desktop for Admin data surfaces

Requirements:
- no unintended horizontal scroll
- sticky/fixed UI must not hide content
- mobile navigation must remain reachable
- important CTAs remain visible without crowding
- text size remains readable without zoom

For future native/mobile apps, respect safe areas and platform navigation conventions.

### 11. Accessibility quality floor

Before delivery verify:
- normal text contrast targets WCAG AA (4.5:1)
- meaningful non-text controls have sufficient contrast
- keyboard navigation works on web
- focus is not hidden by sticky headers/overlays
- form fields have labels and nearby errors
- icon controls have accessible names
- decorative icons are hidden from assistive technology when appropriate
- reduced motion is supported
- status is not encoded by color alone

### 12. Self-critique before declaring done

Review the rendered result, not only the code. If screenshots/browser preview are available, inspect them.

Ask:
- Does this look specifically like FlashX?
- Is green used as a meaningful action/status color rather than decoration?
- Can a first-time user understand the primary action in 3 seconds?
- Is the hierarchy calmer than the previous version?
- Is there one memorable detail rather than ten competing effects?
- Do loading, empty, error, disabled, and success states look intentional?
- Does mobile feel designed rather than merely collapsed?
- Are all visible controls functional?

Remove at least one unnecessary decorative element during the final polish pass.

## Delivery gate

Do not call a UI task complete until:
- production build passes
- TypeScript/lint/test checks relevant to the changed surface pass
- responsive states have been reviewed
- critical actions still work against the existing backend/API
- no unrelated Railway service rebuilds are caused by the change
- there is no placeholder control presented as functional

## Safety and dependency policy

This skill itself must not:
- execute downloaded scripts
- silently install packages
- run third-party CLIs
- send project/private data to external design services
- overwrite a working design system without reviewing it

External libraries or registry components may be used only when they provide clear value, their source/license is acceptable, and integration is reviewed before merge.
