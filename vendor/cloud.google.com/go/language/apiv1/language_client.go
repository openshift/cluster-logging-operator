// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go_gapic. DO NOT EDIT.

package language

import (
	"context"
	"math"
	"time"

	gax "github.com/googleapis/gax-go/v2"
	"google.golang.org/api/option"
	gtransport "google.golang.org/api/transport/grpc"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

// CallOptions contains the retry settings for each method of Client.
type CallOptions struct {
	AnalyzeSentiment       []gax.CallOption
	AnalyzeEntities        []gax.CallOption
	AnalyzeEntitySentiment []gax.CallOption
	AnalyzeSyntax          []gax.CallOption
	ClassifyText           []gax.CallOption
	AnnotateText           []gax.CallOption
}

func defaultClientOptions() []option.ClientOption {
	return []option.ClientOption{
		option.WithEndpoint("language.googleapis.com:443"),
		option.WithGRPCDialOption(grpc.WithDisableServiceConfig()),
		option.WithScopes(DefaultAuthScopes()...),
		option.WithGRPCDialOption(grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(math.MaxInt32))),
	}
}

func defaultCallOptions() *CallOptions {
	return &CallOptions{
		AnalyzeSentiment: []gax.CallOption{
			gax.WithRetry(func() gax.Retryer {
				return gax.OnCodes([]codes.Code{
					codes.DeadlineExceeded,
					codes.Unavailable,
				}, gax.Backoff{
					Initial:    100 * time.Millisecond,
					Max:        60000 * time.Millisecond,
					Multiplier: 1.30,
				})
			}),
		},
		AnalyzeEntities: []gax.CallOption{
			gax.WithRetry(func() gax.Retryer {
				return gax.OnCodes([]codes.Code{
					codes.DeadlineExceeded,
					codes.Unavailable,
				}, gax.Backoff{
					Initial:    100 * time.Millisecond,
					Max:        60000 * time.Millisecond,
					Multiplier: 1.30,
				})
			}),
		},
		AnalyzeEntitySentiment: []gax.CallOption{
			gax.WithRetry(func() gax.Retryer {
				return gax.OnCodes([]codes.Code{
					codes.DeadlineExceeded,
					codes.Unavailable,
				}, gax.Backoff{
					Initial:    100 * time.Millisecond,
					Max:        60000 * time.Millisecond,
					Multiplier: 1.30,
				})
			}),
		},
		AnalyzeSyntax: []gax.CallOption{
			gax.WithRetry(func() gax.Retryer {
				return gax.OnCodes([]codes.Code{
					codes.DeadlineExceeded,
					codes.Unavailable,
				}, gax.Backoff{
					Initial:    100 * time.Millisecond,
					Max:        60000 * time.Millisecond,
					Multiplier: 1.30,
				})
			}),
		},
		ClassifyText: []gax.CallOption{
			gax.WithRetry(func() gax.Retryer {
				return gax.OnCodes([]codes.Code{
					codes.DeadlineExceeded,
					codes.Unavailable,
				}, gax.Backoff{
					Initial:    100 * time.Millisecond,
					Max:        60000 * time.Millisecond,
					Multiplier: 1.30,
				})
			}),
		},
		AnnotateText: []gax.CallOption{
			gax.WithRetry(func() gax.Retryer {
				return gax.OnCodes([]codes.Code{
					codes.DeadlineExceeded,
					codes.Unavailable,
				}, gax.Backoff{
					Initial:    100 * time.Millisecond,
					Max:        60000 * time.Millisecond,
					Multiplier: 1.30,
				})
			}),
		},
	}
}

// Client is a client for interacting with Cloud Natural Language API.
//
// Methods, except Close, may be called concurrently. However, fields must not be modified concurrently with method calls.
type Client struct {
	// Connection pool of gRPC connections to the service.
	connPool gtransport.ConnPool

	// The gRPC API client.
	client languagepb.LanguageServiceClient

	// The call options for this service.
	CallOptions *CallOptions

	// The x-goog-* metadata to be sent with each request.
	xGoogMetadata metadata.MD
}

// NewClient creates a new language service client.
//
// Provides text analysis operations such as sentiment analysis and entity
// recognition.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	connPool, err := gtransport.DialPool(ctx, append(defaultClientOptions(), opts...)...)
	if err != nil {
		return nil, err
	}
	c := &Client{
		connPool:    connPool,
		CallOptions: defaultCallOptions(),

		client: languagepb.NewLanguageServiceClient(connPool),
	}
	c.setGoogleClientInfo()

	return c, nil
}

// Connection returns a connection to the API service.
//
// Deprecated.
func (c *Client) Connection() *grpc.ClientConn {
	return c.connPool.Conn()
}

// Close closes the connection to the API service. The user should invoke this when
// the client is no longer required.
func (c *Client) Close() error {
	return c.connPool.Close()
}

// setGoogleClientInfo sets the name and version of the application in
// the `x-goog-api-client` header passed on each request. Intended for
// use by Google-written clients.
func (c *Client) setGoogleClientInfo(keyval ...string) {
	kv := append([]string{"gl-go", versionGo()}, keyval...)
	kv = append(kv, "gapic", versionClient, "gax", gax.Version, "grpc", grpc.Version)
	c.xGoogMetadata = metadata.Pairs("x-goog-api-client", gax.XGoogHeader(kv...))
}

