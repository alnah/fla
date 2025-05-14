package aiclient

import (
	"context"
	"net/http"

	"github.com/alnah/fla/logger"
)

type Option func(*ChatClient)

func WithContext(c context.Context) Option  { return func(cc *ChatClient) { cc.ctx = c } }
func WithProvider(p Provider) Option        { return func(cc *ChatClient) { cc.provider = p } }
func WithURL(u URL) Option                  { return func(cc *ChatClient) { cc.url = u } }
func WithAPIKey(a APIKey) Option            { return func(cc *ChatClient) { cc.apiKey = a } }
func WithModel(a AIModel) Option            { return func(cc *ChatClient) { cc.Model = a } }
func WithTemperature(t Temperature) Option  { return func(cc *ChatClient) { cc.Temperature = t } }
func WithMessages(ms Messages) Option       { return func(cc *ChatClient) { cc.Messages = ms } }
func WithMaxTokens(mt MaxTokens) Option     { return func(cc *ChatClient) { cc.MaxTokens = mt } }
func WithHTTPClient(hc *http.Client) Option { return func(cc *ChatClient) { cc.httpClient = hc } }
func WithHTTPMethod(hm HTTPMethod) Option   { return func(cc *ChatClient) { cc.httpMethod = hm } }
func WithLogger(l *logger.Logger) Option    { return func(cc *ChatClient) { cc.logger = l } }
