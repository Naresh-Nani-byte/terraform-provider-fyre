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

var _ datasource.DataSource = &DataSourceVMSnapshots{}

func NewDataSourceVMSnapshots() datasource.DataSource {
	return &DataSourceVMSnapshots{}
}

type DataSourceVMSnapshots struct {
	client      *client.ClientWithResponses
	defaultSite string
}

type DataSourceVMSnapshotsModel struct {
	ID            types.String `tfsdk:"id"`
	VMID          types.String `tfsdk:"vm_id"`
	Site          types.String `tfsdk:"site"`
	SnapshotCount types.String `tfsdk:"snapshot_count"`
	SnapshotLimit types.String `tfsdk:"snapshot_limit"`
	Snapshots     types.List   `tfsdk:"snapshots"`
}

type SnapshotModel struct {
	Name         types.String `tfsdk:"name"`
	Comment      types.String `tfsdk:"comment"`
	FullSnapshot types.String `tfsdk:"full_snapshot"`
	Active       types.String `tfsdk:"active"`
	Used         types.Int64  `tfsdk:"used"`
	Enabled      types.String `tfsdk:"enabled"`
	Created      types.String `tfsdk:"created"`
}

func (d *DataSourceVMSnapshots) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_snapshots"
}

func (d *DataSourceVMSnapshots) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of snapshots for a VM. The VM can be identified by VM ID (format: x-xxxxxxx), IP address, or fully qualified domain name (FQDN) that is in DNS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The VM identifier used for this data source",
				Computed:            true,
			},
			"vm_id": schema.StringAttribute{
				MarkdownDescription: "VM identifier (can be VM ID like '1-8103661', IP address, or FQDN in DNS)",
				Required:            true,
			},
			"site": schema.StringAttribute{
				MarkdownDescription: "Site location (svl or rtp). Defaults to 'svl' or inherits from provider configuration.",
				Optional:            true,
				Computed:            true,
			},
			"snapshot_count": schema.StringAttribute{
				MarkdownDescription: "Number of snapshots currently created for this VM",
				Computed:            true,
				Optional:            true,
			},
			"snapshot_limit": schema.StringAttribute{
				MarkdownDescription: "Maximum number of snapshots allowed for this VM",
				Computed:            true,
				Optional:            true,
			},
			"snapshots": schema.ListNestedAttribute{
				MarkdownDescription: "List of snapshots for the VM",
				Computed:            true,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Snapshot name",
							Computed:            true,
							Optional:            true,
						},
						"comment": schema.StringAttribute{
							MarkdownDescription: "Snapshot comment/description",
							Computed:            true,
							Optional:            true,
						},
						"full_snapshot": schema.StringAttribute{
							MarkdownDescription: "Whether this is a full snapshot - 'true' or 'false' as string",
							Computed:            true,
							Optional:            true,
						},
						"active": schema.StringAttribute{
							MarkdownDescription: "Whether snapshot is active - 'true' or 'false' as string",
							Computed:            true,
							Optional:            true,
						},
						"used": schema.Int64Attribute{
							MarkdownDescription: "Disk space used by snapshot in bytes",
							Computed:            true,
							Optional:            true,
						},
						"enabled": schema.StringAttribute{
							MarkdownDescription: "Whether snapshot is enabled - 'y' or 'n'",
							Computed:            true,
							Optional:            true,
						},
						"created": schema.StringAttribute{
							MarkdownDescription: "Snapshot creation timestamp",
							Computed:            true,
							Optional:            true,
						},
					},
				},
			},
		},
	}
}

