---
type: community
cohesion: 0.33
members: 6
---

# PDF Form Processing

**Cohesion:** 0.33 - loosely connected
**Members:** 6 nodes

## Members
- [[A.1 Analyze the Structure]] - document - .agents/skills/pdf/forms.md
- [[Approach A Structure-Based Coordinates (Preferred)]] - document - .agents/skills/pdf/forms.md
- [[Doc forms]] - document - .agents/skills/pdf/forms.md
- [[Fillable fields]] - document - .agents/skills/pdf/forms.md
- [[Non-fillable fields]] - document - .agents/skills/pdf/forms.md
- [[Step 1 Try Structure Extraction First]] - document - .agents/skills/pdf/forms.md

## Live Query (requires Dataview plugin)

```dataview
TABLE source_file, type FROM #community/PDF_Form_Processing
SORT file.name ASC
```
