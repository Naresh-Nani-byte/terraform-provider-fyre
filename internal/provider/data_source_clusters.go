// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp-forge/terraform-provider-fyre/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &DataSourceClusters{}

// NewDataSourceClusters creates a new clusters data source.
func NewDataSourceClusters() datasource.DataSource {
	return &DataSourceClusters{}
}

// DataSourceClusters defines the clusters data source implementation.
type DataSourceClusters struct {
	client      *client.ClientWithResponses
	defaultSite string
}

// ClustersModel describes the data source data model.
type ClustersModel struct {
	ID           types.String `tfsdk:"id"`
	Site         types.String `tfsdk:"site"`
	ClusterCount types.Int64  `tfsdk:"cluster_count"`
	Clusters     types.List   `tfsdk:"clusters"`
}

// ClusterModel describes a single cluster in the list.
type ClusterModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Created     types.String `tfsdk:"created"`
	Updated     types.String `tfsdk:"updated"`
	VMCount     types.Int64  `tfsdk:"vm_count"`
}

// Metadata returns the data source type name.
func (d *DataSourceClusters) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clusters"
}

// Schema defines the schema for the data source.
func (d *DataSourceClusters) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of clusters for the authenticated user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"site": schema.StringAttribute{
				MarkdownDescription: "Site location (svl or rtp). Defaults to 'svl' or inherits from provider configuration.",
				Optional:            true,
				Computed:            true,
			},
			"cluster_count": schema.Int64Attribute{
				MarkdownDescription: "Total number of clusters",
				Computed:            true,
			},
			"clusters": schema.ListNestedAttribute{
				MarkdownDescription: "List of clusters",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "Cluster ID",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Cluster name",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Cluster description",
							Computed:            true,
						},
						"created": schema.StringAttribute{
							MarkdownDescription: "Cluster creation timestamp",
							Computed:            true,
						},
						"updated": schema.StringAttribute{
							MarkdownDescription: "Cluster last update timestamp",
							Computed:            true,
						},
						"vm_count": schema.Int64Attribute{
							MarkdownDescription: "Number of VMs in the cluster",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *DataSourceClusters) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*FyreProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *FyreProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = providerData.Client
	d.defaultSite = providerData.DefaultSite
}

// Read refreshes the Terraform state with the latest data.
func (d *DataSourceClusters) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClustersModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine site to use
	site := client.ListClustersParamsSite(data.Site.ValueString())
	if site == "" {
		site = client.ListClustersParamsSite(d.defaultSite)
	}

	tflog.Debug(ctx, "Fetching clusters", map[string]any{
		"site": string(site),
	})

	// Call API
	clustersResp, err := d.client.ListClustersWithResponse(ctx, &client.ListClustersParams{
		Site: &site,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read clusters: %s", err),
		)
		return
	}

	if clustersResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("API returned status %d: %s", clustersResp.StatusCode(), string(clustersResp.Body)),
		)
		return
	}

	if clustersResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse clusters response. Body: %s", string(clustersResp.Body)),
		)
		return
	}

	// Map response to state
	data.ID = types.StringValue("clusters")
	data.Site = types.StringValue(string(site))

	// Set cluster count - default to 0 if not provided
	if clustersResp.JSON200.ClusterCount != nil {
		data.ClusterCount = types.Int64Value(int64(*clustersResp.JSON200.ClusterCount))
	} else if clustersResp.JSON200.Clusters != nil {
		// If ClusterCount is not provided but we have clusters, use the length
		data.ClusterCount = types.Int64Value(int64(len(*clustersResp.JSON200.Clusters)))
	} else {
		// No clusters and no count, default to 0
		data.ClusterCount = types.Int64Value(0)
	}

	// Map clusters array
	if clustersResp.JSON200.Clusters != nil && len(*clustersResp.JSON200.Clusters) > 0 {
		clustersList := make([]attr.Value, 0, len(*clustersResp.JSON200.Clusters))

		for _, cluster := range *clustersResp.JSON200.Clusters {
			clusterModel := ClusterModel{
				ID:          types.Int64Null(),
				Name:        types.StringNull(),
				Description: types.StringNull(),
				Created:     types.StringNull(),
				Updated:     types.StringNull(),
				VMCount:     types.Int64Null(),
			}

			if cluster.Id != nil {
				clusterModel.ID = types.Int64Value(int64(*cluster.Id))
			}

			if cluster.Name != nil {
				clusterModel.Name = types.StringValue(*cluster.Name)
			}

			if cluster.Description != nil {
				clusterModel.Description = types.StringValue(*cluster.Description)
			}

			if cluster.Created != nil {
				clusterModel.Created = types.StringValue(*cluster.Created)
			}

			if cluster.Updated != nil {
				clusterModel.Updated = types.StringValue(*cluster.Updated)
			}

			if cluster.VmCount != nil {
				clusterModel.VMCount = types.Int64Value(int64(*cluster.VmCount))
			}

			clusterObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
				"id":          types.Int64Type,
				"name":        types.StringType,
				"description": types.StringType,
				"created":     types.StringType,
				"updated":     types.StringType,
				"vm_count":    types.Int64Type,
			}, clusterModel)

			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			clustersList = append(clustersList, clusterObj)
		}

		listValue, diags := types.ListValue(
			types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":          types.Int64Type,
					"name":        types.StringType,
					"description": types.StringType,
					"created":     types.StringType,
					"updated":     types.StringType,
					"vm_count":    types.Int64Type,
				},
			},
			clustersList,
		)

		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.Clusters = listValue
	} else {
		// Empty list
		data.Clusters = types.ListValueMust(
			types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":          types.Int64Type,
					"name":        types.StringType,
					"description": types.StringType,
					"created":     types.StringType,
					"updated":     types.StringType,
					"vm_count":    types.Int64Type,
				},
			},
			[]attr.Value{},
		)
	}

	tflog.Trace(ctx, "read clusters data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
