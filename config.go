package main

import (
	"net/http"
	"time"
)

const (
	graphqlEndpoint = "https://leetcode.com/graphql"
	timeout = 30 * time.Second
)

var httpClient = &http.Client{
	Timeout: timeout,
}