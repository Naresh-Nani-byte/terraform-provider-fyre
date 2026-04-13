// Copyright IBM Corp. 2021, 2026
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

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &DataSourceQuota{}

// NewDataSourceQuota creates a new instance of the quota data source.
// This factory function is called by the provider to instantiate the data source.
func NewDataSourceQuota() datasource.DataSource {
	return &DataSourceQuota{}
}

// DataSourceQuota defines the data source implementation.
type DataSourceQuota struct {
	client      *client.ClientWithResponses
	defaultSite string
}

// QuotaModel describes the data source data model.
type QuotaModel struct {
	ID      types.String `tfsdk:"id"`
	Site    types.String `tfsdk:"site"`
	Status  types.String `tfsdk:"status"`
	Details types.Object `tfsdk:"details"`
}

// QuotaDetailsModel describes the details nested object.
type QuotaDetailsModel struct {
	Ip               types.Object `tfsdk:"ip"`
	Platforms        types.List   `tfsdk:"platforms"`
	ProductGroupId   types.Int64  `tfsdk:"product_group_id"`
	ProductGroupName types.String `tfsdk:"product_group_name"`
	X                types.Object `tfsdk:"x"`
}

// IpQuotaModel describes the IP quota nested object.
type IpQuotaModel struct {
	Public types.Object `tfsdk:"public"`
}

// PublicIpQuotaModel describes the public IP quota nested object.
type PublicIpQuotaModel struct {
	Quota types.Int64 `tfsdk:"quota"`
	Used  types.Int64 `tfsdk:"used"`
}

// PlatformQuotaModel describes the platform (X) quota nested object.
type PlatformQuotaModel struct {
	Cpu           types.Int64 `tfsdk:"cpu"`
	CpuPercent    types.Int64 `tfsdk:"cpu_percent"`
	CpuUsed       types.Int64 `tfsdk:"cpu_used"`
	Disk          types.Int64 `tfsdk:"disk"`
	DiskPercent   types.Int64 `tfsdk:"disk_percent"`
	DiskUsed      types.Int64 `tfsdk:"disk_used"`
	Memory        types.Int64 `tfsdk:"memory"`
	MemoryPercent types.Int64 `tfsdk:"memory_percent"`
	MemoryUsed    types.Int64 `tfsdk:"memory_used"`
}

// Metadata sets the data source type name for the quota data source.
// The type name is used in Terraform configurations as "fyre_quota".
func (d *DataSourceQuota) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quota"
}

// Schema defines the structure and attributes of the quota data source.
// It specifies optional site parameter and computed attributes including IP quotas,
// available platforms, product group details, and platform-specific resource quotas
// (CPU, memory, disk usage and limits).
func (d *DataSourceQuota) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches quota information for the authenticated user, including IP quotas, available platforms, product group details, and platform-specific resource quotas (CPU, memory, disk).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The quota identifier",
				Computed:            true,
			},
			"site": schema.StringAttribute{
				MarkdownDescription: "Site location (svl or rtp). Defaults to 'svl' or inherits from provider configuration.",
				Optional:            true,
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "API response status",
				Computed:            true,
			},
			"details": schema.SingleNestedAttribute{
				MarkdownDescription: "Quota details including IP, platforms, and resource quotas",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"ip": schema.SingleNestedAttribute{
						MarkdownDescription: "IP quota information",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"public": schema.SingleNestedAttribute{
								MarkdownDescription: "Public IP quota details",
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"quota": schema.Int64Attribute{
										MarkdownDescription: "Total public IP quota",
										Computed:            true,
									},
									"used": schema.Int64Attribute{
										MarkdownDescription: "Number of public IPs currently in use",
										Computed:            true,
									},
								},
							},
						},
					},
					"platforms": schema.ListAttribute{
						MarkdownDescription: "List of available platforms",
						Computed:            true,
						ElementType:         types.StringType,
					},
					"product_group_id": schema.Int64Attribute{
						MarkdownDescription: "Product group identifier",
						Computed:            true,
					},
					"product_group_name": schema.StringAttribute{
						MarkdownDescription: "Product group name",
						Computed:            true,
					},
					"x": schema.SingleNestedAttribute{
						MarkdownDescription: "X platform resource quotas",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"cpu": schema.Int64Attribute{
								MarkdownDescription: "Total CPU quota",
								Computed:            true,
							},
							"cpu_percent": schema.Int64Attribute{
								MarkdownDescription: "CPU usage percentage",
								Computed:            true,
							},
							"cpu_used": schema.Int64Attribute{
								MarkdownDescription: "Number of CPUs currently in use",
								Computed:            true,
							},
							"disk": schema.Int64Attribute{
								MarkdownDescription: "Total disk quota in GB",
								Computed:            true,
							},
							"disk_percent": schema.Int64Attribute{
								MarkdownDescription: "Disk usage percentage",
								Computed:            true,
							},
							"disk_used": schema.Int64Attribute{
								MarkdownDescription: "Disk space currently in use in GB",
								Computed:            true,
							},
							"memory": schema.Int64Attribute{
								MarkdownDescription: "Total memory quota in GB",
								Computed:            true,
							},
							"memory_percent": schema.Int64Attribute{
								MarkdownDescription: "Memory usage percentage",
								Computed:            true,
							},
							"memory_used": schema.Int64Attribute{
								MarkdownDescription: "Memory currently in use in GB",
								Computed:            true,
							},
						},
					},
				},
			},
		},
	}
}

