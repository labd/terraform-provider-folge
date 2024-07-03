package internal

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"

	"github.com/labd/terraform-provider-folge/internal/application"
	checkhttpstatus "github.com/labd/terraform-provider-folge/internal/check_http_status"
	checkjsonproperty "github.com/labd/terraform-provider-folge/internal/check_json_property"
	folge_datasource "github.com/labd/terraform-provider-folge/internal/datasource"
	"github.com/labd/terraform-provider-folge/internal/folge"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &folgeProvider{}
)

type OptionFunc func(p *folgeProvider)

func WithRetryableClient(retries int) OptionFunc {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = retries

	return func(p *folgeProvider) {
		p.httpClient = retryClient.StandardClient()
	}
}

func WithDebugClient() OptionFunc {
	return func(p *folgeProvider) {
		p.httpClient.Transport = NewDebugTransport(p.httpClient.Transport)
	}
}

func WithRecorderClient(file string, mode recorder.Mode) (OptionFunc, func() error) {
	r, err := recorder.NewWithOptions(&recorder.Options{
		CassetteName:       file,
		Mode:               mode,
		SkipRequestLatency: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	//Strip all fields we are not interested in
	hook := func(i *cassette.Interaction) error {
		i.Response.Headers = cleanHeaders(i.Response.Headers, "Content-Type")
		i.Request.Headers = cleanHeaders(i.Request.Headers)
		return nil
	}
	r.AddHook(hook, recorder.AfterCaptureHook)

	stop := func() error {
		return r.Stop()
	}

	return func(p *folgeProvider) {
		p.httpClient = r.GetDefaultClient()
	}, stop
}

// New is a helper function to simplify provider server and testing implementation.
func New(opts ...OptionFunc) provider.Provider {
	tp := http.DefaultTransport

	var p = &folgeProvider{
		httpClient: &http.Client{Transport: tp},
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// folgeProvider is the provider implementation.
type folgeProvider struct {
	httpClient *http.Client
}

// folgeProviderModel maps provider schema data to a Go type.
type folgeProviderModel struct {
	URL          types.String `tfsdk:"url"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

// Metadata returns the provider type name.
func (p *folgeProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "folge"
}

// Schema defines the provider-level schema for configuration data.
func (p *folgeProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Folge.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "Management API base URL",
				Optional:    true,
			},
			"client_id": schema.StringAttribute{
				Description: "Client ID",
				Optional:    true,
				Sensitive:   true,
			},
			"client_secret": schema.StringAttribute{
				Description: "Client Secret",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure prepares a Folge API client for data sources and resources.
func (p *folgeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Folge client")

	// Retrieve provider data from configuration
	var config folgeProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	url := os.Getenv("FOLGE_URL")
	clientId := os.Getenv("FOLGE_CLIENT_ID")
	clientSecret := os.Getenv("FOLGE_CLIENT_SECRET")

	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	}

	if !config.ClientID.IsNull() {
		clientId = config.ClientID.ValueString()
	}

	if !config.ClientSecret.IsNull() {
		clientSecret = config.ClientSecret.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if resp.Diagnostics.HasError() {
		return
	}

	if url == "" {
		url = "https://app.folge.io"
	}

	ctx = tflog.SetField(ctx, "folge_url", url)
	ctx = tflog.SetField(ctx, "folge_client_id", clientId)
	ctx = tflog.SetField(ctx, "folge_client_secret", clientSecret)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "folge_client_secret")

	tflog.Debug(ctx, "Creating Folge client")

	apiKeyProvider, err := securityprovider.NewSecurityProviderBasicAuth(clientId, clientSecret)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Folge API Client", err.Error())
	}

	// Create a new Folge client using the configuration values
	client, err := folge.NewClientWithResponses(
		url,
		folge.WithHTTPClient(p.httpClient),
		folge.WithRequestEditorFn(apiKeyProvider.Intercept))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Folge API Client",
			"An unexpected error occurred when creating the Folge API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Folge Client Error: "+err.Error(),
		)
		return
	}

	// Make the Folge client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Folge client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *folgeProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// Resources defines the resources implemented in the provider.
func (p *folgeProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		application.NewApplicationResource,
		folge_datasource.NewDataSourceResource,
		checkhttpstatus.NewCheckHttpStatusResource,
		checkjsonproperty.NewCheckJsonPropertyResource,
	}
}
