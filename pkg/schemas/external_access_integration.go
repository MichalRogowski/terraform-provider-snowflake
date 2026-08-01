package schemas

import (
	"strings"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DescribeExternalAccessIntegrationSchema represents output of DESCRIBE query for the single ExternalAccessIntegration.
var DescribeExternalAccessIntegrationSchema = map[string]*schema.Schema{
	"enabled": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"allowed_network_rules": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"allowed_api_authentication_integrations": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"allowed_authentication_secrets": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"comment": {
		Type:     schema.TypeString,
		Computed: true,
	},
}

var _ = DescribeExternalAccessIntegrationSchema

func ExternalAccessIntegrationPropertiesToSchema(properties []sdk.ExternalAccessIntegrationProperty) map[string]any {
	externalAccessIntegrationSchema := make(map[string]any)
	for _, property := range properties {
		switch property.Name {
		case "ENABLED",
			"ALLOWED_NETWORK_RULES",
			"ALLOWED_API_AUTHENTICATION_INTEGRATIONS",
			"ALLOWED_AUTHENTICATION_SECRETS",
			"COMMENT":
			externalAccessIntegrationSchema[strings.ToLower(property.Name)] = property.Value
		}
	}
	return externalAccessIntegrationSchema
}

var _ = ExternalAccessIntegrationPropertiesToSchema
