package api

import (
    "context"
    "fmt"
    "time"
    "net/url"

    "github.com/google/go-github/v66/github"
    "github.com/spf13/viper"
    "golang.org/x/oauth2"
)

const maxRetries = 3

func newGitHubClient(token string) (*github.Client, error) {
    if token == "" {
        return nil, fmt.Errorf("GitHub token is required")
    }

    ctx := context.Background()
    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: token},
    )
    tc := oauth2.NewClient(ctx, ts)
    
    hostname := viper.GetString("SOURCE_HOSTNAME")
    
    if hostname == "" {
        return github.NewClient(tc), nil
    }

    baseURL, err := url.Parse(hostname)
    if err != nil {
        return nil, fmt.Errorf("invalid hostname URL provided (%s): %w", hostname, err)
    }

    if baseURL.Path == "" {
        baseURL.Path = "/"
    }

    client, err := github.NewEnterpriseClient(baseURL.String(), baseURL.String(), tc)
    if err != nil {
        return nil, fmt.Errorf("failed to create Enterprise client for %s: %w", hostname, err)
    }

    return client, nil
}

// RepoExists checks if a repository exists in the organization
func RepoExists(org, repo, token string) (bool, error) {
    client, err := newGitHubClient(token)
    if err != nil {
        return false, fmt.Errorf("failed to initialize GitHub client: %w", err)
    }
    
    _, resp, err := client.Repositories.Get(context.Background(), org, repo)
    if err != nil {
        return false, nil
    }
    return resp.StatusCode == 200, nil
}

// GetOrgVariables fetches all organization variables
func GetOrgVariables(org, token string) ([]map[string]string, error) {
    if org == "" {
        return nil, fmt.Errorf("organization name is required")
    }

    client, err := newGitHubClient(token)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize GitHub client: %w", err)
    }

    var variables *github.ActionsVariables
    var apiErr error

    for attempt := 1; attempt <= maxRetries; attempt++ {
        variables, _, apiErr = client.Actions.ListOrgVariables(context.Background(), org, nil)
        if apiErr == nil && variables != nil {
            break
        }

        if attempt < maxRetries {
            waitTime := time.Duration(1<<uint(attempt-1)) * time.Second
            fmt.Printf("Attempt %d failed for org %s, retrying in %v: %v\n", 
                attempt, org, waitTime, apiErr)
            time.Sleep(waitTime)
        }
    }

    if apiErr != nil {
        return nil, fmt.Errorf("failed to fetch organization variables after %d attempts: %w", 
            maxRetries, apiErr)
    }

    if variables == nil {
        return nil, fmt.Errorf("no variables data returned for organization %s", org)
    }

    var orgVariables []map[string]string
    for _, variable := range variables.Variables {
        if variable == nil {
            continue
        }

        visibility := "private"
        if variable.Visibility != nil {
            visibility = *variable.Visibility
        }

        if variable.Name != "" {
            orgVariables = append(orgVariables, map[string]string{
                "Name":       variable.Name,
                "Value":      variable.Value,
                "Scope":      "organization",
                "Visibility": visibility,
            })
        }
    }

    return orgVariables, nil
}

// GetRepoVariables fetches all repository variables
func GetRepoVariables(org, repo, token string) ([]map[string]string, error) {
    if org == "" || repo == "" {
        return nil, fmt.Errorf("organization name and repository name are required")
    }

    client, err := newGitHubClient(token)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize GitHub client: %w", err)
    }

    var variables *github.ActionsVariables
    var apiErr error

    for attempt := 1; attempt <= maxRetries; attempt++ {
        variables, _, apiErr = client.Actions.ListRepoVariables(context.Background(), org, repo, nil)
        if apiErr == nil && variables != nil {
            break
        }

        if attempt < maxRetries {
            waitTime := time.Duration(1<<uint(attempt-1)) * time.Second
            fmt.Printf("Attempt %d failed for repo %s, retrying in %v: %v\n", 
                attempt, repo, waitTime, apiErr)
            time.Sleep(waitTime)
        }
    }

    if apiErr != nil {
        return nil, fmt.Errorf("failed to fetch repository variables for %s after %d attempts: %w", 
            repo, maxRetries, apiErr)
    }

    if variables == nil {
        return nil, fmt.Errorf("no variables data returned for repository %s", repo)
    }

    var repoVariables []map[string]string
    for _, variable := range variables.Variables {
        if variable == nil {
            continue
        }

        visibility := "private"
        if variable.Visibility != nil {
            visibility = *variable.Visibility
        }

        if variable.Name != "" {
            repoVariables = append(repoVariables, map[string]string{
                "Name":       variable.Name,
                "Value":      variable.Value,
                "Scope":      repo,
                "Visibility": visibility,
            })
        }
    }

    return repoVariables, nil
}

// GetRepositories fetches all repositories for a given organization
func GetRepositories(org, token string) ([]string, error) {
    if org == "" {
        return nil, fmt.Errorf("organization name is required")
    }

    client, err := newGitHubClient(token)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize GitHub client: %w", err)
    }

    var allRepos []string
    opts := &github.RepositoryListByOrgOptions{
        ListOptions: github.ListOptions{PerPage: 100},
    }

    for {
        repos, resp, err := client.Repositories.ListByOrg(context.Background(), org, opts)
        if err != nil {
            return nil, fmt.Errorf("failed to list repositories for %s: %w", org, err)
        }

        if repos == nil {
            return nil, fmt.Errorf("no repositories data returned for organization %s", org)
        }

        for _, repo := range repos {
            if repo != nil && repo.Name != nil {
                allRepos = append(allRepos, *repo.Name)
            }
        }

        if resp == nil || resp.NextPage == 0 {
            break
        }
        opts.Page = resp.NextPage
    }

    return allRepos, nil
}

// CreateOrgVariable creates an organization variable in the target org
func CreateOrgVariable(org, name, value, visibility, token string) error {
    if org == "" || name == "" {
        return fmt.Errorf("organization name and variable name are required")
    }

    client, err := newGitHubClient(token)
    if err != nil {
        return fmt.Errorf("failed to initialize GitHub client: %w", err)
    }

    if visibility == "" {
        visibility = "private"
    }

    variable := &github.ActionsVariable{
        Name:       name,
        Value:      value,
        Visibility: github.String(visibility),
    }

    _, err = client.Actions.CreateOrgVariable(context.Background(), org, variable)
    if err != nil {
        return fmt.Errorf("failed to create org variable %s: %w", name, err)
    }
    return nil
}

// CreateRepoVariable creates a repository variable in the target repo
func CreateRepoVariable(org, repo, name, value, visibility, token string) error {
    if org == "" || repo == "" || name == "" {
        return fmt.Errorf("organization name, repository name, and variable name are required")
    }

    exists, err := RepoExists(org, repo, token)
    if err != nil {
        return fmt.Errorf("failed to check repository existence: %w", err)
    }
    if !exists {
        return fmt.Errorf("repository %s does not exist in organization %s", repo, org)
    }

    client, err := newGitHubClient(token)
    if err != nil {
        return fmt.Errorf("failed to initialize GitHub client: %w", err)
    }

    if visibility == "" {
        visibility = "private"
    }

    variable := &github.ActionsVariable{
        Name:       name,
        Value:      value,
        Visibility: github.String(visibility),
    }

    _, err = client.Actions.CreateRepoVariable(context.Background(), org, repo, variable)
    if err != nil {
        return fmt.Errorf("failed to create repo variable %s in repo %s: %w", name, repo, err)
    }
    return nil
}