---
type: community
cohesion: 0.25
members: 8
---

# Page-Level Components

**Cohesion:** 0.25 - loosely connected
**Members:** 8 nodes

## Members
- [[FormSuratPage (pages)]] - code - web/kiosk-ui/src/pages/FormSuratPage.tsx
- [[HomePage (pages)]] - code - web/kiosk-ui/src/pages/HomePage.tsx
- [[Package pages]] - code - web/dashboard/src/pages
- [[PreviewPage (pages)]] - code - web/kiosk-ui/src/pages/PreviewPage.tsx
- [[SelectSuratPage (pages)]] - code - web/kiosk-ui/src/pages/SelectSuratPage.tsx
- [[SuccessPage (pages)]] - code - web/kiosk-ui/src/pages/SuccessPage.tsx
- [[dashboard (pages)]] - code - web/dashboard/src/pages/dashboard.astro
- [[index (pages)]] - code - web/dashboard/src/pages/index.astro

## Live Query (requires Dataview plugin)

```dataview
TABLE source_file, type FROM #community/Page-Level_Components
SORT file.name ASC
```
