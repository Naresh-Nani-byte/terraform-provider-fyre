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
var _ datasource.DataSource = &DataSourceUser{}

// NewDataSourceUser creates a new instance of the user data source.
// This factory function is called by the provider to instantiate the data source.
func NewDataSourceUser() datasource.DataSource {
	return &DataSourceUser{}
}

// DataSourceUser defines the data source implementation.
type DataSourceUser struct {
	client      *client.ClientWithResponses
	defaultSite string
}

// UserModel describes the data source data model.
type UserModel struct {
	ID                      types.String `tfsdk:"id"`
	Site                    types.String `tfsdk:"site"`
	Authenticated           types.Bool   `tfsdk:"authenticated"`
	ClientIp                types.String `tfsdk:"client_ip"`
	Email                   types.String `tfsdk:"email"`
	FullName                types.String `tfsdk:"full_name"`
	Login                   types.String `tfsdk:"login"`
	LoginExpiration         types.Int64  `tfsdk:"login_expiration"`
	Programs                types.List   `tfsdk:"programs"`
	UnixExpirationTimestamp types.Int64  `tfsdk:"unix_expiration_timestamp"`
	Development             types.Object `tfsdk:"development"`
	Sentry                  types.Object `tfsdk:"sentry"`
}

// DevelopmentModel describes the development nested object.
type DevelopmentModel struct {
	Authenticated             types.Bool   `tfsdk:"authenticated"`
	Authorizations            types.List   `tfsdk:"authorizations"`
	DefaultLocation           types.String `tfsdk:"default_location"`
	DefaultPasswordExpiration types.String `tfsdk:"default_password_expiration"`
	DefaultProductGroupId     types.Int64  `tfsdk:"default_product_group_id"`
	DisplayName               types.String `tfsdk:"display_name"`
	Email                     types.String `tfsdk:"email"`
	Id                        types.Int64  `tfsdk:"id"`
	Login                     types.String `tfsdk:"login"`
	LoginExpiration           types.Int64  `tfsdk:"login_expiration"`
	PasswordSet               types.String `tfsdk:"password_set"`
	ProductGroups             types.List   `tfsdk:"product_groups"`
	ProductGroupOwner         types.String `tfsdk:"product_group_owner"`
	QuickBurn                 types.String `tfsdk:"quick_burn"`
	Roles                     types.List   `tfsdk:"roles"`
	Settings                  types.List   `tfsdk:"settings"`
	Username                  types.String `tfsdk:"username"`
	ValidApiKey               types.String `tfsdk:"valid_api_key"`
}

// SentryModel describes the sentry nested object.
type SentryModel struct {
	TwoFaStatus      types.String `tfsdk:"two_fa_status"`
	Access           types.String `tfsdk:"access"`
	AuthMethod       types.String `tfsdk:"auth_method"`
	Authenticator    types.String `tfsdk:"authenticator"`
	Default2faMethod types.String `tfsdk:"default_2fa_method"`
	Status           types.String `tfsdk:"status"`
	UserStatus       types.String `tfsdk:"user_status"`
}

// ProductGroupsModel is
type ProductGroupsModel struct {
	ID          types.Int32  `tfsdk:"id"`
	ProductName types.String `tfsdk:"product_name"`
	GroupName   types.String `tfsdk:"group_name"`
}

