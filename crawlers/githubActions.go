package crawlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type Billing_response struct {
	TotalMinutes     int            `json:"total_minutes_used"`
	TotalPaidMinutes int            `json:"total_paid_minutes_used"`
	IncludedMinutes  int            `json:"included_minutes"`
	Breakdown        map[string]int `json:"minutes_used_breakdown"`
}

func GetGithubActionsUsage() (*Billing_response, error) {
	response := Billing_response{}

	githubOrg := os.Getenv("GITHUB_ORG")
	githubToken := os.Getenv("GITHUB_TOKEN")

	fmt.Printf("Requesting billing information for %s...", githubOrg)
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/orgs/%s/settings/billing/actions", githubOrg), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", githubToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &response)

	return &response, nil

}