// AnalyzeSentiment analyzes the sentiment of the provided text.
func (c *Client) AnalyzeSentiment(ctx context.Context, req *languagepb.AnalyzeSentimentRequest, opts ...gax.CallOption) (*languagepb.AnalyzeSentimentResponse, error) {
	ctx = insertMetadata(ctx, c.xGoogMetadata)
	opts = append(c.CallOptions.AnalyzeSentiment[0:len(c.CallOptions.AnalyzeSentiment):len(c.CallOptions.AnalyzeSentiment)], opts...)
	var resp *languagepb.AnalyzeSentimentResponse
	err := gax.Invoke(ctx, func(ctx context.Context, settings gax.CallSettings) error {
		var err error
		resp, err = c.client.AnalyzeSentiment(ctx, req, settings.GRPC...)
		return err
	}, opts...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// AnalyzeEntities finds named entities (currently proper names and common nouns) in the text
// along with entity types, salience, mentions for each entity, and
// other properties.
func (c *Client) AnalyzeEntities(ctx context.Context, req *languagepb.AnalyzeEntitiesRequest, opts ...gax.CallOption) (*languagepb.AnalyzeEntitiesResponse, error) {
	ctx = insertMetadata(ctx, c.xGoogMetadata)
	opts = append(c.CallOptions.AnalyzeEntities[0:len(c.CallOptions.AnalyzeEntities):len(c.CallOptions.AnalyzeEntities)], opts...)
	var resp *languagepb.AnalyzeEntitiesResponse
	err := gax.Invoke(ctx, func(ctx context.Context, settings gax.CallSettings) error {
		var err error
		resp, err = c.client.AnalyzeEntities(ctx, req, settings.GRPC...)
		return err
	}, opts...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// AnalyzeEntitySentiment finds entities, similar to [AnalyzeEntities][google.cloud.language.v1.LanguageService.AnalyzeEntities] in the text and analyzes
// sentiment associated with each entity and its mentions.
func (c *Client) AnalyzeEntitySentiment(ctx context.Context, req *languagepb.AnalyzeEntitySentimentRequest, opts ...gax.CallOption) (*languagepb.AnalyzeEntitySentimentResponse, error) {
	ctx = insertMetadata(ctx, c.xGoogMetadata)
	opts = append(c.CallOptions.AnalyzeEntitySentiment[0:len(c.CallOptions.AnalyzeEntitySentiment):len(c.CallOptions.AnalyzeEntitySentiment)], opts...)
	var resp *languagepb.AnalyzeEntitySentimentResponse
	err := gax.Invoke(ctx, func(ctx context.Context, settings gax.CallSettings) error {
		var err error
		resp, err = c.client.AnalyzeEntitySentiment(ctx, req, settings.GRPC...)
		return err
	}, opts...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// AnalyzeSyntax analyzes the syntax of the text and provides sentence boundaries and
// tokenization along with part of speech tags, dependency trees, and other
// properties.
func (c *Client) AnalyzeSyntax(ctx context.Context, req *languagepb.AnalyzeSyntaxRequest, opts ...gax.CallOption) (*languagepb.AnalyzeSyntaxResponse, error) {
	ctx = insertMetadata(ctx, c.xGoogMetadata)
	opts = append(c.CallOptions.AnalyzeSyntax[0:len(c.CallOptions.AnalyzeSyntax):len(c.CallOptions.AnalyzeSyntax)], opts...)
	var resp *languagepb.AnalyzeSyntaxResponse
	err := gax.Invoke(ctx, func(ctx context.Context, settings gax.CallSettings) error {
		var err error
		resp, err = c.client.AnalyzeSyntax(ctx, req, settings.GRPC...)
		return err
	}, opts...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ClassifyText classifies a document into categories.
func (c *Client) ClassifyText(ctx context.Context, req *languagepb.ClassifyTextRequest, opts ...gax.CallOption) (*languagepb.ClassifyTextResponse, error) {
	ctx = insertMetadata(ctx, c.xGoogMetadata)
	opts = append(c.CallOptions.ClassifyText[0:len(c.CallOptions.ClassifyText):len(c.CallOptions.ClassifyText)], opts...)
	var resp *languagepb.ClassifyTextResponse
	err := gax.Invoke(ctx, func(ctx context.Context, settings gax.CallSettings) error {
		var err error
		resp, err = c.client.ClassifyText(ctx, req, settings.GRPC...)
		return err
	}, opts...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// AnnotateText a convenience method that provides all the features that analyzeSentiment,
// analyzeEntities, and analyzeSyntax provide in one call.
func (c *Client) AnnotateText(ctx context.Context, req *languagepb.AnnotateTextRequest, opts ...gax.CallOption) (*languagepb.AnnotateTextResponse, error) {
	ctx = insertMetadata(ctx, c.xGoogMetadata)
	opts = append(c.CallOptions.AnnotateText[0:len(c.CallOptions.AnnotateText):len(c.CallOptions.AnnotateText)], opts...)
	var resp *languagepb.AnnotateTextResponse
	err := gax.Invoke(ctx, func(ctx context.Context, settings gax.CallSettings) error {
		var err error
		resp, err = c.client.AnnotateText(ctx, req, settings.GRPC...)
		return err
	}, opts...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