// Metadata sets the data source type name for the user data source.
// The type name is used in Terraform configurations as "fyre_user".
func (d *DataSourceUser) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the structure and attributes of the user data source.
// It specifies computed attributes including authentication status, email, login information,
// development environment details (authorizations, product groups, roles), and sentry
// authentication details (2FA status, access level).
func (d *DataSourceUser) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches user details for the authenticated user, including authentication status, email, login information, and development/sentry details.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The user ID if one is returned",
				Computed:            true,
			},
			"site": schema.StringAttribute{
				MarkdownDescription: "Site location where user details were fetched from",
				Computed:            true,
			},
			"authenticated": schema.BoolAttribute{
				MarkdownDescription: "Whether the user is authenticated",
				Computed:            true,
			},
			"client_ip": schema.StringAttribute{
				MarkdownDescription: "Client IP address",
				Computed:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "User's email address",
				Computed:            true,
			},
			"full_name": schema.StringAttribute{
				MarkdownDescription: "User's full name",
				Computed:            true,
			},
			"login": schema.StringAttribute{
				MarkdownDescription: "User's login identifier",
				Computed:            true,
			},
			"login_expiration": schema.Int64Attribute{
				MarkdownDescription: "Login expiration timestamp",
				Computed:            true,
			},
			"programs": schema.ListAttribute{
				MarkdownDescription: "List of programs the user has access to",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"unix_expiration_timestamp": schema.Int64Attribute{
				MarkdownDescription: "Unix timestamp for expiration",
				Computed:            true,
			},
			"development": schema.SingleNestedAttribute{
				MarkdownDescription: "Development environment details",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"authenticated": schema.BoolAttribute{
						MarkdownDescription: "Development environment authentication status",
						Computed:            true,
						Optional:            true,
					},
					"authorizations": schema.ListAttribute{
						MarkdownDescription: "Development environment authorizations",
						Computed:            true,
						ElementType:         types.StringType,
						Optional:            true,
					},
					"default_location": schema.StringAttribute{
						MarkdownDescription: "Development environment default location",
						Computed:            true,
						Optional:            true,
					},
					"default_password_expiration": schema.StringAttribute{
						MarkdownDescription: "Development environment default password expiration",
						Computed:            true,
						Optional:            true,
					},
					"default_product_group_id": schema.Int64Attribute{
						MarkdownDescription: "Development environment default product group ID",
						Computed:            true,
						Optional:            true,
					},
					"display_name": schema.StringAttribute{
						MarkdownDescription: "Development environment display name",
						Computed:            true,
						Optional:            true,
					},
					"email": schema.StringAttribute{
						MarkdownDescription: "Development environment email",
						Computed:            true,
						Optional:            true,
					},
					"id": schema.Int64Attribute{
						MarkdownDescription: "Development environment user ID",
						Computed:            true,
						Optional:            true,
					},
					"login": schema.StringAttribute{
						MarkdownDescription: "Development environment login",
						Computed:            true,
						Optional:            true,
					},
					"login_expiration": schema.Int64Attribute{
						MarkdownDescription: "Development environment login expiration",
						Computed:            true,
						Optional:            true,
					},
					"password_set": schema.StringAttribute{
						MarkdownDescription: "Development environment password set status",
						Computed:            true,
						Optional:            true,
					},
					"product_groups": schema.ListNestedAttribute{
						MarkdownDescription: "User product groups",
						Computed:            true,
						Optional:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.Int32Attribute{
									MarkdownDescription: "Product group ID",
									Computed:            true,
									Optional:            true,
								},
								"product_name": schema.StringAttribute{
									MarkdownDescription: "Product name",
									Computed:            true,
									Optional:            true,
								},
								"group_name": schema.StringAttribute{
									MarkdownDescription: "Group name",
									Computed:            true,
									Optional:            true,
								},
							},
						},
					},
					"product_group_owner": schema.StringAttribute{
						MarkdownDescription: "Development environment product group owner",
						Computed:            true,
						Optional:            true,
					},
					"quick_burn": schema.StringAttribute{
						MarkdownDescription: "Development environment quick burn status",
						Computed:            true,
						Optional:            true,
					},
					"roles": schema.ListAttribute{
						MarkdownDescription: "Development environment roles",
						Computed:            true,
						ElementType:         types.StringType,
						Optional:            true,
					},
					"settings": schema.ListAttribute{
						MarkdownDescription: "Development environment settings",
						Computed:            true,
						ElementType:         types.StringType,
						Optional:            true,
					},
					"username": schema.StringAttribute{
						MarkdownDescription: "Development environment username",
						Computed:            true,
						Optional:            true,
					},
					"valid_api_key": schema.StringAttribute{
						MarkdownDescription: "Development environment valid API key status",
						Computed:            true,
						Optional:            true,
					},
				},
			},
			"sentry": schema.SingleNestedAttribute{
				MarkdownDescription: "Sentry authentication details",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"two_fa_status": schema.StringAttribute{
						MarkdownDescription: "Sentry 2FA status",
						Computed:            true,
					},
					"access": schema.StringAttribute{
						MarkdownDescription: "Sentry access level",
						Computed:            true,
					},
					"auth_method": schema.StringAttribute{
						MarkdownDescription: "Sentry authentication method",
						Computed:            true,
					},
					"authenticator": schema.StringAttribute{
						MarkdownDescription: "Sentry authenticator",
						Computed:            true,
					},
					"default_2fa_method": schema.StringAttribute{
						MarkdownDescription: "Sentry default 2FA method",
						Computed:            true,
					},
					"status": schema.StringAttribute{
						MarkdownDescription: "Sentry status",
						Computed:            true,
					},
					"user_status": schema.StringAttribute{
						MarkdownDescription: "Sentry user status",
						Computed:            true,
					},
				},
			},
		},
	}
}

