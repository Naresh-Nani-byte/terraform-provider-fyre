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

var _ datasource.DataSource = &DataSourceVMOSAvailable{}

// NewDataSourceVMOSAvailable creates a new instance of the VM OS available data source.
// This factory function is called by the provider to instantiate the data source.
func NewDataSourceVMOSAvailable() datasource.DataSource {
	return &DataSourceVMOSAvailable{}
}

type DataSourceVMOSAvailable struct {
	client      *client.ClientWithResponses
	defaultSite string
}

type VMOSAvailableModel struct {
	ID          types.String `tfsdk:"id"`
	Platform    types.String `tfsdk:"platform"`
	Site        types.String `tfsdk:"site"`
	CentOS      types.List   `tfsdk:"centos"`
	RedHat      types.List   `tfsdk:"redhat"`
	Rocky       types.List   `tfsdk:"rocky"`
	SLES        types.List   `tfsdk:"sles"`
	Ubuntu      types.List   `tfsdk:"ubuntu"`
	Windows     types.List   `tfsdk:"windows"`
	DefaultSize types.Object `tfsdk:"default_size"`
}

type DefaultSizeModel struct {
	Count            types.String `tfsdk:"count"`
	CPU              types.String `tfsdk:"cpu"`
	Memory           types.String `tfsdk:"memory"`
	MaxCount         types.String `tfsdk:"max_count"`
	MaxCPU           types.String `tfsdk:"max_cpu"`
	MaxMemory        types.String `tfsdk:"max_memory"`
	MaxDiskCount     types.String `tfsdk:"max_disk_count"`
	MaxDiskSize      types.String `tfsdk:"max_disk_size"`
	MaxTotalDiskSize types.String `tfsdk:"max_total_disk_size"`
	PVMMaxCPU        types.String `tfsdk:"pvm_max_cpu"`
	PVMMaxMemory     types.String `tfsdk:"pvm_max_memory"`
}

// Metadata sets the data source type name for the VM OS available data source.
// The type name is used in Terraform configurations as "fyre_vm_os_available".
func (d *DataSourceVMOSAvailable) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_os_available"
}

