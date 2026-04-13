// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp-forge/terraform-provider-fyre/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &DataSourceStencils{}

// NewDataSourceStencils creates a new stencils data source.
func NewDataSourceStencils() datasource.DataSource {
	return &DataSourceStencils{}
}

// DataSourceStencils defines the stencils data source implementation.
type DataSourceStencils struct {
	client      *client.ClientWithResponses
	defaultSite string
}

// StencilsModel describes the stencils data source data model.
type StencilsModel struct {
	ID             types.String `tfsdk:"id"`
	Site           types.String `tfsdk:"site"`
	ProductGroupID types.Int64  `tfsdk:"product_group_id"`
	Stencils       types.List   `tfsdk:"stencils"`
}

// StencilModel describes a single stencil.
type StencilModel struct {
	StencilIDs     types.List   `tfsdk:"stencil_ids"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Platform       types.String `tfsdk:"platform"`
	OS             types.String `tfsdk:"os"`
	CPU            types.Int64  `tfsdk:"cpu"`
	Memory         types.Int64  `tfsdk:"memory"`
	Disk           types.Int64  `tfsdk:"disk"`
	Owner          types.Object `tfsdk:"owner"`
	ProductGroupID types.Int64  `tfsdk:"product_group_id"`
}

// StencilOwnerModel describes the owner nested object for stencils.
type StencilOwnerModel struct {
	UserID   types.Int64  `tfsdk:"user_id"`
	Name     types.String `tfsdk:"name"`
	Username types.String `tfsdk:"username"`
	Email    types.String `tfsdk:"email"`
}

// Metadata returns the data source type name.
func (d *DataSourceStencils) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stencils"
}

// Schema defines the schema for the data source.
func (d *DataSourceStencils) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of Fyre stencils filtered by product group.",
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
			"product_group_id": schema.Int64Attribute{
				MarkdownDescription: "Product group ID to filter stencils",
				Required:            true,
			},
			"stencils": schema.ListNestedAttribute{
				MarkdownDescription: "List of stencils in the product group",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"stencil_ids": schema.ListAttribute{
							MarkdownDescription: "List of stencil identifiers",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Stencil name",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Stencil description",
							Computed:            true,
						},
						"platform": schema.StringAttribute{
							MarkdownDescription: "Platform type",
							Computed:            true,
						},
						"os": schema.StringAttribute{
							MarkdownDescription: "Operating system",
							Computed:            true,
						},
						"cpu": schema.Int64Attribute{
							MarkdownDescription: "Number of CPUs",
							Computed:            true,
						},
						"memory": schema.Int64Attribute{
							MarkdownDescription: "Memory in GB",
							Computed:            true,
						},
						"disk": schema.Int64Attribute{
							MarkdownDescription: "Disk size in GB",
							Computed:            true,
						},
						"owner": schema.SingleNestedAttribute{
							MarkdownDescription: "Stencil owner information",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"user_id": schema.Int64Attribute{
									MarkdownDescription: "Owner user ID",
									Computed:            true,
								},
								"name": schema.StringAttribute{
									MarkdownDescription: "Owner name",
									Computed:            true,
								},
								"username": schema.StringAttribute{
									MarkdownDescription: "Owner username",
									Computed:            true,
								},
								"email": schema.StringAttribute{
									MarkdownDescription: "Owner email address",
									Computed:            true,
								},
							},
						},
						"product_group_id": schema.Int64Attribute{
							MarkdownDescription: "Product group ID",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *DataSourceStencils) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *DataSourceStencils) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StencilsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine site
	site := data.Site.ValueString()
	if site == "" {
		site = d.defaultSite
	}

	// Call API
	productGroupID := client.ProductGroupId(strconv.FormatInt(data.ProductGroupID.ValueInt64(), 10))
	stencilsResp, err := d.client.ListStencilsByProductGroupWithResponse(ctx, productGroupID, &client.ListStencilsByProductGroupParams{
		Site: (*client.ListStencilsByProductGroupParamsSite)(&site),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read stencils: %s", err),
		)
		return
	}

	if stencilsResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("API returned status %d: %s", stencilsResp.StatusCode(), string(stencilsResp.Body)),
		)
		return
	}

	if stencilsResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse stencils response. Body: %s", string(stencilsResp.Body)),
		)
		return
	}

	tflog.Debug(ctx, "Successfully retrieved stencils", map[string]any{
		"product_group_id": productGroupID,
		"count":            len(*stencilsResp.JSON200),
	})

	// Map to Terraform state
	data.ID = types.StringValue(fmt.Sprintf("stencils-%s", productGroupID))
	data.Site = types.StringValue(site)

	// Convert stencils to list
	stencilsList := make([]attr.Value, 0, len(*stencilsResp.JSON200))
	for _, stencil := range *stencilsResp.JSON200 {
		stencilModel := StencilModel{
			StencilIDs:  types.ListNull(types.StringType),
			Name:        types.StringNull(),
			Description: types.StringNull(),
			Platform:    types.StringNull(),
			OS:          types.StringNull(),
			CPU:         types.Int64Null(),
			Memory:      types.Int64Null(),
			Disk:        types.Int64Null(),
			Owner: types.ObjectNull(map[string]attr.Type{
				"user_id":  types.Int64Type,
				"name":     types.StringType,
				"username": types.StringType,
				"email":    types.StringType,
			}),
			ProductGroupID: types.Int64Null(),
		}

		// Map stencil_ids
		if stencil.StencilIds != nil && len(*stencil.StencilIds) > 0 {
			stencilIDsList := make([]attr.Value, 0, len(*stencil.StencilIds))
			for _, id := range *stencil.StencilIds {
				if id.StencilId != nil {
					stencilIDsList = append(stencilIDsList, types.StringValue(*id.StencilId))
				}
			}
			listValue, diags := types.ListValue(types.StringType, stencilIDsList)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				stencilModel.StencilIDs = listValue
			}
		}

		// Map simple fields
		if stencil.Name != nil {
			stencilModel.Name = types.StringValue(*stencil.Name)
		}
		if stencil.Description != nil {
			stencilModel.Description = types.StringValue(*stencil.Description)
		}
		if stencil.Platform != nil {
			stencilModel.Platform = types.StringValue(*stencil.Platform)
		}
		if stencil.Os != nil {
			stencilModel.OS = types.StringValue(*stencil.Os)
		}
		if stencil.Cpu != nil {
			stencilModel.CPU = types.Int64Value(int64(*stencil.Cpu))
		}
		if stencil.Memory != nil {
			stencilModel.Memory = types.Int64Value(int64(*stencil.Memory))
		}
		if stencil.Disk != nil {
			// Disk is a string in the API response, parse it
			if diskInt, err := strconv.ParseInt(*stencil.Disk, 10, 64); err == nil {
				stencilModel.Disk = types.Int64Value(diskInt)
			}
		}
		if stencil.ProductGroupId != nil {
			stencilModel.ProductGroupID = types.Int64Value(int64(*stencil.ProductGroupId))
		}

		// Map owner nested object
		if stencil.Owner != nil {
			ownerModel := StencilOwnerModel{
				UserID:   types.Int64Null(),
				Name:     types.StringNull(),
				Username: types.StringNull(),
				Email:    types.StringNull(),
			}

			if stencil.Owner.UserId != nil {
				ownerModel.UserID = types.Int64Value(int64(*stencil.Owner.UserId))
			}
			if stencil.Owner.Name != nil {
				ownerModel.Name = types.StringValue(*stencil.Owner.Name)
			}
			if stencil.Owner.Username != nil {
				ownerModel.Username = types.StringValue(*stencil.Owner.Username)
			}
			if stencil.Owner.Email != nil {
				ownerModel.Email = types.StringValue(*stencil.Owner.Email)
			}

			ownerObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
				"user_id":  types.Int64Type,
				"name":     types.StringType,
				"username": types.StringType,
				"email":    types.StringType,
			}, ownerModel)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				stencilModel.Owner = ownerObj
			}
		}

		// Convert stencil model to object
		stencilObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"stencil_ids": types.ListType{ElemType: types.StringType},
			"name":        types.StringType,
			"description": types.StringType,
			"platform":    types.StringType,
			"os":          types.StringType,
			"cpu":         types.Int64Type,
			"memory":      types.Int64Type,
			"disk":        types.Int64Type,
			"owner": types.ObjectType{AttrTypes: map[string]attr.Type{
				"user_id":  types.Int64Type,
				"name":     types.StringType,
				"username": types.StringType,
				"email":    types.StringType,
			}},
			"product_group_id": types.Int64Type,
		}, stencilModel)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		stencilsList = append(stencilsList, stencilObj)
	}

	stencilsListValue, diags := types.ListValue(
		types.ObjectType{AttrTypes: map[string]attr.Type{
			"stencil_ids": types.ListType{ElemType: types.StringType},
			"name":        types.StringType,
			"description": types.StringType,
			"platform":    types.StringType,
			"os":          types.StringType,
			"cpu":         types.Int64Type,
			"memory":      types.Int64Type,
			"disk":        types.Int64Type,
			"owner": types.ObjectType{AttrTypes: map[string]attr.Type{
				"user_id":  types.Int64Type,
				"name":     types.StringType,
				"username": types.StringType,
				"email":    types.StringType,
			}},
			"product_group_id": types.Int64Type,
		}},
		stencilsList,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Stencils = stencilsListValue

	tflog.Trace(ctx, "read stencils data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
