---
name: iris
description: OCR tool for extracting text from images and PDFs. Use when user wants to OCR scan documents, extract text from screenshots, or process PDF files.
---

# iris — OCR Processing

`iris` is a CLI tool powered by PaddleOCR-VL-1.5 that extracts text from images and PDFs. It auto-selects sync (single image) or async (PDF/multi-file) API mode.

Binary location: `iris`

## When to Use

- User wants to extract text from an image or screenshot
- User needs to OCR a PDF document
- User wants to batch-process multiple images/PDFs
- User has a URL pointing to an image or PDF to extract text from

## Commands

```bash
# Single image (fast sync API)
iris scan photo.png

# PDF (auto-selects async API with progress polling)
iris scan document.pdf

# Multiple files (async)
iris scan page1.png page2.png page3.png

# URL (async)
iris scan https://example.com/document.pdf

# Custom output directory (default: ./output)
iris scan document.pdf -o ./results

# Force async mode for a single image
iris scan photo.png -a
```

### Auth

```bash
# Set up API key (interactive, password-masked input)
iris auth login

# Check current auth status
iris auth status

# Show config
iris config
```

## Output Structure

```
output/
├── result.json              # Raw API response
├── page_000.md              # Extracted text per page (markdown)
├── page_001.md
├── imgs/                    # Inline images extracted from document
│   └── img_in_image_box_*.jpg
├── layout_det_res_000.jpg   # Layout detection visualization
└── layout_det_res_001.jpg
```

When scanning multiple inputs, each source gets its own numbered subdirectory (e.g., `01_filename/`, `02_filename/`).

## Reading Results

- **Markdown files** (`page_*.md`) — the extracted text, ready to use
- **Layout images** (`layout_det_res_*.jpg`) — visual overlay showing detected regions, use `Read` to view
- **result.json** — full API response if you need raw data

## Strategy

1. **Always check auth first** if the user hasn't used iris before: `iris auth status`
2. **Default output dir is `./output`** — use `-o` to avoid overwriting previous results
3. **Single image → sync** (instant result), **PDF/multi-file → async** (polls every 3s with progress)
4. After scan completes, read the `page_*.md` files to show the user extracted text

## Important Notes

- Requires a Paddle API key from [aistudio.baidu.com](https://aistudio.baidu.com/)
- API key is stored at `~/.config/iris/config.yaml` (Linux) or `~/Library/Application Support/iris/config.yaml` (macOS)
- Do NOT run `iris scan` without confirming the user has set up auth
- For large PDFs, the async polling may take a while — inform the user about progress
