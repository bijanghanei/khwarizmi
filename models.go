package main

type GraphQLQuery struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type ProblemDetail struct {
	Title            string
	Content          string
	Difficulty       string
	ExampleTestcases string
}

type ProblemResponse struct {
	Data struct {
		Question struct {
			Title            string `json:"title"`
			Content          string `json:"content"`
			Difficulty       string `json:"difficulty"`
			ExampleTestcases string `json:"exampleTestcases"`
		} `json:"question"`
	} `json:"data"`
}

type ProblemListResponse struct {
	Data struct {
		QuestionList struct {
			Total int `json:"totalNum"`
			Data  []struct {
				TitleSlug  string `json:"titleSlug"`
				Difficulty string `json:"difficulty"`
			} `json:"data"`
		} `json:"questionList"`
	} `json:"data"`
}