func (d *DataSourceVMSnapshots) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataSourceVMSnapshots) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DataSourceVMSnapshotsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare API request parameters
	site := client.GetVMSnapshotsParamsSite(data.Site.ValueString())
	if site == "" {
		site = client.GetVMSnapshotsParamsSite(d.defaultSite)
	}

	vmIdentifier := client.VmIdentifier(data.VMID.ValueString())

	// Call API
	snapshotsResp, err := d.client.GetVMSnapshotsWithResponse(ctx, vmIdentifier, &client.GetVMSnapshotsParams{
		Site: &site,
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read VM snapshots: %s", err),
		)
		return
	}

	if snapshotsResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("API returned status %d: %s", snapshotsResp.StatusCode(), string(snapshotsResp.Body)),
		)
		return
	}

	if snapshotsResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse VM snapshots response. Body: %s", string(snapshotsResp.Body)),
		)
		return
	}

	tflog.Debug(ctx, "VM snapshots API response", map[string]any{
		"vm_identifier":  vmIdentifier,
		"snapshot_count": snapshotsResp.JSON200.SnapshotCount,
		"snapshot_limit": snapshotsResp.JSON200.SnapshotLimit,
	})

	// Map to Terraform state
	data.ID = types.StringValue(string(vmIdentifier))
	data.Site = types.StringValue(string(site))

	// Map snapshot count and limit
	if snapshotsResp.JSON200.SnapshotCount != nil {
		data.SnapshotCount = types.StringValue(*snapshotsResp.JSON200.SnapshotCount)
	} else {
		data.SnapshotCount = types.StringNull()
	}

	if snapshotsResp.JSON200.SnapshotLimit != nil {
		data.SnapshotLimit = types.StringValue(*snapshotsResp.JSON200.SnapshotLimit)
	} else {
		data.SnapshotLimit = types.StringNull()
	}

	// Map snapshots list
	if snapshotsResp.JSON200.Snapshots != nil && len(*snapshotsResp.JSON200.Snapshots) > 0 {
		snapshotsList := make([]SnapshotModel, 0, len(*snapshotsResp.JSON200.Snapshots))
		for _, snapshot := range *snapshotsResp.JSON200.Snapshots {
			snapshotModel := SnapshotModel{
				Name:         types.StringNull(),
				Comment:      types.StringNull(),
				FullSnapshot: types.StringNull(),
				Active:       types.StringNull(),
				Used:         types.Int64Null(),
				Enabled:      types.StringNull(),
				Created:      types.StringNull(),
			}

			if snapshot.Name != nil {
				snapshotModel.Name = types.StringValue(*snapshot.Name)
			}
			if snapshot.Comment != nil {
				snapshotModel.Comment = types.StringValue(*snapshot.Comment)
			}
			if snapshot.FullSnapshot != nil {
				snapshotModel.FullSnapshot = types.StringValue(*snapshot.FullSnapshot)
			}
			if snapshot.Active != nil {
				snapshotModel.Active = types.StringValue(*snapshot.Active)
			}
			if snapshot.Used != nil {
				snapshotModel.Used = types.Int64Value(int64(*snapshot.Used))
			}
			if snapshot.Enabled != nil {
				snapshotModel.Enabled = types.StringValue(*snapshot.Enabled)
			}
			if snapshot.Created != nil {
				snapshotModel.Created = types.StringValue(*snapshot.Created)
			}

			snapshotsList = append(snapshotsList, snapshotModel)
		}

		// Convert to types.List
		snapshotsListValue, diags := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name":          types.StringType,
				"comment":       types.StringType,
				"full_snapshot": types.StringType,
				"active":        types.StringType,
				"used":          types.Int64Type,
				"enabled":       types.StringType,
				"created":       types.StringType,
			},
		}, snapshotsList)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Snapshots = snapshotsListValue
		}
	} else {
		// Empty list
		data.Snapshots, _ = types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name":          types.StringType,
				"comment":       types.StringType,
				"full_snapshot": types.StringType,
				"active":        types.StringType,
				"used":          types.Int64Type,
				"enabled":       types.StringType,
				"created":       types.StringType,
			},
		}, []SnapshotModel{})
	}

	tflog.Trace(ctx, "read VM snapshots data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
