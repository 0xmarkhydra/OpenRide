# FlashX Design System — MASTER

Status: canonical source of truth for FlashX visual/interaction design.

Brand promise: **“Tài xế của bạn, khi bạn cần.”**

This system applies to FlashX Admin, Landing, Rider, Driver, and future web/app surfaces unless a documented page-level override exists.

## 1. Design character

FlashX should feel:
- premium
- calm
- trustworthy
- native-like
- modern
- spatial
- fast
- human

Visual inspiration: contemporary iOS quality, without copying Apple screens or branding.

Core rule: **white owns the canvas; green earns attention.**

## 2. Signature

### FlashX Flowline

A thin route/timeline line with a restrained green active pulse. It represents service progress from request → arrival → vehicle receipt → execution → handover.

Use it in:
- landing hero/product illustration
- active-job timeline
- inspection timeline
- live operational summaries

Do not use it as generic decoration on every card.

## 3. Color system

Use semantic tokens in components. Raw color values belong in the theme layer only.

### Core neutrals

| Token | Value | Use |
|---|---:|---|
| `background` | `#F7F9F8` | app/page background |
| `surface` | `#FFFFFF` | cards, sheets, nav |
| `surface-subtle` | `#F1F5F3` | secondary grouped surface |
| `foreground` | `#101828` | primary text |
| `muted-foreground` | `#667085` | secondary text |
| `border` | `rgba(16, 24, 40, 0.08)` | subtle boundaries |
| `border-strong` | `rgba(16, 24, 40, 0.14)` | strong separation |

### Brand green

| Token | Value | Use |
|---|---:|---|
| `brand-50` | `#ECFDF3` | selected/soft success surface |
| `brand-100` | `#D1FADF` | soft brand surface |
| `brand-400` | `#32D583` | decorative highlight only |
| `brand-500` | `#12B76A` | FlashX accent |
| `brand-600` | `#079455` | primary actions |
| `brand-700` | `#067647` | pressed/strong brand state |
| `brand-900` | `#054F31` | deep brand text/icon |

Primary action should normally use `brand-600` with a contrast-safe foreground. Bright green (`brand-400/500`) is an accent, not default body/button background when contrast would suffer.

### State colors

| Token | Value | Use |
|---|---:|---|
| `success` | `#079455` | completed/healthy |
| `warning` | `#DC6803` | waiting/attention |
| `destructive` | `#D92D20` | failure/cancel/incident |
| `info` | `#175CD3` | neutral information |

Status must also include text/icon/shape; never color alone.

## 4. Typography

Use a native-feeling system stack; do not bundle proprietary Apple fonts.

Recommended web stack:

```css
font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Inter, Roboto, Helvetica, Arial, sans-serif;
```

Utility/IDs:

```css
font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
```

Type scale:
- Hero display: `52–72px`, tight tracking, 700–800
- Page title: `28–36px`, 700
- Section title: `20–24px`, 650–700
- Card title: `16–18px`, 600–700
- Body: `15–17px`, 400–500
- UI label: `13–15px`, 500–600
- Caption/metadata: `12–13px`, 500

Avoid body copy below 14px except compact metadata that remains legible.

## 5. Spacing rhythm

Base rhythm: 4/8px.

Preferred spacing tokens:
- 4
- 8
- 12
- 16
- 20
- 24
- 32
- 40
- 48
- 64
- 80
- 96

Use larger whitespace to establish hierarchy before adding borders or backgrounds.

Page gutters:
- mobile: 16px
- tablet: 24px
- desktop: 28–32px
- large marketing sections: up to 40px internal gutters

## 6. Radius

- compact controls: 10–12px
- primary controls/input: 12–14px
- cards: 16–20px
- sheets/modals: 20–28px
- hero/product panels: up to 28px
- badges/status pills: full radius only when semantically appropriate

Avoid making every container a pill.

## 7. Depth and translucency

Shadows are quiet.

Suggested layers:
- `shadow-xs`: subtle control separation
- `shadow-sm`: cards
- `shadow-md`: floating sheets/dropdowns
- `shadow-lg`: rare hero/floating product preview

Prefer low-opacity dark shadows with larger blur over hard gray drops.

Translucency:
- sticky nav/header can use white at ~82–92% opacity
- backdrop blur ~16–28px
- only use blur where content moves beneath a persistent surface
- do not make normal content cards translucent

## 8. Interaction tokens

Minimum practical hit target: 44×44 CSS px.

