package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// generatePDF renders a slice of ProblemDetail objects into a PDF using headless Chrome.
//
// Features:
//  - Sorts problems by difficulty (Easy -> Medium -> Hard)
//  - Safely injects full HTML into the page (no data URL length limit)
//  - Renders example testcases in <pre> blocks
//  - Fully context-aware for graceful shutdown
//  - Automatically creates output directories
//  - Automatically ensures .pdf extension (UPDATE)
func generatePDF(ctx context.Context, problems []*ProblemDetail, output, pageSize string) error {

	// ===== UPDATE: Ensure the output has .pdf extension =====
	if !strings.HasSuffix(output, ".pdf") {
		output += ".pdf"
	}

	// Sort problems by difficulty
	sort.Slice(problems, func(i, j int) bool {
		order := map[string]int{"Easy": 1, "Medium": 2, "Hard": 3}
		return order[problems[i].Difficulty] < order[problems[j].Difficulty]
	})

	// Build HTML content
	html := buildHTML(problems)

	// Create Chrome allocator (headless)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx,
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer cancelAlloc()

	// Create browser context
	chromeCtx, cancelChrome := chromedp.NewContext(allocCtx)
	defer cancelChrome()

	var pdfBuffer []byte

	err := chromedp.Run(chromeCtx,
		// Navigate to blank page
		chromedp.Navigate("about:blank"),

		// ===== UPDATE: Inject HTML safely (no data URL) =====
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}

			if err := page.SetDocumentContent(frameTree.Frame.ID, html).Do(ctx); err != nil {
				return err
			}

			// ===== UPDATE: Optional debug to ensure content loaded =====
			// var bodyLength int
			// if err := chromedp.Evaluate(`document.body.innerHTML.length`, &bodyLength).Do(ctx); err != nil {
			//     return err
			// }
			// fmt.Println("HTML body length:", bodyLength)

			return nil
		}),

		// Small delay to allow DOM to fully render
		chromedp.Sleep(500*time.Millisecond),

		// Print to PDF
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(8.27).  // A4 width in inches
				WithPaperHeight(11.69). // A4 height in inches
				Do(ctx)
			if err != nil {
				return err
			}

			// ===== UPDATE: Check PDF buffer before writing =====
			if len(buf) == 0 {
				return fmt.Errorf("PDF buffer is empty — Chrome failed to generate PDF")
			}

			pdfBuffer = buf
			return nil
		}),
	)

	if err != nil {
		return err
	}

	// Ensure output directory exists
	dir := filepath.Dir(output)
	if dir != "." {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	// ===== UPDATE: Write PDF file as bytes =====
	return os.WriteFile(output, pdfBuffer, 0644)
}

// buildHTML generates the full HTML document for the problems slice.
func buildHTML(problems []*ProblemDetail) string {
	var html strings.Builder

	html.WriteString(`
	<html>
	<head>
	<meta charset="UTF-8">
	<style>
	body { font-family: Arial, sans-serif; margin: 40px; }
	h1 { page-break-before: always; font-size: 18pt; }
	h2 { font-size: 14pt; margin-top: 12px; }
	pre {
		background: #f4f4f4;
		padding: 10px;
		white-space: pre-wrap;
		word-wrap: break-word;
		font-family: monospace;
	}
	code { font-family: monospace; }
	hr { border: 0; border-top: 1px solid #ccc; margin: 20px 0; }
	</style>
	</head>
	<body>
	`)

	for _, p := range problems {
		html.WriteString("<h1>" + p.Title + " (" + p.Difficulty + ")</h1>")
		html.WriteString(p.Content)
		html.WriteString("<h2>Example Testcases</h2>")
		html.WriteString("<pre>" + p.ExampleTestcases + "</pre>")
		html.WriteString("<hr>")
	}

	html.WriteString("</body></html>")

	return html.String()
}