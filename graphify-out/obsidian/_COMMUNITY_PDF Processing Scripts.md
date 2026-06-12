---
type: community
cohesion: 0.13
members: 15
---

# PDF Processing Scripts

**Cohesion:** 0.13 - loosely connected
**Members:** 15 nodes

## Members
- [[Package scripts]] - code - .agents/skills/diagnose/scripts
- [[build (scripts)]] - code - scripts/build.ps1
- [[check_bounding_boxes (scripts)]] - code - .agents/skills/pdf/scripts/check_bounding_boxes.py
- [[check_fillable_fields (scripts)]] - code - .agents/skills/pdf/scripts/check_fillable_fields.py
- [[convert_pdf_to_images (scripts)]] - code - .agents/skills/pdf/scripts/convert_pdf_to_images.py
- [[create_validation_image (scripts)]] - code - .agents/skills/pdf/scripts/create_validation_image.py
- [[extract_form_field_info (scripts)]] - code - .agents/skills/pdf/scripts/extract_form_field_info.py
- [[extract_form_structure (scripts)]] - code - .agents/skills/pdf/scripts/extract_form_structure.py
- [[fill_fillable_fields (scripts)]] - code - .agents/skills/pdf/scripts/fill_fillable_fields.py
- [[fill_pdf_form_with_annotations (scripts)]] - code - .agents/skills/pdf/scripts/fill_pdf_form_with_annotations.py
- [[hitl-loop.template (scripts)]] - code - .agents/skills/diagnose/scripts/hitl-loop.template.sh
- [[mock-rfid (scripts)]] - code - scripts/mock-rfid.ps1
- [[read_docx (scripts)]] - code - scripts/read_docx.ps1
- [[setup-kiosk (scripts)]] - code - scripts/setup-kiosk.ps1
- [[with_server (scripts)]] - code - .agents/skills/webapp-testing/scripts/with_server.py

## Live Query (requires Dataview plugin)

```dataview
TABLE source_file, type FROM #community/PDF_Processing_Scripts
SORT file.name ASC
```
