# gh-migrate-variables

`gh-migrate-variables` is a [GitHub CLI](https://cli.github.com) extension to assist in the migration of variables between GitHub organizations. While [GitHub Enterprise Importer](https://github.com/github/gh-gei) provides excellent features for organization migration, there are gaps when it comes to migrating GitHub Actions variables. This extension aims to fill those gaps. Whether you're consolidating organizations, setting up new environments, or need to replicate variables across organizations, this extension can help.

## Install

```bash
gh extension install mona-actions/gh-migrate-variables
```

## Usage: Export

Export organization-level and repository-level variables to a CSV file.

```bash
Usage:
  migrate-variables export [flags]

Flags:
  -f, --file-prefix string    Output filenames prefix
  -h, --help                  help for export
  -u, --hostname string       GitHub Enterprise hostname url (optional) Ex. https://github.example.com
  -o, --organization string   Organization to export (required)
  -t, --token string          GitHub token (required)
```

### Example Export Command

```bash
gh migrate-variables export \
  --organization my-org \
  --token ghp_xxxxxxxxxxxx \
  --file-prefix my-vars
```

This will create a file named `my-vars_variables.csv` containing all organization and repository variables. The export process provides detailed feedback:

```
📊 Export Summary:
Total repositories found: 25
✅ Successfully processed: 23 repositories
❌ Failed to process: 2 repositories
📝 Total variables exported: 23
📁 Output file: my-vars_variables.csv
```

## Usage: Sync

Recreates variables from a CSV file to a target organization, maintaining visibility settings and scopes.

```bash
Usage:
  migrate-variables sync [flags]

Flags:
  -h, --help                         help for sync
  -m, --mapping-file string          Mapping file path to use for syncing variables (required)
  -t, --target-organization string   Target Organization to sync variables to (required)
  -b, --target-token string          Target Organization GitHub token. Scopes: admin:org (required)
```

### Example Sync Command

```bash
gh migrate-variables sync \
  --mapping-file my-vars_variables.csv \
  --target-organization target-org \
  --target-token ghp_xxxxxxxxxxxx
```

The sync process provides detailed feedback:

```
📊 Sync Summary:
Total variables processed: 45
✅ Successfully created: 42
❌ Failed: 2
⚠️ Skipped: 1
```

### Variables CSV Format

The tool exports and imports variables using the following CSV format:

```csv
Name,Value,Scope,Visibility
ORG_VAR,org-value,organization,all
REPO_VAR,repo-value,repository-name,private
```

- `Scope`: Use "organization" for org-level variables, or the repository name for repo-level variables
- `Visibility`: One of "all", "private", or "selected" for org variables; always "private" for repo variables

## Required Permissions

### For Export
- Organization variables: `read:org`
- Repository variables: `repo`

### For Sync
- `admin:org` scope is required for creating organization variables
- `repo` scope is required for creating repository variables


## Limitations

- Repository-level variables can only be created if the repository exists in the target organization
- Environment-specific variables should be reviewed before syncing to ensure appropriate values
- Repository visibility settings must be considered when setting organization variable visibility
- The tool will retry failed API calls but may still encounter persistent access issues for specific repositories

## License

- [MIT](./license) (c) [Mona-Actions](https://github.com/mona-actions)
- [Contributing](./contributing.md)