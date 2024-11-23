package api

import (
    "context"
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "time"

    "github.com/google/go-github/v66/github"
    "github.com/spf13/viper"
    "golang.org/x/oauth2"
)

const maxRetries = 3

type ProxyConfig struct {
    HTTPProxy  string
    HTTPSProxy string
    NoProxy    string
}

// Helper function to handle optional hostname parameter
func getHostname(hostname ...string) string {
    if len(hostname) > 0 {
        return hostname[0]
    }
    return ""
}

func newGitHubClientWithHostname(token string, hostname string) (*github.Client, error) {
    // First create the base client with proxy support
    client, err := newGitHubClientWithProxy(token, GetProxyConfigFromEnv())
    if err != nil {
        return nil, err
    }

    // If no hostname provided, return the base client
    if hostname == "" {
        return client, nil
    }

    baseURL, err := url.Parse(hostname)
    if err != nil {
        return nil, fmt.Errorf("invalid hostname URL provided (%s): %w", baseURL, err)
    }

    enterpriseClient, err := client.WithEnterpriseURLs(hostname, hostname)
    if err != nil {
        return nil, fmt.Errorf("failed to configure enterprise URLs for %s: %w", hostname, err)
    }
    
    return enterpriseClient, nil
}

func newGitHubClientWithProxy(token string, proxyConfig *ProxyConfig) (*github.Client, error) {
    if token == "" {
        return nil, fmt.Errorf("GitHub token is required")
    }

    ctx := context.Background()
    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: token},
    )

    // Create transport with proxy support
    transport := &http.Transport{
        Proxy: func(req *http.Request) (*url.URL, error) {
            // Skip proxy if host matches no_proxy
            if proxyConfig != nil && proxyConfig.NoProxy != "" {
                noProxyURLs := strings.Split(proxyConfig.NoProxy, ",")
                reqHost := req.URL.Host
                for _, noProxy := range noProxyURLs {
                    if strings.TrimSpace(noProxy) == reqHost {
                        return nil, nil
                    }
                }
            }

            // Use HTTPS_PROXY for https requests and HTTP_PROXY for http
            if proxyConfig != nil {
                if req.URL.Scheme == "https" && proxyConfig.HTTPSProxy != "" {
                    return url.Parse(proxyConfig.HTTPSProxy)
                }
                if req.URL.Scheme == "http" && proxyConfig.HTTPProxy != "" {
                    return url.Parse(proxyConfig.HTTPProxy)
                }
            }
            return nil, nil
        },
    }

    // Create OAuth2 client with custom transport using the context
    tc := oauth2.NewClient(ctx, ts)
    tc.Transport = &oauth2.Transport{
        Base:   transport,
        Source: ts,
    }

    return github.NewClient(tc), nil
}

// GetProxyConfigFromEnv retrieves proxy configuration from environment variables
func GetProxyConfigFromEnv() *ProxyConfig {
    return &ProxyConfig{
        HTTPProxy:  viper.GetString("HTTP_PROXY"),
        HTTPSProxy: viper.GetString("HTTPS_PROXY"),
        NoProxy:    viper.GetString("NO_PROXY"),
    }
}

func retryOperation(operation func() error) error {
    var apiErr error
    for attempt := 1; attempt <= maxRetries; attempt++ {
        apiErr = operation()
        if apiErr == nil {
            return nil
        }

        if attempt < maxRetries {
            waitTime := time.Duration(1<<uint(attempt-1)) * time.Second
            fmt.Printf("Attempt %d failed, retrying in %v: %v\n", attempt, waitTime, apiErr)
            time.Sleep(waitTime)
        }
    }
    return apiErr
}

func parseVariable(variable *github.ActionsVariable, scope string) map[string]string {
    if variable == nil {
        return nil
    }

    visibility := "private"
    if variable.Visibility != nil {
        visibility = *variable.Visibility
    }

    if variable.Name == "" {
        return nil
    }

    return map[string]string{
        "Name":       variable.Name,
        "Value":      variable.Value,
        "Scope":      scope,
        "Visibility": visibility,
    }
}

