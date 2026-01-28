resource "thecloud_secret" "db_pass" {
  name        = "DB_PASSWORD"
  value       = "secret-value-123"
  description = "Production database password"
}
