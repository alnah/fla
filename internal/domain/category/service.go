package category

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/alnah/fla/internal/domain/kernel"
)

// PathService handles URL generation and parsing for hierarchical navigation.
// Enables clean URLs and breadcrumb navigation for educational content structure.
type PathService struct {
	repository Repository
}

// NewPathService creates path service with repository dependency.
// Provides URL management capabilities for category-based content organization.
func NewPathService(repository Repository) *PathService {
	return &PathService{
		repository: repository,
	}
}

// BuildURL generates SEO-friendly URL paths from category hierarchy.
// Creates clean URLs like "a1/comprehension-ecrite/sports" for optimal navigation.
func (s *PathService) BuildURL(categoryID kernel.ID[Category]) (string, error) {
	path, err := s.repository.BuildPath(categoryID)
	if err != nil {
		return "", err
	}

	return path.String(), nil
}

// ParseURL converts URL paths back to category entities for routing.
// Enables dynamic content serving based on hierarchical URL structure.
func (s *PathService) ParseURL(urlPath string) (*Category, error) {
	urlPath = strings.Trim(urlPath, "/")
	if urlPath == "" {
		return nil, errors.New("empty path not supported")
	}

	segments := strings.Split(urlPath, "/")

	for i, segment := range segments {
		decoded, err := url.QueryUnescape(segment)
		if err != nil {
			return nil, fmt.Errorf("invalid URL segment: %s", segment)
		}
		segments[i] = decoded
	}

	return s.repository.FindByPath(segments)
}

// GetBreadcrumbs creates navigation trails for hierarchical content browsing.
// Enables users to understand location and navigate through category levels.
func (s *PathService) GetBreadcrumbs(categoryID kernel.ID[Category]) ([]CategoryBreadcrumb, error) {
	path, err := s.repository.BuildPath(categoryID)
	if err != nil {
		return nil, err
	}

	breadcrumbs := make([]CategoryBreadcrumb, len(path))
	for i, category := range path {
		breadcrumbs[i] = CategoryBreadcrumb{
			Category: category,
			IsLast:   i == len(path)-1,
			Level:    i,
		}
	}

	return breadcrumbs, nil
}