// Schema defines the structure and attributes of the VM OS available data source.
// It specifies the required platform parameter, optional site parameter, and computed
// attributes for each OS distribution family along with default VM sizing constraints.
func (d *DataSourceVMOSAvailable) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of available operating systems for a specific platform and site, organized by distribution family, along with default VM sizing constraints.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier (format: platform-site)",
				Computed:            true,
			},
			"platform": schema.StringAttribute{
				MarkdownDescription: "Platform type (x, pvm, or z)",
				Required:            true,
			},
			"site": schema.StringAttribute{
				MarkdownDescription: "Site location (svl or rtp). Defaults to 'svl' or inherits from provider configuration.",
				Optional:            true,
				Computed:            true,
			},
			"centos": schema.ListAttribute{
				MarkdownDescription: "List of available CentOS versions",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"redhat": schema.ListAttribute{
				MarkdownDescription: "List of available RedHat versions",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"rocky": schema.ListAttribute{
				MarkdownDescription: "List of available Rocky Linux versions",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"sles": schema.ListAttribute{
				MarkdownDescription: "List of available SUSE Linux Enterprise Server versions",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"ubuntu": schema.ListAttribute{
				MarkdownDescription: "List of available Ubuntu versions",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"windows": schema.ListAttribute{
				MarkdownDescription: "List of available Windows versions",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"default_size": schema.SingleNestedAttribute{
				MarkdownDescription: "Default and maximum VM sizing constraints for this platform and site",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"count": schema.StringAttribute{
						MarkdownDescription: "Default VM count",
						Computed:            true,
					},
					"cpu": schema.StringAttribute{
						MarkdownDescription: "Default CPU count",
						Computed:            true,
					},
					"memory": schema.StringAttribute{
						MarkdownDescription: "Default memory in GB",
						Computed:            true,
					},
					"max_count": schema.StringAttribute{
						MarkdownDescription: "Maximum VM count",
						Computed:            true,
					},
					"max_cpu": schema.StringAttribute{
						MarkdownDescription: "Maximum CPU count",
						Computed:            true,
					},
					"max_memory": schema.StringAttribute{
						MarkdownDescription: "Maximum memory in GB",
						Computed:            true,
					},
					"max_disk_count": schema.StringAttribute{
						MarkdownDescription: "Maximum number of disks",
						Computed:            true,
					},
					"max_disk_size": schema.StringAttribute{
						MarkdownDescription: "Maximum disk size in GB",
						Computed:            true,
					},
					"max_total_disk_size": schema.StringAttribute{
						MarkdownDescription: "Maximum total disk size in GB",
						Computed:            true,
					},
					"pvm_max_cpu": schema.StringAttribute{
						MarkdownDescription: "PVM-specific maximum CPU count (only present for some platforms/sites)",
						Computed:            true,
					},
					"pvm_max_memory": schema.StringAttribute{
						MarkdownDescription: "PVM-specific maximum memory in GB (only present for some platforms/sites)",
						Computed:            true,
					},
				},
			},
		},
	}
}

// Configure initializes the VM OS available data source with the Fyre API client
// and default site configuration from the provider. This method is called by the
// framework during provider initialization.
func (d *DataSourceVMOSAvailable) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read retrieves the list of available operating systems and default VM sizing
// constraints from the Fyre API for a specified platform (x, pvm, or z) and site.
// It returns separate lists for each OS distribution family along with default and
// maximum VM resource limits.
func (d *DataSourceVMOSAvailable) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VMOSAvailableModel

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
	siteParam := client.GetAvailableOSParamsSite(site)

	platform := data.Platform.ValueString()

	// Call API
	osResp, err := d.client.GetAvailableOSWithResponse(ctx, client.GetAvailableOSParamsPlatform(platform), &client.GetAvailableOSParams{
		Site: &siteParam,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read available OS list: %s", err),
		)
		return
	}

	if osResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("API returned status %d: %s", osResp.StatusCode(), string(osResp.Body)),
		)
		return
	}

	if osResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse OS list response. Body: %s", string(osResp.Body)),
		)
		return
	}

	tflog.Debug(ctx, "Received OS list response", map[string]any{
		"platform": platform,
		"site":     site,
		"status":   osResp.JSON200.Status,
	})

	// Set basic fields
	data.ID = types.StringValue(fmt.Sprintf("%s-%s", platform, site))
	data.Site = types.StringValue(site)
	data.Platform = types.StringValue(platform)

	// Helper function to convert string slice to List
	asTFList := func(versions *[]string) types.List {
		if versions == nil || len(*versions) < 1 {
			return types.ListNull(types.StringType)
		}

		versionList := make([]attr.Value, 0, len(*versions))
		for _, version := range *versions {
			versionList = append(versionList, types.StringValue(version))
		}
		listValue, diags := types.ListValue(types.StringType, versionList)
		resp.Diagnostics.Append(diags...)

		return listValue
	}

	// Map each OS distribution
	if osResp.JSON200.OperatingSystems != nil {
		data.CentOS = asTFList(osResp.JSON200.OperatingSystems.CentOS)
		data.RedHat = asTFList(osResp.JSON200.OperatingSystems.RedHat)
		data.Rocky = asTFList(osResp.JSON200.OperatingSystems.Rocky)
		data.SLES = asTFList(osResp.JSON200.OperatingSystems.SLES)
		data.Ubuntu = asTFList(osResp.JSON200.OperatingSystems.Ubuntu)
		data.Windows = asTFList(osResp.JSON200.OperatingSystems.Windows)
	} else {
		data.CentOS = types.ListNull(types.StringType)
		data.RedHat = types.ListNull(types.StringType)
		data.Rocky = types.ListNull(types.StringType)
		data.SLES = types.ListNull(types.StringType)
		data.Ubuntu = types.ListNull(types.StringType)
		data.Windows = types.ListNull(types.StringType)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Map default_size
	defaultSizeModel := DefaultSizeModel{
		Count:            types.StringNull(),
		CPU:              types.StringNull(),
		Memory:           types.StringNull(),
		MaxCount:         types.StringNull(),
		MaxCPU:           types.StringNull(),
		MaxMemory:        types.StringNull(),
		MaxDiskCount:     types.StringNull(),
		MaxDiskSize:      types.StringNull(),
		MaxTotalDiskSize: types.StringNull(),
		PVMMaxCPU:        types.StringNull(),
		PVMMaxMemory:     types.StringNull(),
	}

	if osResp.JSON200.DefaultSize != nil && osResp.JSON200.DefaultSize.Vm != nil {
		vm := osResp.JSON200.DefaultSize.Vm
		if vm.Count != nil {
			defaultSizeModel.Count = types.StringValue(*vm.Count)
		}
		if vm.Cpu != nil {
			defaultSizeModel.CPU = types.StringValue(*vm.Cpu)
		}
		if vm.Memory != nil {
			defaultSizeModel.Memory = types.StringValue(*vm.Memory)
		}
		if vm.MaxCount != nil {
			defaultSizeModel.MaxCount = types.StringValue(*vm.MaxCount)
		}
		if vm.MaxCpu != nil {
			defaultSizeModel.MaxCPU = types.StringValue(*vm.MaxCpu)
		}
		if vm.MaxMemory != nil {
			defaultSizeModel.MaxMemory = types.StringValue(*vm.MaxMemory)
		}
		if vm.MaxDiskCount != nil {
			defaultSizeModel.MaxDiskCount = types.StringValue(*vm.MaxDiskCount)
		}
		if vm.MaxDiskSize != nil {
			defaultSizeModel.MaxDiskSize = types.StringValue(*vm.MaxDiskSize)
		}
		if vm.MaxTotalDiskSize != nil {
			defaultSizeModel.MaxTotalDiskSize = types.StringValue(*vm.MaxTotalDiskSize)
		}
		if vm.Pvm != nil {
			if vm.Pvm.MaxCpu != nil {
				defaultSizeModel.PVMMaxCPU = types.StringValue(*vm.Pvm.MaxCpu)
			}
			if vm.Pvm.MaxMemory != nil {
				defaultSizeModel.PVMMaxMemory = types.StringValue(*vm.Pvm.MaxMemory)
			}
		}
	}

	defaultSizeObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"count":               types.StringType,
		"cpu":                 types.StringType,
		"memory":              types.StringType,
		"max_count":           types.StringType,
		"max_cpu":             types.StringType,
		"max_memory":          types.StringType,
		"max_disk_count":      types.StringType,
		"max_disk_size":       types.StringType,
		"max_total_disk_size": types.StringType,
		"pvm_max_cpu":         types.StringType,
		"pvm_max_memory":      types.StringType,
	}, defaultSizeModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.DefaultSize = defaultSizeObj

	tflog.Trace(ctx, "read vm_os_available data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
