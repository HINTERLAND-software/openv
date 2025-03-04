## openv import

Import environment variables into 1Password

### Synopsis

Import environment variables from a specified .env file into 1Password. 
The variables are stored securely with metadata and can be synchronized with different profiles.
Example usage:
  openv import --url github.com/org/repo --env staging --file .env.staging

```
openv import [flags]
```

### Options

```
      --env string              Environment (e.g., production, staging)
      --file string             Path to the environment file to import
  -h, --help                    help for import
      --sync-profiles strings   Sync profiles to use
      --url string              Service URL
      --vault string            1Password vault to use (default "service-account")
```

### Options inherited from parent commands

```
      --config string     config file (default is $HOME/.openv.yaml)
  -j, --json              Output in JSON format
      --op-token string   The 1Password service account token for authentication
  -q, --quiet             Suppress all output except errors
  -v, --verbose           Enable verbose output
```

### SEE ALSO

* [openv](openv.md)	 - OpenV is a CLI tool to manage environment variables in 1Password

###### Auto generated by spf13/cobra on 21-Feb-2025
