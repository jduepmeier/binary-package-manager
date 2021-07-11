package bpm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-github/v36/github"
	"github.com/rs/zerolog"
)

type GithubProvider struct {
	client *github.Client
	logger zerolog.Logger
}

func init() {
	PackageProviders["github.com"] = NewGithubProvider
}

func NewGithubProvider(logger zerolog.Logger) PackageProvider {
	return &GithubProvider{
		client: github.NewClient(nil),
		logger: logger.With().Str("module", "github").Logger(),
	}
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
func (provider *GithubProvider) FetchPackage(pkg Package, cacheDir string) (path string, err error) {
	ctx := context.TODO()
	release, err := provider.getLatestRelease(pkg)
	if err != nil {
		return "", err
	}
	assetPattern, err := regexp.Compile(pkg.patternExpand(pkg.AssetPattern))
	if err != nil {
		return "", err
	}
	provider.logger.Debug().Msgf("search for pattern %s", assetPattern.String())
	for _, asset := range release.Assets {
		name := asset.GetName()
		provider.logger.Debug().Msgf("try asset %s", name)
		nameLower := strings.ToLower(name)
		if assetPattern.Match([]byte(nameLower)) {
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
