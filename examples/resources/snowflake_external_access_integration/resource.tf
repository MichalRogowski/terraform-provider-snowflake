## Minimal
resource "snowflake_external_access_integration" "basic" {
  name                  = "external_access_integration_name"
  allowed_network_rules = [snowflake_network_rule.one.fully_qualified_name]
  enabled               = true
}

## Complete (with every optional set)
resource "snowflake_external_access_integration" "complete" {
  name                                    = "external_access_integration_name"
  allowed_network_rules                   = [snowflake_network_rule.one.fully_qualified_name]
  allowed_api_authentication_integrations = [snowflake_api_authentication_integration_with_client_credentials.one.fully_qualified_name]
  allowed_authentication_secrets          = [snowflake_secret_with_basic_authentication.one.fully_qualified_name]
  enabled                                 = true
  comment                                 = "my external access integration"
}
