# LeetCode PDF Exporter CLI

A Go CLI tool to fetch LeetCode problems by topic, difficulty, public list, or single URL, and generate a PDF with full problem statements and example test cases.

This tool uses **headless Chrome** (`chromedp`) for HTML rendering and PDF generation. It does **not** require `wkhtmltopdf` or any other external PDF tools.

---

## Features

* Fetch problems by topic tag (e.g., dynamic programming, arrays, etc.)
* Filter problems by difficulty (Easy, Medium, Hard)
* Fetch a single problem by URL
* Fetch all problems from a public LeetCode list
* Generate a well-formatted PDF with problem content and examples
* Automatic `.pdf` extension handling
* Graceful shutdown with context cancellation
* Fully portable, works on Windows, Linux, and macOS

---

## Prerequisites

* **Go** >= 1.20 installed
* **Google Chrome** or **Chromium** installed and available in your PATH

Check Chrome installation:

```bash
chrome --version
# or
chromium --version
```

---

## Installation

Clone the repository:

```bash
git clone https://github.com/yourusername/leetcode-pdf-exporter.git
cd leetcode-pdf-exporter
```

Get dependencies:

```bash
go mod tidy
```

---

## Usage

### Fetch problems by topic and difficulty

Generate a PDF of all Medium dynamic programming problems:

```bash
go run . -topic=dynamic-programming -difficulty=Medium -output=medium_dp.pdf
```

### Fetch a single problem by URL

```bash
go run . -problem=https://leetcode.com/problems/coin-change/ -output=coin_change.pdf
```

### Notes

* The `-output` filename **does not need `.pdf`**, it will be added automatically if missing:

```bash
go run . -topic=arrays -difficulty=Easy -output=easy_arrays
# Output file: easy_arrays.pdf
```

* Large topic PDFs may take a few minutes to generate depending on the number of problems.

---

## Flags

| Flag          | Description                                                             |
| ------------- | ----------------------------------------------------------------------- |
| `-topic`      | LeetCode topic tag slug (e.g., `dynamic-programming`)                   |
| `-difficulty` | Difficulty level (`Easy`, `Medium`, `Hard`)                             |
| `-problem`    | Single problem URL (e.g., `https://leetcode.com/problems/coin-change/`) |
| `-output`     | Output PDF file path (automatically adds `.pdf`)                        |

---

## Example

Generate a PDF for all Medium dynamic programming problems:

```bash
go run . -topic=dynamic-programming -difficulty=Medium -output=medium_dp
```

* Produces `medium_dp.pdf` in the current directory
* Properly formatted with headings, problem content, and example testcases

---

