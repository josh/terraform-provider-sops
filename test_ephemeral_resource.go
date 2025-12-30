package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ ephemeral.EphemeralResource = &TestDynamicEphemeralResource{}

type TestDynamicEphemeralResource struct{}

type TestDynamicEphemeralResourceModel struct {
	Value  types.Dynamic `tfsdk:"value"`
	Output types.Dynamic `tfsdk:"output"`
}

func NewTestDynamicEphemeralResource() ephemeral.EphemeralResource {
	return &TestDynamicEphemeralResource{}
}

func (r *TestDynamicEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_test_dynamic"
}

func (r *TestDynamicEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Test-only ephemeral resource that echoes input to output for testing unknown values",
		Attributes: map[string]schema.Attribute{
			"value": schema.DynamicAttribute{
				MarkdownDescription: "Input value to echo",
				Required:            true,
			},
			"output": schema.DynamicAttribute{
				MarkdownDescription: "Output value (echoes input)",
				Computed:            true,
			},
		},
	}
}

func (r *TestDynamicEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data TestDynamicEphemeralResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Output = data.Value

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
