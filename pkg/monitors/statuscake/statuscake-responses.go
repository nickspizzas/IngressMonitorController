package statuscake

import (
	statuscake "github.com/StatusCakeDev/statuscake-go"
)

// StatusCakeMonitor response Structure for GetAll and GetByName API's for Statuscake

type StatusCakeMonitor struct {
	StatusCakeUptimeTestData     []StatusCakeUptimeTestData   `json:"data"`
	StatusCakeUptimeTestMetadata StatusCakeUptimeTestMetadata `json:"metadata"`
}

type StatusCakeUptimeTestMetadata struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	PageCount  int `json:"page_count"`
	TotalCount int `json:"total_count"`
}

type StatusCakeUptimeTestData struct {
	statuscake.UptimeTest
}

type StatusCakeUptimeTestOverviewData struct {
	statuscake.UptimeTestOverview
}
