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

var _ datasource.DataSource = &DataSourceClusterDetails{}

// NewDataSourceClusterDetails creates a new cluster details data source.
func NewDataSourceClusterDetails() datasource.DataSource {
	return &DataSourceClusterDetails{}
}

// DataSourceClusterDetails defines the data source implementation.
type DataSourceClusterDetails struct {
	client      *client.ClientWithResponses
	defaultSite string
}

// ClusterDetailsModel describes the data source data model.
type ClusterDetailsModel struct {
	ID          types.String `tfsdk:"id"`
	ClusterID   types.String `tfsdk:"cluster_id"`
	Site        types.String `tfsdk:"site"`
	IncludeVMs  types.Bool   `tfsdk:"include_vms"`
	UserID      types.Int64  `tfsdk:"user_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Created     types.String `tfsdk:"created"`
	Updated     types.String `tfsdk:"updated"`
	VMs         types.List   `tfsdk:"vms"`
}

// ClusterVMModel describes a VM in the cluster.
type ClusterVMModel struct {
	Location            types.String `tfsdk:"location"`
	ProductGroupID      types.Int64  `tfsdk:"product_group_id"`
	VMID                types.String `tfsdk:"vm_id"`
	Hostname            types.String `tfsdk:"hostname"`
	FQDN                types.String `tfsdk:"fqdn"`
	Platform            types.String `tfsdk:"platform"`
	State               types.String `tfsdk:"state"`
	CPU                 types.Int64  `tfsdk:"cpu"`
	Memory              types.Int64  `tfsdk:"memory"`
	Compliance          types.String `tfsdk:"compliance"`
	AutoPatch           types.String `tfsdk:"auto_patch"`
	Created             types.String `tfsdk:"created"`
	CreatedISO8601      types.String `tfsdk:"created_iso8601"`
	AllowFloatingIP     types.String `tfsdk:"allow_floating_ip"`
	OS                  types.String `tfsdk:"os"`
	OSDisk              types.Int64  `tfsdk:"os_disk"`
	IPs                 types.List   `tfsdk:"ips"`
	HostDown            types.String `tfsdk:"host_down"`
	CurrentOwner        types.Object `tfsdk:"current_owner"`
	AddedToCluster      types.String `tfsdk:"added_to_cluster"`
}

// ClusterVMIPModel describes an IP address.
type ClusterVMIPModel struct {
	IPAddress types.String `tfsdk:"ip_address"`
}

// ClusterVMOwnerModel describes the VM owner.
type ClusterVMOwnerModel struct {
	UserID      types.Int64  `tfsdk:"user_id"`
	Email       types.String `tfsdk:"email"`
	Username    types.String `tfsdk:"username"`
	DisplayName types.String `tfsdk:"displayname"`
}

// Metadata returns the data source type name.
func (d *DataSourceClusterDetails) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_details"
}

// Schema defines the schema for the data source.
func (d *DataSourceClusterDetails) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches details about a Fyre cluster, optionally including VM information.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Terraform resource identifier",
				Computed:            true,
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "Cluster identifier (can be cluster ID or name)",
				Required:            true,
			},
			"site": schema.StringAttribute{
				MarkdownDescription: "Site location (svl or rtp). Defaults to 'svl' or inherits from provider configuration.",
				Optional:            true,
				Computed:            true,
			},
			"include_vms": schema.BoolAttribute{
				MarkdownDescription: "Whether to include VM details in the response. Defaults to false.",
				Optional:            true,
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "User ID of the cluster owner",
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
			"vms": schema.ListNestedAttribute{
				MarkdownDescription: "List of VMs in the cluster (only populated when include_vms is true)",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"location": schema.StringAttribute{
							MarkdownDescription: "VM location",
							Computed:            true,
						},
						"product_group_id": schema.Int64Attribute{
							MarkdownDescription: "Product group ID",
							Computed:            true,
						},
						"vm_id": schema.StringAttribute{
							MarkdownDescription: "VM identifier",
							Computed:            true,
						},
						"hostname": schema.StringAttribute{
							MarkdownDescription: "VM hostname",
							Computed:            true,
						},
						"fqdn": schema.StringAttribute{
							MarkdownDescription: "Fully qualified domain name",
							Computed:            true,
						},
						"platform": schema.StringAttribute{
							MarkdownDescription: "Platform type (x, pvm, z)",
							Computed:            true,
						},
						"state": schema.StringAttribute{
							MarkdownDescription: "VM state",
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
						"compliance": schema.StringAttribute{
							MarkdownDescription: "Compliance status",
							Computed:            true,
						},
						"auto_patch": schema.StringAttribute{
							MarkdownDescription: "Auto-patch setting",
							Computed:            true,
						},
						"created": schema.StringAttribute{
							MarkdownDescription: "VM creation timestamp",
							Computed:            true,
						},
						"created_iso8601": schema.StringAttribute{
							MarkdownDescription: "VM creation timestamp in ISO 8601 format",
							Computed:            true,
						},
						"allow_floating_ip": schema.StringAttribute{
							MarkdownDescription: "Whether floating IPs are allowed",
							Computed:            true,
						},
						"os": schema.StringAttribute{
							MarkdownDescription: "Operating system",
							Computed:            true,
						},
						"os_disk": schema.Int64Attribute{
							MarkdownDescription: "OS disk size in GB",
							Computed:            true,
						},
						"ips": schema.ListNestedAttribute{
							MarkdownDescription: "List of IP addresses",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"ip_address": schema.StringAttribute{
										MarkdownDescription: "IP address",
										Computed:            true,
									},
								},
							},
						},
						"host_down": schema.StringAttribute{
							MarkdownDescription: "Whether the host is down",
							Computed:            true,
						},
						"current_owner": schema.SingleNestedAttribute{
							MarkdownDescription: "Current VM owner information",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"user_id": schema.Int64Attribute{
									MarkdownDescription: "Owner user ID",
									Computed:            true,
								},
								"email": schema.StringAttribute{
									MarkdownDescription: "Owner email",
									Computed:            true,
								},
								"username": schema.StringAttribute{
									MarkdownDescription: "Owner username",
									Computed:            true,
								},
								"displayname": schema.StringAttribute{
									MarkdownDescription: "Owner display name",
									Computed:            true,
								},
							},
						},
						"added_to_cluster": schema.StringAttribute{
							MarkdownDescription: "Timestamp when VM was added to cluster",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *DataSourceClusterDetails) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *DataSourceClusterDetails) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClusterDetailsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine site
	site := data.Site.ValueString()
	if site == "" {
		site = d.defaultSite
		if site == "" {
			site = "svl"
		}
	}

	// Determine if we should include VMs (default to false if not set)
	includeVMs := false
	if !data.IncludeVMs.IsNull() {
		includeVMs = data.IncludeVMs.ValueBool()
	}

	clusterID := data.ClusterID.ValueString()

	tflog.Debug(ctx, "Reading cluster details", map[string]any{
		"cluster_id":  clusterID,
		"site":        site,
		"include_vms": includeVMs,
	})

	var clusterResp *client.GetClusterDetailsResponse
	var clusterWithVMsResp *client.GetClusterDetailsWithVMsResponse
	var err error

	if includeVMs {
		// Call API with VMs included
		params := &client.GetClusterDetailsWithVMsParams{
			Site: (*client.GetClusterDetailsWithVMsParamsSite)(&site),
		}
		clusterWithVMsResp, err = d.client.GetClusterDetailsWithVMsWithResponse(ctx, clusterID, params)
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to read cluster details with VMs: %s", err),
			)
			return
		}

		if clusterWithVMsResp.StatusCode() != 200 {
			resp.Diagnostics.AddError(
				"API Error",
				fmt.Sprintf("API returned status %d: %s", clusterWithVMsResp.StatusCode(), string(clusterWithVMsResp.Body)),
			)
			return
		}

		if clusterWithVMsResp.JSON200 == nil {
			resp.Diagnostics.AddError(
				"Parse Error",
				fmt.Sprintf("Unable to parse cluster response. Body: %s", string(clusterWithVMsResp.Body)),
			)
			return
		}
	} else {
		// Call API without VMs
		params := &client.GetClusterDetailsParams{
			Site: (*client.GetClusterDetailsParamsSite)(&site),
		}
		clusterResp, err = d.client.GetClusterDetailsWithResponse(ctx, clusterID, params)
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to read cluster details: %s", err),
			)
			return
		}

		if clusterResp.StatusCode() != 200 {
			resp.Diagnostics.AddError(
				"API Error",
				fmt.Sprintf("API returned status %d: %s", clusterResp.StatusCode(), string(clusterResp.Body)),
			)
			return
		}

		if clusterResp.JSON200 == nil {
			resp.Diagnostics.AddError(
				"Parse Error",
				fmt.Sprintf("Unable to parse cluster response. Body: %s", string(clusterResp.Body)),
			)
			return
		}
	}

	// Map response to model
	var cluster *struct {
		Created     *string              `json:"created,omitempty"`
		Description *string              `json:"description,omitempty"`
		Id          *int                 `json:"id,omitempty"`
		Name        *string              `json:"name,omitempty"`
		Updated     *string              `json:"updated,omitempty"`
		UserId      *int                 `json:"user_id,omitempty"`
		Vms         *[]client.VMSummary  `json:"vms,omitempty"`
	}

	if includeVMs {
		cluster = clusterWithVMsResp.JSON200.Cluster
	} else {
		cluster = clusterResp.JSON200.Cluster
	}

	if cluster == nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			"Cluster data is missing from response",
		)
		return
	}

	// Set computed values
	data.ID = types.StringValue(fmt.Sprintf("cluster-%s", clusterID))
	data.Site = types.StringValue(site)
	data.IncludeVMs = types.BoolValue(includeVMs)

	if cluster.Id != nil {
		data.ClusterID = types.StringValue(fmt.Sprintf("%d", *cluster.Id))
	}

	if cluster.UserId != nil {
		data.UserID = types.Int64Value(int64(*cluster.UserId))
	} else {
		data.UserID = types.Int64Null()
	}

	if cluster.Name != nil {
		data.Name = types.StringValue(*cluster.Name)
	} else {
		data.Name = types.StringNull()
	}

	if cluster.Description != nil {
		data.Description = types.StringValue(*cluster.Description)
	} else {
		data.Description = types.StringNull()
	}

	if cluster.Created != nil {
		data.Created = types.StringValue(*cluster.Created)
	} else {
		data.Created = types.StringNull()
	}

	if cluster.Updated != nil {
		data.Updated = types.StringValue(*cluster.Updated)
	} else {
		data.Updated = types.StringNull()
	}

	// Handle VMs list
	if cluster.Vms != nil && len(*cluster.Vms) > 0 {
		vmsList := make([]ClusterVMModel, 0, len(*cluster.Vms))
		for _, vm := range *cluster.Vms {
			vmModel := ClusterVMModel{
				Location:        types.StringNull(),
				ProductGroupID:  types.Int64Null(),
				VMID:            types.StringNull(),
				Hostname:        types.StringNull(),
				FQDN:            types.StringNull(),
				Platform:        types.StringNull(),
				State:           types.StringNull(),
				CPU:             types.Int64Null(),
				Memory:          types.Int64Null(),
				Compliance:      types.StringNull(),
				AutoPatch:       types.StringNull(),
				Created:         types.StringNull(),
				CreatedISO8601:  types.StringNull(),
				AllowFloatingIP: types.StringNull(),
				OS:              types.StringNull(),
				OSDisk:          types.Int64Null(),
				IPs:             types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{"ip_address": types.StringType}}),
				HostDown:        types.StringNull(),
				CurrentOwner:    types.ObjectNull(map[string]attr.Type{"user_id": types.Int64Type, "email": types.StringType, "username": types.StringType, "displayname": types.StringType}),
				AddedToCluster:  types.StringNull(),
			}

			if vm.Location != nil {
				vmModel.Location = types.StringValue(*vm.Location)
			}
			if vm.ProductGroupId != nil {
				vmModel.ProductGroupID = types.Int64Value(int64(*vm.ProductGroupId))
			}
			if vm.VmId != nil {
				vmModel.VMID = types.StringValue(*vm.VmId)
			}
			if vm.Hostname != nil {
				vmModel.Hostname = types.StringValue(*vm.Hostname)
			}
			if vm.Fqdn != nil {
				vmModel.FQDN = types.StringValue(*vm.Fqdn)
			}
			if vm.Platform != nil {
				vmModel.Platform = types.StringValue(string(*vm.Platform))
			}
			if vm.State != nil {
				vmModel.State = types.StringValue(*vm.State)
			}
			if vm.Cpu != nil {
				vmModel.CPU = types.Int64Value(int64(*vm.Cpu))
			}
			if vm.Memory != nil {
				vmModel.Memory = types.Int64Value(int64(*vm.Memory))
			}
			if vm.Compliance != nil {
				vmModel.Compliance = types.StringValue(*vm.Compliance)
			}
			if vm.AutoPatch != nil {
				vmModel.AutoPatch = types.StringValue(*vm.AutoPatch)
			}
			if vm.Created != nil {
				vmModel.Created = types.StringValue(*vm.Created)
			}
			if vm.CreatedIso8601 != nil {
				vmModel.CreatedISO8601 = types.StringValue(*vm.CreatedIso8601)
			}
			if vm.AllowFloatingIp != nil {
				vmModel.AllowFloatingIP = types.StringValue(*vm.AllowFloatingIp)
			}
			if vm.Os != nil {
				vmModel.OS = types.StringValue(*vm.Os)
			}
			if vm.OsDisk != nil {
				vmModel.OSDisk = types.Int64Value(int64(*vm.OsDisk))
			}
			if vm.HostDown != nil {
				vmModel.HostDown = types.StringValue(*vm.HostDown)
			}
			if vm.AddedToCluster != nil {
				vmModel.AddedToCluster = types.StringValue(*vm.AddedToCluster)
			}

			// Handle IPs
			if vm.Ips != nil && len(*vm.Ips) > 0 {
				ipsList := make([]ClusterVMIPModel, 0, len(*vm.Ips))
				for _, ip := range *vm.Ips {
					ipModel := ClusterVMIPModel{
						IPAddress: types.StringNull(),
					}
					if ip.IpAddress != nil {
						ipModel.IPAddress = types.StringValue(*ip.IpAddress)
					}
					ipsList = append(ipsList, ipModel)
				}
				ipsListValue, diags := types.ListValueFrom(ctx, types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"ip_address": types.StringType,
					},
				}, ipsList)
				resp.Diagnostics.Append(diags...)
				if !resp.Diagnostics.HasError() {
					vmModel.IPs = ipsListValue
				}
			}

			// Handle current owner
			if vm.CurrentOwner != nil {
				ownerModel := ClusterVMOwnerModel{
					UserID:      types.Int64Null(),
					Email:       types.StringNull(),
					Username:    types.StringNull(),
					DisplayName: types.StringNull(),
				}
				if vm.CurrentOwner.UserId != nil {
					ownerModel.UserID = types.Int64Value(int64(*vm.CurrentOwner.UserId))
				}
				if vm.CurrentOwner.Email != nil {
					ownerModel.Email = types.StringValue(*vm.CurrentOwner.Email)
				}
				if vm.CurrentOwner.Username != nil {
					ownerModel.Username = types.StringValue(*vm.CurrentOwner.Username)
				}
				if vm.CurrentOwner.Displayname != nil {
					ownerModel.DisplayName = types.StringValue(*vm.CurrentOwner.Displayname)
				}
				ownerObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
					"user_id":     types.Int64Type,
					"email":       types.StringType,
					"username":    types.StringType,
					"displayname": types.StringType,
				}, ownerModel)
				resp.Diagnostics.Append(diags...)
				if !resp.Diagnostics.HasError() {
					vmModel.CurrentOwner = ownerObj
				}
			}

			vmsList = append(vmsList, vmModel)
		}

		vmsListValue, diags := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"location":           types.StringType,
				"product_group_id":   types.Int64Type,
				"vm_id":              types.StringType,
				"hostname":           types.StringType,
				"fqdn":               types.StringType,
				"platform":           types.StringType,
				"state":              types.StringType,
				"cpu":                types.Int64Type,
				"memory":             types.Int64Type,
				"compliance":         types.StringType,
				"auto_patch":         types.StringType,
				"created":            types.StringType,
				"created_iso8601":    types.StringType,
				"allow_floating_ip":  types.StringType,
				"os":                 types.StringType,
				"os_disk":            types.Int64Type,
				"ips":                types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{"ip_address": types.StringType}}},
				"host_down":          types.StringType,
				"current_owner":      types.ObjectType{AttrTypes: map[string]attr.Type{"user_id": types.Int64Type, "email": types.StringType, "username": types.StringType, "displayname": types.StringType}},
				"added_to_cluster":   types.StringType,
			},
		}, vmsList)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.VMs = vmsListValue
		}
	} else {
		data.VMs = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"location":           types.StringType,
				"product_group_id":   types.Int64Type,
				"vm_id":              types.StringType,
				"hostname":           types.StringType,
				"fqdn":               types.StringType,
				"platform":           types.StringType,
				"state":              types.StringType,
				"cpu":                types.Int64Type,
				"memory":             types.Int64Type,
				"compliance":         types.StringType,
				"auto_patch":         types.StringType,
				"created":            types.StringType,
				"created_iso8601":    types.StringType,
				"allow_floating_ip":  types.StringType,
				"os":                 types.StringType,
				"os_disk":            types.Int64Type,
				"ips":                types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{"ip_address": types.StringType}}},
				"host_down":          types.StringType,
				"current_owner":      types.ObjectType{AttrTypes: map[string]attr.Type{"user_id": types.Int64Type, "email": types.StringType, "username": types.StringType, "displayname": types.StringType}},
				"added_to_cluster":   types.StringType,
			},
		})
	}

	tflog.Trace(ctx, "read cluster details data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
