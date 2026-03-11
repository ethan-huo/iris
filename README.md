# iris

OCR CLI powered by [PaddleOCR-VL-1.5](https://github.com/PaddlePaddle/PaddleOCR) — extract text from images and PDFs.

- Single image → sync API (fast, direct response)
- PDF / multiple files / URLs → async API (job polling with progress)
- Output: markdown, images, layout visualizations, raw JSON

## Install

```bash
go install github.com/ethan-huo/iris@latest
```

Or build from source:

```bash
git clone https://github.com/ethan-huo/iris.git
cd iris
go build -o bin/iris ./
```

### Claude Code Skill

```bash
bunx skills add ethan-huo/iris
```

## Setup

```bash
iris auth login
```

Paste your Paddle API key (get one at [aistudio.baidu.com](https://aistudio.baidu.com/)). Key is stored in your OS config directory, typically `~/.config/iris/config.yaml` on Linux or `~/Library/Application Support/iris/config.yaml` on macOS.

## Usage

```bash
# Single image (sync)
iris scan photo.png

# PDF (async, auto-detected)
iris scan document.pdf

# Multiple files (async, auto-detected)
iris scan page1.png page2.png page3.png

# URL
iris scan https://example.com/document.pdf

# Custom output directory
iris scan document.pdf -o ./results

# Force async mode
iris scan photo.png -a
```

## Output

```
output/
├── result.json              # Raw API response
├── page_000.md              # Markdown per page
├── page_001.md
├── imgs/                    # Extracted images from document
│   └── img_in_image_box_*.jpg
├── layout_det_res_000.jpg   # Layout detection visualization
└── layout_det_res_001.jpg
```

When scanning multiple inputs, `iris` writes each source into its own subdirectory under the chosen output directory.

## License

MIT
