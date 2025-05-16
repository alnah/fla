package aiclient

import (
	"context"
	"net/http"

	"github.com/alnah/fla/logger"
)

func WithContext[T hasBase](c context.Context) option[T] {
	return func(t T) { (t).BaseClient().ctx = c }
}
func WithLogger[T hasBase](l *logger.Logger) option[T] {
	return func(t T) { (t).BaseClient().logger = l }
}
func WithProvider[T hasBase](p provider) option[T] {
	return func(t T) { (t).BaseClient().provider = p }
}
func WithURL[T hasBase](u url) option[T] {
	return func(t T) { (t).BaseClient().url = u }
}
func WithAPIKey[T hasBase](a apiKey) option[T] {
	return func(t T) { (t).BaseClient().apiKey = a }
}
func WithModel[T hasBase](a model) option[T] {
	return func(t T) { (t).BaseClient().model = a }
}
func WithHTTPClient[T hasBase](hc *http.Client) option[T] {
	return func(t T) { (t).BaseClient().httpClient = hc }
}
func WithHTTPMethod[T hasBase](hm httpMethod) option[T] {
	return func(t T) { (t).BaseClient().httpMethod = hm }
}
