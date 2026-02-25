package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	problem := flag.String("problem", "", "Single problem slug or URL")
	output := flag.String("output", "leetcode.pdf", "Output file")
	pageSize := flag.String("size", "A4", "Page size")
	topic := flag.String("topic", "", "Topic slug")
	difficulty := flag.String("difficulty", "", "Easy|Medium|Hard")
	list := flag.String("list", "", "Public list slug")
	workers := flag.Int("workers", 5, "Consurrent workers")
	flag.Parse()

	var slugs []string
	var err error
	var problems []*ProblemDetail

	if *problem != "" {

		slug := normalizeProblemInput(*problem)
		fmt.Println("Fetching single problem:", slug)

		p, err := fetchProblem(ctx, slug)
		if err != nil {
			log.Fatal(err)
		}

		problems = []*ProblemDetail{p}

	} else {
		if *list != "" {
			slugs, err = fetchSlugsFromList(ctx, *list)
		} else if *topic != "" {
			slugs, err = fetchSlugsByTopic(ctx, *topic, *difficulty)
		} else {
			log.Fatal("You must provide --topic, --list, or --problem")
		}

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Found %d problems\n", len(slugs))

		problems = fetchAllProblems(ctx, slugs, *workers)
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Started to generate pdf, data was fetched successfully")

	problems = fetchAllProblems(ctx, slugs, *workers)
	if ctx.Err() != nil {
		log.Println("Shutdown requested. Exiting cleanly.")
		return
	}

	if err := generatePDF(ctx, problems, *output, *pageSize); err != nil {
		log.Fatal(err)
	}

	fmt.Println("PDF generated:", *output)
}