// RepoExists checks if a repository exists in the organization
func RepoExists(org, repo, token string, hostname ...string) (bool, error) {
    client, err := newGitHubClientWithHostname(token, getHostname(hostname...))
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
func GetOrgVariables(org, token string, hostname ...string) ([]map[string]string, error) {
    if org == "" {
        return nil, fmt.Errorf("organization name is required")
    }

    client, err := newGitHubClientWithHostname(token, getHostname(hostname...))
    if err != nil {
        return nil, fmt.Errorf("failed to initialize GitHub client: %w", err)
    }

    var variables *github.ActionsVariables
    err = retryOperation(func() error {
        var apiErr error
        variables, _, apiErr = client.Actions.ListOrgVariables(context.Background(), org, nil)
        return apiErr
    })

    if err != nil {
        return nil, fmt.Errorf("failed to fetch organization variables after %d attempts: %w", maxRetries, err)
    }

    if variables == nil {
        return nil, fmt.Errorf("no variables data returned for organization %s", org)
    }

    var orgVariables []map[string]string
    for _, variable := range variables.Variables {
        parsedVar := parseVariable(variable, "organization")
        if parsedVar != nil {
            orgVariables = append(orgVariables, parsedVar)
        }
    }

    return orgVariables, nil
}

// GetRepoVariables fetches all repository variables
func GetRepoVariables(org, repo, token string, hostname ...string) ([]map[string]string, error) {
    if org == "" || repo == "" {
        return nil, fmt.Errorf("organization name and repository name are required")
    }

    client, err := newGitHubClientWithHostname(token, getHostname(hostname...))
    if err != nil {
        return nil, fmt.Errorf("failed to initialize GitHub client: %w", err)
    }

    var variables *github.ActionsVariables
    err = retryOperation(func() error {
        var apiErr error
        variables, _, apiErr = client.Actions.ListRepoVariables(context.Background(), org, repo, nil)
        return apiErr
    })

    if err != nil {
        return nil, fmt.Errorf("failed to fetch repository variables for %s after %d attempts: %w", repo, maxRetries, err)
    }

    if variables == nil {
        return nil, fmt.Errorf("no variables data returned for repository %s", repo)
    }

    var repoVariables []map[string]string
    for _, variable := range variables.Variables {
        parsedVar := parseVariable(variable, repo)
        if parsedVar != nil {
            repoVariables = append(repoVariables, parsedVar)
        }
    }

    return repoVariables, nil
}

// GetRepositories fetches all repositories for a given organization
func GetRepositories(org, token string, hostname ...string) ([]string, error) {
    if org == "" {
        return nil, fmt.Errorf("organization name is required")
    }

    client, err := newGitHubClientWithHostname(token, getHostname(hostname...))
    if err != nil {
        return nil, fmt.Errorf("failed to initialize GitHub client: %w", err)
    }

    var allRepos []string
    opts := &github.RepositoryListByOrgOptions{
        ListOptions: github.ListOptions{PerPage: 100},
    }

    err = retryOperation(func() error {
        for {
            repos, resp, apiErr := client.Repositories.ListByOrg(context.Background(), org, opts)
            if apiErr != nil {
                return apiErr
            }

            if repos == nil {
                return fmt.Errorf("no repositories data returned for organization %s", org)
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
        return nil
    })

    if err != nil {
        return nil, fmt.Errorf("failed to list repositories for %s: %w", org, err)
    }

    return allRepos, nil
}

// CreateOrgVariable creates an organization variable in the target org
func CreateOrgVariable(org, name, value, visibility, token string, hostname ...string) error {
    if org == "" || name == "" {
        return fmt.Errorf("organization name and variable name are required")
    }

    client, err := newGitHubClientWithHostname(token, getHostname(hostname...))
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

    err = retryOperation(func() error {
        _, apiErr := client.Actions.CreateOrgVariable(context.Background(), org, variable)
        return apiErr
    })

    if err != nil {
        return fmt.Errorf("failed to create org variable %s after %d attempts: %w", name, maxRetries, err)
    }

    return nil
}

// CreateRepoVariable creates a repository variable in the target repo
func CreateRepoVariable(org, repo, name, value, visibility, token string, hostname ...string) error {
    if org == "" || repo == "" || name == "" {
        return fmt.Errorf("organization name, repository name, and variable name are required")
    }

    exists, err := RepoExists(org, repo, token, getHostname(hostname...))
    if err != nil {
        return fmt.Errorf("failed to check repository existence: %w", err)
    }
    if !exists {
        return fmt.Errorf("repository %s does not exist in organization %s", repo, org)
    }

    client, err := newGitHubClientWithHostname(token, getHostname(hostname...))
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

    err = retryOperation(func() error {
        _, apiErr := client.Actions.CreateRepoVariable(context.Background(), org, repo, variable)
        return apiErr
    })

    if err != nil {
        return fmt.Errorf("failed to create repo variable %s in repo %s after %d attempts: %w", name, repo, maxRetries, err)
    }

    return nil
}