// Configure initializes the quota data source with the Fyre API client
// and default site configuration from the provider. This method is called by the
// framework during provider initialization.
func (d *DataSourceQuota) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

// Read retrieves quota information for the authenticated user from the Fyre API.
// Returns IP allocation quotas, available platforms, product group details, and
// platform-specific resource quotas including CPU, memory, and disk usage with
// percentage utilization.
func (d *DataSourceQuota) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data QuotaModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare API request parameters
	site := client.GetQuotaParamsSite(data.Site.ValueString())
	if site == "" {
		site = client.GetQuotaParamsSite(d.defaultSite)
	}

	// Call API
	quotaResp, err := d.client.GetQuotaWithResponse(ctx, &client.GetQuotaParams{
		Site: &site,
	})
	if quotaResp != nil {
		// Log response for debugging
		tflog.Debug(ctx, "GetQuota API response", map[string]any{
			"status_code": quotaResp.StatusCode(),
			"has_json200": quotaResp.JSON200 != nil,
		})
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read quota: %s", err),
		)
		return
	}

	// Check HTTP status
	if quotaResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("API returned status %d: %s", quotaResp.StatusCode(), string(quotaResp.Body)),
		)
		return
	}

	// Parse response
	if quotaResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse quota response. Body: %s", string(quotaResp.Body)),
		)
		return
	}

	// Map to Terraform state
	data.ID = types.StringValue("quota")
	data.Site = types.StringValue(string(site))

	// Initialize status
	if quotaResp.JSON200.Status != nil {
		data.Status = types.StringValue(*quotaResp.JSON200.Status)
	} else {
		data.Status = types.StringNull()
	}

	// Initialize details with null values
	detailsModel := QuotaDetailsModel{
		Ip: types.ObjectNull(map[string]attr.Type{
			"public": types.ObjectType{AttrTypes: map[string]attr.Type{
				"quota": types.Int64Type,
				"used":  types.Int64Type,
			}},
		}),
		Platforms:        types.ListNull(types.StringType),
		ProductGroupId:   types.Int64Null(),
		ProductGroupName: types.StringNull(),
		X: types.ObjectNull(map[string]attr.Type{
			"cpu":            types.Int64Type,
			"cpu_percent":    types.Int64Type,
			"cpu_used":       types.Int64Type,
			"disk":           types.Int64Type,
			"disk_percent":   types.Int64Type,
			"disk_used":      types.Int64Type,
			"memory":         types.Int64Type,
			"memory_percent": types.Int64Type,
			"memory_used":    types.Int64Type,
		}),
	}

	// Extract details if available - API returns an array, we take the first element
	if qRes := quotaResp.JSON200; qRes != nil && qRes.Details != nil && len(*qRes.Details) > 0 {
		details := (*qRes.Details)[0]

		// Map IP quotas
		if details.Ip != nil && details.Ip.Public != nil {
			publicModel := PublicIpQuotaModel{
				Quota: types.Int64Null(),
				Used:  types.Int64Null(),
			}

			if details.Ip.Public.Quota != nil {
				publicModel.Quota = types.Int64Value(int64(*details.Ip.Public.Quota))
			}
			if details.Ip.Public.Used != nil {
				publicModel.Used = types.Int64Value(int64(*details.Ip.Public.Used))
			}

			publicObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
				"quota": types.Int64Type,
				"used":  types.Int64Type,
			}, publicModel)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				ipModel := IpQuotaModel{Public: publicObj}
				ipObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
					"public": types.ObjectType{AttrTypes: map[string]attr.Type{
						"quota": types.Int64Type,
						"used":  types.Int64Type,
					}},
				}, ipModel)
				resp.Diagnostics.Append(diags...)
				if !resp.Diagnostics.HasError() {
					detailsModel.Ip = ipObj
				}
			}
		}

		// Map platforms list
		if details.Platforms != nil && len(*details.Platforms) > 0 {
			platformsList := make([]attr.Value, 0, len(*details.Platforms))
			for _, platform := range *details.Platforms {
				platformsList = append(platformsList, types.StringValue(platform))
			}
			listValue, diags := types.ListValue(types.StringType, platformsList)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				detailsModel.Platforms = listValue
			}
		}

		// Map product group info
		if details.ProductGroupId != nil {
			detailsModel.ProductGroupId = types.Int64Value(int64(*details.ProductGroupId))
		}
		if details.ProductGroupName != nil {
			detailsModel.ProductGroupName = types.StringValue(*details.ProductGroupName)
		}

		// Map X platform quotas
		if details.X != nil {
			xModel := PlatformQuotaModel{
				Cpu:           types.Int64Null(),
				CpuPercent:    types.Int64Null(),
				CpuUsed:       types.Int64Null(),
				Disk:          types.Int64Null(),
				DiskPercent:   types.Int64Null(),
				DiskUsed:      types.Int64Null(),
				Memory:        types.Int64Null(),
				MemoryPercent: types.Int64Null(),
				MemoryUsed:    types.Int64Null(),
			}

			if details.X.Cpu != nil {
				xModel.Cpu = types.Int64Value(int64(*details.X.Cpu))
			}
			if details.X.CpuPercent != nil {
				xModel.CpuPercent = types.Int64Value(int64(*details.X.CpuPercent))
			}
			if details.X.CpuUsed != nil {
				xModel.CpuUsed = types.Int64Value(int64(*details.X.CpuUsed))
			}
			if details.X.Disk != nil {
				xModel.Disk = types.Int64Value(int64(*details.X.Disk))
			}
			if details.X.DiskPercent != nil {
				xModel.DiskPercent = types.Int64Value(int64(*details.X.DiskPercent))
			}
			if details.X.DiskUsed != nil {
				xModel.DiskUsed = types.Int64Value(int64(*details.X.DiskUsed))
			}
			if details.X.Memory != nil {
				xModel.Memory = types.Int64Value(int64(*details.X.Memory))
			}
			if details.X.MemoryPercent != nil {
				xModel.MemoryPercent = types.Int64Value(int64(*details.X.MemoryPercent))
			}
			if details.X.MemoryUsed != nil {
				xModel.MemoryUsed = types.Int64Value(int64(*details.X.MemoryUsed))
			}

			xObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
				"cpu":            types.Int64Type,
				"cpu_percent":    types.Int64Type,
				"cpu_used":       types.Int64Type,
				"disk":           types.Int64Type,
				"disk_percent":   types.Int64Type,
				"disk_used":      types.Int64Type,
				"memory":         types.Int64Type,
				"memory_percent": types.Int64Type,
				"memory_used":    types.Int64Type,
			}, xModel)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				detailsModel.X = xObj
			}
		}
	}

	// Convert details to object
	detailsObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"ip": types.ObjectType{AttrTypes: map[string]attr.Type{
			"public": types.ObjectType{AttrTypes: map[string]attr.Type{
				"quota": types.Int64Type,
				"used":  types.Int64Type,
			}},
		}},
		"platforms":          types.ListType{ElemType: types.StringType},
		"product_group_id":   types.Int64Type,
		"product_group_name": types.StringType,
		"x": types.ObjectType{AttrTypes: map[string]attr.Type{
			"cpu":            types.Int64Type,
			"cpu_percent":    types.Int64Type,
			"cpu_used":       types.Int64Type,
			"disk":           types.Int64Type,
			"disk_percent":   types.Int64Type,
			"disk_used":      types.Int64Type,
			"memory":         types.Int64Type,
			"memory_percent": types.Int64Type,
			"memory_used":    types.Int64Type,
		}},
	}, detailsModel)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		data.Details = detailsObj
	}

	// Write logs using the tflog package
	tflog.Trace(ctx, "read quota data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