Motion durations:
- press/hover response: 140–180ms
- standard transition: 200–260ms
- sheet/page spatial movement: 280–360ms

Preferred easing:
- enter/move: `cubic-bezier(0.22, 1, 0.36, 1)`
- exit: slightly faster than enter

Rules:
- animate transform/opacity where possible
- no layout-jumping scale effects on tables/forms
- honor `prefers-reduced-motion`
- one orchestrated motion moment is better than many scattered animations

## 9. Iconography

Web default: Lucide.

Rules:
- one icon family per surface
- consistent stroke weight
- do not use emoji for navigation/system actions
- standalone icon buttons need accessible names/tooltips where appropriate
- decorative icons beside clear text should not duplicate screen-reader output

## 10. Components

### Button

Primary:
- green, high contrast
- 44–48px height for major actions
- clear pressed state

Secondary:
- white or subtle surface
- visible border or tonal background

Destructive:
- red semantic treatment; never green

Avoid multiple primary buttons in the same decision area.

### Input

- 44–48px minimum height for main forms
- labels remain visible; placeholder is not the label
- focus ring uses brand semantic token
- errors appear next to the field
- OTP may use segmented input when component support is available

### Card

- white surface
- minimal border
- subtle elevation
- use header/content/footer composition
- spacing and grouping should carry hierarchy before decorative effects

### Badge

Use for short statuses only. Pair color with text.

### Table / data list

Desktop Admin:
- readable rows with strong first column
- sticky header only when useful
- filters/search outside the table body
- row action menus do not overwhelm scanning

Small screens:
- prefer list/card transformations for operational data
- avoid forcing wide desktop tables into horizontal scroll unless the data truly requires it

### Dialog / Sheet / Drawer

- Dialog: focused decision/task
- Sheet: contextual detail on desktop
- Drawer: mobile bottom interaction/details
- always provide title and clear close/back behavior

## 11. Navigation

### Admin desktop

- left navigation around 248–264px
- white/light surface, not permanent dark SaaS sidebar
- active item uses soft green surface + strong green/foreground
- sticky translucent topbar may be used

### Admin mobile

Keep up to five highest-frequency destinations directly reachable. Put lower-frequency destinations in a More sheet/drawer.

### Landing

Sticky white/translucent nav. Do not hide primary CTA behind complex mega menus.

## 12. Admin information architecture

Priority order:
1. Attention / incidents
2. Active jobs
3. Driver/provider supply
4. Upcoming scheduled work
5. Operational metrics
6. Historical analytics

Dashboard should emphasize actions, not decorative statistics.

Job detail should prioritize:
- lifecycle timeline
- service
- customer
- vehicle
- driver/provider
- schedule
- price
- incident state
- valid next operational actions

Inspection service needs its expanded lifecycle and separate inspection result.

## 13. Landing page direction

Hero thesis:
- “Tài xế của bạn, khi bạn cần.”
- make own-vehicle model obvious
- show the three services plainly

Primary CTA: booking/demo action appropriate to current product state.
Secondary CTA: driver/provider acquisition when relevant.

Recommended visual signature: an iPhone-like product/service panel or abstracted route interface using the FlashX Flowline, not a generic analytics dashboard.

Do not use unverifiable trust claims.

## 14. Content tone

Vietnamese UI:
- concise
- natural
- direct
- sentence case
- action verbs

Examples:
- `Duyệt tài xế`
- `Từ chối hồ sơ`
- `Xem công việc`
- `Báo sự cố`
- `Hoàn tất bàn giao`

Avoid internal jargon when a user-facing term exists.

## 15. Accessibility floor

- WCAG AA target for normal text: 4.5:1
- visible keyboard focus
- no color-only statuses
- accessible names for icon actions
- labeled fields
- reduced motion support
- sticky UI must not cover focused content
- readable at zoom and small viewport

## 16. Quality checklist

Before merge:
- [ ] Looks recognizably FlashX, not generic shadcn/SaaS
- [ ] White remains dominant; green is purposeful
- [ ] Primary action is obvious within ~3 seconds
- [ ] All visible controls work
- [ ] Loading state exists
- [ ] Empty state exists where relevant
- [ ] Error state explains recovery
- [ ] Disabled state is clear
- [ ] 375px layout reviewed
- [ ] 768px layout reviewed
- [ ] 1280px+ layout reviewed
- [ ] Keyboard/focus behavior checked on web
- [ ] Reduced motion supported
- [ ] Build/type checks pass
- [ ] Business/API behavior preserved unless explicitly changed
