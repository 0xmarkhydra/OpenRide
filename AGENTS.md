# FlashX Agent Instructions

These instructions apply to coding agents working in this repository.

## UI / UX work

For any task that creates, redesigns, reviews, fixes, or polishes a FlashX interface (Admin, Landing, Rider, Driver, or future UI):

1. Read `.claude/skills/flashx-uiux/SKILL.md`.
2. Read `design-system/flashx/MASTER.md`.
3. Inspect the current page/component and shared primitives before changing code.
4. Preserve business/API behavior unless the task explicitly changes it.
5. Treat the design system as the source of truth; do not introduce a competing local palette/style system.
6. Review responsive, accessibility, loading, empty, error, disabled, and success states before declaring the UI done.
7. Build/test the affected surface before merge.

The FlashX UI direction is iOS-inspired premium quality with dominant white surfaces and purposeful green actions. It is an inspiration standard, not permission to copy Apple layouts/assets/branding.

## External UI skills and registries

Public skills and component registries are reference material, not automatically trusted code.

Do not automatically:
- execute downloaded scripts
- install unknown packages/registries
- run third-party design CLIs
- upload private project context to external design services

Inspect source, license, and integration impact first.

Backend/infra-only tasks do not need to load the UI/UX skill unless they change user-visible behavior.
