data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./models",
    "--dialect", "postgres",
  ]
}

env "gorm" {
  src = data.external_schema.gorm.url
  dev = "postgres://migrationsdev:migrationsdev@postgres_migrations_dev/migrationsdev?sslmode=disable"
  migration {
    dir = "file://db-migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}