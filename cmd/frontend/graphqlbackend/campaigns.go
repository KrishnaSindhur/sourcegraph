package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// Campaigns is the implementation of the GraphQL campaigns queries and mutations. If it is not set
// at runtime, a "not implemented" error is returned to API clients who invoke it.
//
// This is contributed by enterprise.
var Campaigns CampaignsResolver

var errCampaignsNotImplemented = errors.New("campaigns is not implemented")

// CampaignByID is called to look up a Campaign given its GraphQL ID.
func CampaignByID(ctx context.Context, id graphql.ID) (Campaign, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.CampaignByID(ctx, id)
}

// CampaignsInNamespace returns an instance of the GraphQL CampaignConnection type with the list of
// campaigns defined in a namespace.
func CampaignsInNamespace(ctx context.Context, namespace graphql.ID, arg *graphqlutil.ConnectionArgs) (CampaignConnection, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.CampaignsInNamespace(ctx, namespace, arg)
}

func (schemaResolver) Campaigns(ctx context.Context, arg *graphqlutil.ConnectionArgs) (CampaignConnection, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.Campaigns(ctx, arg)
}

func (r schemaResolver) CreateCampaign(ctx context.Context, arg *CreateCampaignArgs) (Campaign, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.CreateCampaign(ctx, arg)
}

func (r schemaResolver) UpdateCampaign(ctx context.Context, arg *UpdateCampaignArgs) (Campaign, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.UpdateCampaign(ctx, arg)
}

func (r schemaResolver) PublishPreviewCampaign(ctx context.Context, arg *PublishPreviewCampaignArgs) (Campaign, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.PublishPreviewCampaign(ctx, arg)
}

func (r schemaResolver) DeleteCampaign(ctx context.Context, arg *DeleteCampaignArgs) (*EmptyResponse, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.DeleteCampaign(ctx, arg)
}

func (r schemaResolver) AddThreadsToCampaign(ctx context.Context, arg *AddRemoveThreadsToFromCampaignArgs) (*EmptyResponse, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.AddThreadsToCampaign(ctx, arg)
}

func (r schemaResolver) RemoveThreadsFromCampaign(ctx context.Context, arg *AddRemoveThreadsToFromCampaignArgs) (*EmptyResponse, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.RemoveThreadsFromCampaign(ctx, arg)
}

// CampaignsResolver is the interface for the GraphQL campaigns queries and mutations.
type CampaignsResolver interface {
	// Queries
	Campaigns(context.Context, *graphqlutil.ConnectionArgs) (CampaignConnection, error)

	// Mutations
	CreateCampaign(context.Context, *CreateCampaignArgs) (Campaign, error)
	UpdateCampaign(context.Context, *UpdateCampaignArgs) (Campaign, error)
	PublishPreviewCampaign(context.Context, *PublishPreviewCampaignArgs) (Campaign, error)
	DeleteCampaign(context.Context, *DeleteCampaignArgs) (*EmptyResponse, error)
	AddThreadsToCampaign(context.Context, *AddRemoveThreadsToFromCampaignArgs) (*EmptyResponse, error)
	RemoveThreadsFromCampaign(context.Context, *AddRemoveThreadsToFromCampaignArgs) (*EmptyResponse, error)

	// CampaignByID is called by the CampaignByID func but is not in the GraphQL API.
	CampaignByID(context.Context, graphql.ID) (Campaign, error)

	// CampaignsInNamespace is called by the CampaignsInNamespace func but is not in the GraphQL
	// API.
	CampaignsInNamespace(ctx context.Context, namespace graphql.ID, arg *graphqlutil.ConnectionArgs) (CampaignConnection, error)
}

type CreateCampaignArgs struct {
	Input struct {
		Namespace   graphql.ID
		Name        string
		Description *string
		Preview     *bool
		Rules       *string
	}
}

type UpdateCampaignArgs struct {
	Input struct {
		ID          graphql.ID
		Name        *string
		Description *string
		Rules       *string
	}
}

type PublishPreviewCampaignArgs struct {
	Campaign graphql.ID
}

type DeleteCampaignArgs struct {
	Campaign graphql.ID
}

type AddRemoveThreadsToFromCampaignArgs struct {
	Campaign graphql.ID
	Threads  []graphql.ID
}

// Campaign is the interface for the GraphQL type Campaign.
type Campaign interface {
	ID() graphql.ID
	Namespace(context.Context) (*NamespaceResolver, error)
	Name() string
	Description() *string
	IsPreview() bool
	Rules() string
	URL(context.Context) (string, error)
	ThreadOrIssueOrChangesets(context.Context, *graphqlutil.ConnectionArgs) (ThreadOrIssueOrChangesetConnection, error)
	Repositories(context.Context) ([]*RepositoryResolver, error)
	Commits(context.Context) ([]*GitCommitResolver, error)
	RepositoryComparisons(context.Context) ([]*RepositoryComparisonResolver, error)
}

// CampaignConnection is the interface for the GraphQL type CampaignConnection.
type CampaignConnection interface {
	Nodes(context.Context) ([]Campaign, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}