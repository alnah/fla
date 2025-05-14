package aiclient

import (
	"context"
	"net/http"

	"github.com/alnah/fla/logger"
)

func WithContext[T hasBase](c context.Context) Option[T] {
	return func(t T) { (t).BaseClient().ctx = c }
}
func WithProvider[T hasBase](p Provider) Option[T] {
	return func(t T) { (t).BaseClient().provider = p }
}
func WithURL[T hasBase](u URL) Option[T] {
	return func(t T) { (t).BaseClient().url = u }
}
func WithAPIKey[T hasBase](a APIKey) Option[T] {
	return func(t T) { (t).BaseClient().apiKey = a }
}
func WithModel[T hasBase](a AIModel) Option[T] {
	return func(t T) { (t).BaseClient().Model = a }
}
func WithHTTPClient[T hasBase](hc *http.Client) Option[T] {
	return func(t T) { (t).BaseClient().httpClient = hc }
}
func WithHTTPMethod[T hasBase](hm HTTPMethod) Option[T] {
	return func(t T) { (t).BaseClient().httpMethod = hm }
}
func WithLogger[T hasBase](l *logger.Logger) Option[T] {
	return func(t T) { (t).BaseClient().logger = l }
}
