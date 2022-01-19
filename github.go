package bpm

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-github/v36/github"
	"github.com/rs/zerolog"
)

type GithubConfig struct {
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
}

type GithubProvider struct {
	client *github.Client
	logger zerolog.Logger
}

func init() {
	PackageProviders["github.com"] = NewGithubProvider
}

func NewGithubProvider(logger zerolog.Logger, config *Config) PackageProvider {
	logger = logger.With().Str("module", "github").Logger()
	githubConfig := config.Github
	client := &http.Client{}
	if githubConfig.Username != "" {
		if githubConfig.Token == "" {
			logger.Error().Msgf("If github username is set a token must be set!")
		} else {
			client.Transport = newBasicAuthTransport(githubConfig.Username, githubConfig.Token, nil)
		}
		logger.Debug().Msgf("use provided username %s", githubConfig.Username)
	}
	provider := &GithubProvider{
		client: github.NewClient(client),
		logger: logger,
	}
	limits, _, err := provider.client.RateLimits(context.Background())
	if err != nil {
		logger.Err(err).Msgf("cannot get rate limits")
	}
	logger.Debug().Msgf("got rate limits: %d (remaining %d, resets at %s)", limits.Core.Limit, limits.Core.Remaining, limits.Core.Reset.String())
	return provider
}

func (provider *GithubProvider) getLatestRelease(pkg Package) (*github.RepositoryRelease, error) {
	ctx := context.TODO()
	splits := strings.SplitN(pkg.URL, "/", 3)
	if len(splits) < 3 {
		return nil, fmt.Errorf("%w: url (%s) has not the correct github format (github.com/<owner>/<repo>)", ErrProviderConfig, pkg.URL)
	}
	owner := splits[1]
	repoName := splits[2]

	release, _, err := provider.client.Repositories.GetLatestRelease(ctx, owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("%w: cannot get repo: %s", ErrProviderConfig, err)
	}
	return release, nil
}

func (provider *GithubProvider) GetLatest(pkg Package) (version string, err error) {
	release, err := provider.getLatestRelease(pkg)
	if err != nil {
		return "", err
	}

	return release.GetTagName(), err
}
func (provider *GithubProvider) FetchPackage(pkg Package, version string, cacheDir string) (path string, err error) {
	ctx := context.TODO()
	release, err := provider.getLatestRelease(pkg)
	if err != nil {
		return "", err
	}
	assetPattern, err := regexp.Compile(pkg.patternExpand(pkg.AssetPattern, version))
	if err != nil {
		return "", err
	}
	provider.logger.Debug().Msgf("search for pattern %s", assetPattern.String())
	for _, asset := range release.Assets {
		name := asset.GetName()
		provider.logger.Debug().Msgf("try asset %s", name)
		if assetPattern.Match([]byte(name)) {
			url := asset.GetBrowserDownloadURL()
			provider.logger.Debug().Msgf("get asset from %s", url)
			req, err := provider.client.NewRequest("GET", url, nil)
			if err != nil {
				return path, err
			}
			path = filepath.Join(cacheDir, asset.GetName())
			file, err := os.Create(path)
			if err != nil {
				return path, err
			}
			defer file.Close()
			_, err = provider.client.Do(ctx, req, file)
			if err != nil {
				return path, err
			}
			return path, nil
		}
	}
	return path, fmt.Errorf("%w: no matching asset found", ErrProviderFetch)
}