// Configure initializes the user data source with the Fyre API client
// and default site configuration from the provider. This method is called by the
// framework during provider initialization.
func (d *DataSourceUser) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read retrieves authenticated user details from the Fyre API including authentication
// status, email, login information, development environment details (authorizations,
// product groups, roles, settings), and sentry authentication information (2FA status,
// access level, authentication method).
func (d *DataSourceUser) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare API request parameters
	site := client.GetUserDetailsParamsSite(data.Site.ValueString())
	if site == "" {
		site = client.GetUserDetailsParamsSite(d.defaultSite)
	}

	// Call API
	userResp, err := d.client.GetUserDetailsWithResponse(ctx, &client.GetUserDetailsParams{
		Site: &site,
	})
	if userResp != nil {
		// Log response for debugging
		tflog.Debug(ctx, "GetUserDetails API response", map[string]any{
			"status_code": userResp.StatusCode(),
			"has_json200": userResp.JSON200 != nil,
		})
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read user details: %s", err),
		)
		return
	}

	// Check HTTP status
	if userResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("API returned status %d: %s", userResp.StatusCode(), string(userResp.Body)),
		)
		return
	}

	// Parse response
	if userResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse user details response. Body: %s", string(userResp.Body)),
		)
		return
	}

	// Map to Terraform state
	data.ID = types.StringValue("user")
	data.Site = types.StringValue(d.defaultSite)

	// Initialize all fields with null/empty values
	data.Authenticated = types.BoolNull()
	data.ClientIp = types.StringNull()
	data.Email = types.StringNull()
	data.FullName = types.StringNull()
	data.Login = types.StringNull()
	data.LoginExpiration = types.Int64Null()
	data.UnixExpirationTimestamp = types.Int64Null()
	emptyList, diags := types.ListValue(types.StringType, []attr.Value{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Programs = emptyList
	data.Development = types.ObjectNull(map[string]attr.Type{
		"authenticated":               types.BoolType,
		"authorizations":              types.ListType{ElemType: types.StringType},
		"default_location":            types.StringType,
		"default_password_expiration": types.StringType,
		"default_product_group_id":    types.Int64Type,
		"display_name":                types.StringType,
		"email":                       types.StringType,
		"id":                          types.Int64Type,
		"login":                       types.StringType,
		"login_expiration":            types.Int64Type,
		"password_set":                types.StringType,
		"product_groups": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
			"id":           types.Int32Type,
			"product_name": types.StringType,
			"group_name":   types.StringType,
		}}},
		"product_group_owner": types.StringType,
		"quick_burn":          types.StringType,
		"roles":               types.ListType{ElemType: types.StringType},
		"settings":            types.ListType{ElemType: types.StringType},
		"username":            types.StringType,
		"valid_api_key":       types.StringType,
	})
	data.Sentry = types.ObjectNull(map[string]attr.Type{
		"two_fa_status":      types.StringType,
		"access":             types.StringType,
		"auth_method":        types.StringType,
		"authenticator":      types.StringType,
		"default_2fa_method": types.StringType,
		"status":             types.StringType,
		"user_status":        types.StringType,
	})

	// Extract uRes details if available
	if uRes := userResp.JSON200; uRes != nil {
		// Top-level fields
		if uRes.Authenticated != nil {
			data.Authenticated = types.BoolValue(*uRes.Authenticated)
		}
		if uRes.ClientIp != nil {
			data.ClientIp = types.StringValue(*uRes.ClientIp)
		}
		if uRes.Email != nil {
			data.Email = types.StringValue(*uRes.Email)
		}
		if uRes.FullName != nil {
			data.FullName = types.StringValue(*uRes.FullName)
		}
		if uRes.Login != nil {
			data.Login = types.StringValue(*uRes.Login)
		}
		if uRes.LoginExpiration != nil {
			data.LoginExpiration = types.Int64Value(int64(*uRes.LoginExpiration))
		}
		if uRes.UnixExpirationTimestamp != nil {
			data.UnixExpirationTimestamp = types.Int64Value(int64(*uRes.UnixExpirationTimestamp))
		}

		// Programs list
		if uRes.Programs != nil && len(*uRes.Programs) > 0 {
			programsList := make([]attr.Value, 0, len(*uRes.Programs))
			for _, prog := range *uRes.Programs {
				programsList = append(programsList, types.StringValue(prog))
			}
			listValue, diags := types.ListValue(types.StringType, programsList)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Programs = listValue
			}
		}

		// Development nested object
		if dev := uRes.Development; dev != nil {
			devModel := DevelopmentModel{
				Authenticated:             types.BoolNull(),
				Authorizations:            emptyList,
				DefaultLocation:           types.StringNull(),
				DefaultPasswordExpiration: types.StringNull(),
				DefaultProductGroupId:     types.Int64Null(),
				DisplayName:               types.StringNull(),
				Email:                     types.StringNull(),
				Id:                        types.Int64Null(),
				Login:                     types.StringNull(),
				LoginExpiration:           types.Int64Null(),
				PasswordSet:               types.StringNull(),
				ProductGroups: types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{
					"id":           types.Int32Type,
					"product_name": types.StringType,
					"group_name":   types.StringType,
				}}),
				ProductGroupOwner: types.StringNull(),
				QuickBurn:         types.StringNull(),
				Roles:             emptyList,
				Settings:          emptyList,
				Username:          types.StringNull(),
				ValidApiKey:       types.StringNull(),
			}

			if dev.Authenticated != nil {
				devModel.Authenticated = types.BoolValue(*dev.Authenticated)
			}
			if dev.DefaultLocation != nil {
				devModel.DefaultLocation = types.StringValue(*dev.DefaultLocation)
			}
			if dev.DefaultPasswordExpiration != nil {
				devModel.DefaultPasswordExpiration = types.StringValue(*dev.DefaultPasswordExpiration)
			}
			if dev.DefaultProductGroupId != nil {
				devModel.DefaultProductGroupId = types.Int64Value(int64(*dev.DefaultProductGroupId))
			}
			if dev.DisplayName != nil {
				devModel.DisplayName = types.StringValue(*dev.DisplayName)
			}
			if dev.Email != nil {
				devModel.Email = types.StringValue(*dev.Email)
			}
			if dev.Id != nil {
				devModel.Id = types.Int64Value(int64(*dev.Id))
				data.ID = types.StringValue(fmt.Sprintf("%d", *dev.Id))
			}
			if dev.Login != nil {
				devModel.Login = types.StringValue(*dev.Login)
			}
			if dev.LoginExpiration != nil {
				devModel.LoginExpiration = types.Int64Value(int64(*dev.LoginExpiration))
			}
			if dev.PasswordSet != nil {
				devModel.PasswordSet = types.StringValue(*dev.PasswordSet)
			}
			if dev.ProductGroups != nil && len(*dev.ProductGroups) > 0 {
				prodGroupsList := make([]attr.Value, 0, len(*dev.ProductGroups))
				for _, pg := range *dev.ProductGroups {
					var idVal types.Int32
					if pg.Id != nil {
						idVal = types.Int32Value(int32(*pg.Id))
					} else {
						idVal = types.Int32Null()
					}

					pgObj, diags := types.ObjectValue(
						map[string]attr.Type{
							"id":           types.Int32Type,
							"product_name": types.StringType,
							"group_name":   types.StringType,
						},
						map[string]attr.Value{
							"id":           idVal,
							"product_name": types.StringPointerValue(pg.ProductName),
							"group_name":   types.StringPointerValue(pg.GroupName),
						},
					)
					resp.Diagnostics.Append(diags...)
					if !resp.Diagnostics.HasError() {
						prodGroupsList = append(prodGroupsList, pgObj)
					}
				}
				if !resp.Diagnostics.HasError() {
					listValue, diags := types.ListValue(
						types.ObjectType{AttrTypes: map[string]attr.Type{
							"id":           types.Int32Type,
							"product_name": types.StringType,
							"group_name":   types.StringType,
						}},
						prodGroupsList,
					)
					resp.Diagnostics.Append(diags...)
					if !resp.Diagnostics.HasError() {
						devModel.ProductGroups = listValue
					}
				}
			}
			if dev.ProductGroupOwner != nil {
				devModel.ProductGroupOwner = types.StringValue(*dev.ProductGroupOwner)
			}
			if dev.QuickBurn != nil {
				devModel.QuickBurn = types.StringValue(*dev.QuickBurn)
			}
			if dev.Username != nil {
				devModel.Username = types.StringValue(*dev.Username)
			}
			if dev.ValidApiKey != nil {
				devModel.ValidApiKey = types.StringValue(*dev.ValidApiKey)
			}

			// Development list fields
			if dev.Authorizations != nil && len(*dev.Authorizations) > 0 {
				authList := make([]attr.Value, 0, len(*dev.Authorizations))
				for _, auth := range *dev.Authorizations {
					authList = append(authList, types.StringValue(auth))
				}
				listValue, diags := types.ListValue(types.StringType, authList)
				resp.Diagnostics.Append(diags...)
				if !resp.Diagnostics.HasError() {
					devModel.Authorizations = listValue
				}
			}

			if dev.Roles != nil && len(*dev.Roles) > 0 {
				rolesList := make([]attr.Value, 0, len(*dev.Roles))
				for _, role := range *dev.Roles {
					rolesList = append(rolesList, types.StringValue(role))
				}
				listValue, diags := types.ListValue(types.StringType, rolesList)
				resp.Diagnostics.Append(diags...)
				if !resp.Diagnostics.HasError() {
					devModel.Roles = listValue
				}
			}

			if dev.Settings != nil && len(*dev.Settings) > 0 {
				settingsList := make([]attr.Value, 0, len(*dev.Settings))
				for _, setting := range *dev.Settings {
					settingsList = append(settingsList, types.StringValue(setting))
				}
				listValue, diags := types.ListValue(types.StringType, settingsList)
				resp.Diagnostics.Append(diags...)
				if !resp.Diagnostics.HasError() {
					devModel.Settings = listValue
				}
			}

			// Convert to object
			devObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
				"authenticated":               types.BoolType,
				"authorizations":              types.ListType{ElemType: types.StringType},
				"default_location":            types.StringType,
				"default_password_expiration": types.StringType,
				"default_product_group_id":    types.Int64Type,
				"display_name":                types.StringType,
				"email":                       types.StringType,
				"id":                          types.Int64Type,
				"login":                       types.StringType,
				"login_expiration":            types.Int64Type,
				"password_set":                types.StringType,
				"product_groups": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
					"id":           types.Int32Type,
					"product_name": types.StringType,
					"group_name":   types.StringType,
				}}},
				"product_group_owner": types.StringType,
				"quick_burn":          types.StringType,
				"roles":               types.ListType{ElemType: types.StringType},
				"settings":            types.ListType{ElemType: types.StringType},
				"username":            types.StringType,
				"valid_api_key":       types.StringType,
			}, devModel)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Development = devObj
			}
		}

		// Sentry nested object
		if uRes.Sentry != nil {
			sentry := uRes.Sentry
			sentryModel := SentryModel{
				TwoFaStatus:      types.StringNull(),
				Access:           types.StringNull(),
				AuthMethod:       types.StringNull(),
				Authenticator:    types.StringNull(),
				Default2faMethod: types.StringNull(),
				Status:           types.StringNull(),
				UserStatus:       types.StringNull(),
			}

			if sentry.N2faStatus != nil {
				sentryModel.TwoFaStatus = types.StringValue(*sentry.N2faStatus)
			}
			if sentry.Access != nil {
				sentryModel.Access = types.StringValue(*sentry.Access)
			}
			if sentry.AuthMethod != nil {
				sentryModel.AuthMethod = types.StringValue(*sentry.AuthMethod)
			}
			if sentry.Authenticator != nil {
				sentryModel.Authenticator = types.StringValue(*sentry.Authenticator)
			}
			if sentry.Default2faMethod != nil {
				sentryModel.Default2faMethod = types.StringValue(*sentry.Default2faMethod)
			}
			if sentry.Status != nil {
				sentryModel.Status = types.StringValue(*sentry.Status)
			}
			if sentry.UserStatus != nil {
				sentryModel.UserStatus = types.StringValue(*sentry.UserStatus)
			}

			// Convert to object
			sentryObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
				"two_fa_status":      types.StringType,
				"access":             types.StringType,
				"auth_method":        types.StringType,
				"authenticator":      types.StringType,
				"default_2fa_method": types.StringType,
				"status":             types.StringType,
				"user_status":        types.StringType,
			}, sentryModel)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.Sentry = sentryObj
			}
		}
	}

	// Write logs using the tflog package
	tflog.Trace(ctx, "read user data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
