package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// doGraphQL executes a GraphQL POST request against the LeetCode endpoint.
//
// It is context-aware and will abort immediately if the provided context
// is canceled (e.g., on SIGINT or shutdown).
//
// Behavior:
//   - Retries up to 3 times with incremental backoff.
//   - Aborts early if ctx is canceled.
//   - Returns the last encountered error if all retries fail.
//   - Decodes the response JSON into the provided output struct.
//
// This function is the single entry point for all GraphQL calls in the system.
// All network requests should flow through here to ensure consistent retry
// and lifecycle behavior.
func doGraphQL(ctx context.Context, query GraphQLQuery, out any) error {
	jsonData, err := json.Marshal(query);
	if err != nil {
		return err
	}

	var lastErr error

	for attemp := range 3 {
		req, err := http.NewRequestWithContext(ctx, "POST", graphqlEndpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Referer", "https:leetcode.com")

		resp, err := httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			lastErr = err
			resp.Body.Close()
			time.Sleep(time.Second * time.Duration(attemp+1))
			continue			
		}

		err = json.NewDecoder(resp.Body).Decode(out)
		resp.Body.Close()
		if err != nil {
			return err
		}
		return nil
	}
	return lastErr
}