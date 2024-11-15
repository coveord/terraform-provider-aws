// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package amp

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amp"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*types.ServicePackageFrameworkDataSource {
	return []*types.ServicePackageFrameworkDataSource{
		{
			Factory: newDefaultScraperConfigurationDataSource,
			Name:    "Default Scraper Configuration",
		},
	}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*types.ServicePackageFrameworkResource {
	return []*types.ServicePackageFrameworkResource{
		{
			Factory: newScraperResource,
			Name:    "Scraper",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
	}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*types.ServicePackageSDKDataSource {
	return []*types.ServicePackageSDKDataSource{
		{
			Factory:  dataSourceWorkspace,
			TypeName: "aws_prometheus_workspace",
			Name:     "Workspace",
			Tags:     &types.ServicePackageResourceTags{},
		},
		{
			Factory:  dataSourceWorkspaces,
			TypeName: "aws_prometheus_workspaces",
			Name:     "Workspaces",
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{
		{
			Factory:  resourceAlertManagerDefinition,
			TypeName: "aws_prometheus_alert_manager_definition",
			Name:     "Alert Manager Definition",
		},
		{
			Factory:  resourceRuleGroupNamespace,
			TypeName: "aws_prometheus_rule_group_namespace",
			Name:     "Rule Group Namespace",
		},
		{
			Factory:  resourceWorkspace,
			TypeName: "aws_prometheus_workspace",
			Name:     "Workspace",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.AMP
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*amp.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return amp.NewFromConfig(cfg,
		amp.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
	), nil
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
