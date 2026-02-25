package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

const questionQuery = `
query questionData($titleSlug: String!) {
  question(titleSlug: $titleSlug) {
    title
    content
    difficulty
    exampleTestcases
  }
}
`

const problemListQuery = `
query problemsetQuestionList($categorySlug: String, $limit: Int, $skip: Int, $filters: QuestionListFilterInput) {
  questionList(
    categorySlug: $categorySlug
    limit: $limit
    skip: $skip
    filters: $filters
  ) {
    totalNum
    data {
      titleSlug
      difficulty
    }
  }
}
`

// fetchProblem retrieves full problem details for a single LeetCode question.
//
// It calls the `question` GraphQL endpoint using the provided title slug.
//
// Returns:
//   - Full problem HTML content
//   - Title
//   - Difficulty
//   - Example test cases
//
// This function is used in the second stage of the pipeline after
// slug discovery. It performs one HTTP request per problem.
func fetchProblem(ctx context.Context, slug string) (*ProblemDetail, error) {
	query := GraphQLQuery{
		Query: questionQuery,
		Variables: map[string]any{
			"titleSlug": slug,
		},
	}

	var resp ProblemResponse
	if err := doGraphQL(ctx, query, &resp); err != nil {
		return nil, err
	}

	q := resp.Data.Question
	return &ProblemDetail{
		Title:            q.Title,
		Content:          q.Content,
		Difficulty:       q.Difficulty,
		ExampleTestcases: q.ExampleTestcases,
	}, nil
}

// fetchAllProblems concurrently retrieves full problem details
// for all provided title slugs.
//
// This function implements a worker pool pattern:
//
//   - `workers` defines the number of concurrent goroutines.
//   - A jobs channel distributes slugs to workers.
//   - A results channel collects successful fetches.
//   - All goroutines are context-aware and exit immediately
//     when ctx is canceled.
//
// Lifecycle behavior:
//   - If ctx is canceled (SIGINT/SIGTERM), workers stop processing.
//   - In-flight HTTP requests are aborted via context propagation.
//   - The function returns partial results collected so far.
//
// Returns:
//   - A slice of successfully fetched ProblemDetail objects.
//
// This is the second stage of the data pipeline following slug discovery.
func fetchAllProblems(ctx context.Context, slugs []string, worker int) []*ProblemDetail {
	fmt.Println("Fetching problem details ...")
	var wg sync.WaitGroup
	jobs := make(chan string)
	results := make(chan *ProblemDetail)

	// Workers
	for i := 0; i < worker; i ++ {
		wg.Add(1)
		go func ()  {
			defer wg.Done()
			for {
				select  {
				case <- ctx.Done():
					return
				case slug, ok := <-jobs:
					if !ok {
						return
					}

					problem, err := fetchProblem(ctx, slug)
					if err != nil {
						fmt.Println("fetch error:", slug, err)
						continue
					}

					select {
					case <- ctx.Done():
						return
					case results <- problem:
					}
				}
			}
		}()
	}

	// Job feeder
	go func ()  {
		defer close(jobs)
		for _, slug := range slugs {
			select {
			case <- ctx.Done():
				return
			case jobs <- slug:
			}
		}
	}()

	//Rsult closer
	go func ()  {
		wg.Wait()
		close(results)
	}()

	var problems []*ProblemDetail
	for {
		select {
		case <- ctx.Done():
			return problems
		case problem, ok := <-results:
			if !ok {
				return problems
			}
			problems = append(problems, problem)
		}
	}
}

// fetchSlugsByTopic retrieves all problem title slugs for a given topic tag
// and optional difficulty filter.
//
// This function performs paginated GraphQL requests using `limit` and `skip`
// until all matching problems are collected.
//
// Parameters:
//   - topic: LeetCode tag slug (e.g., "dynamic-programming")
//   - difficulty: Optional filter ("Easy", "Medium", "Hard")
//
// Returns:
//   - A slice of title slugs matching the criteria.
//
// Note:
//   This function only retrieves metadata (titleSlug + difficulty).
//   It does NOT retrieve full problem content.
func fetchSlugsByTopic(ctx context.Context, topic string, difficulty string) ([]string, error) {
	var slugs []string
	skip := 0 
	limit := 50
	total := -1

	filters := map[string]any {
		"tags": []string{topic},
	}
	if difficulty != "" {
		filters["difficulty"] = strings.ToUpper(difficulty)
	}

	for {
		query := GraphQLQuery {
			Query: problemListQuery,
			Variables: map[string]any{
				"categorySlug": "",
				"skip":         skip,
				"limit":        limit,
				"filters":      filters,
			},
		}

		var resp ProblemListResponse
		if err := doGraphQL(ctx, query, &resp); err != nil {
			return nil, err
		}

		if total == -1 {
			total = resp.Data.QuestionList.Total
		}

		for _, q := range resp.Data.QuestionList.Data {
			slugs = append(slugs, q.TitleSlug)
		}

		if len(slugs) >= total {
			break
		}

		skip += limit
	}
	return slugs, nil
}

// fetchSlugsFromList retrieves all problem slugs from a specific public
// LeetCode problem list identified by its categorySlug.
//
// Example slugs:
//   - "top-100-liked-questions"
//   - "blind-75"
//
// Internally, this reuses the same paginated query mechanism as topic fetching.
//
// Returns:
//   - A slice of title slugs belonging to the list.
//
// Note:
//   Only metadata is retrieved at this stage.
func fetchSlugsFromList(ctx context.Context, categorySlug string) ([]string, error) {
	return fetchSlugsByTopic(ctx, categorySlug, "")
}