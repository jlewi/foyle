```bash
Sync the dev infra
```
```bash
pulumi up -C /Users/jlewi/git_foyle/iac/dev --non-interactive --yes

```
```output
exitCode: 0
```
```output
stdout:
Previewing update (dev)

View Live: https://app.pulumi.com/jlewi/foyle-dev/dev/previews/e0d6037d-1bf5-4fbf-92d7-2b1d9f574053


@ Previewing update..................
    pulumi:pulumi:Stack foyle-dev-dev running 
@ Previewing update.....
    pulumi:pulumi:Stack foyle-dev-dev running warning: unable to detect a global setting for GCP Project;
@ Previewing update....
    gcp:organizations:Project project  
    gcp:projects:Service cloudbuild.googleapis.com  
    gcp:projects:Service artifactregistry.googleapis.com  
    gcp:projects:Service container.googleapis.com  
    gcp:projects:Service storage.googleapis.com  
    gcp:projects:Service secretmanager.googleapis.com  
    gcp:container:Cluster dev  
    pulumi:pulumi:Stack foyle-dev-dev  1 warning
Diagnostics:
  pulumi:pulumi:Stack (foyle-dev-dev):
    warning: unable to detect a global setting for GCP Project;
    Pulumi will rely on per-resource settings for this operation.
    Set the GCP Project by using:
    	`pulumi config set gcp:project <project>`

Resources:
    8 unchanged

Updating (dev)

View Live: https://app.pulumi.com/jlewi/foyle-dev/dev/updates/7


@ Updating.....
    pulumi:pulumi:Stack foyle-dev-dev running 
@ Updating....
    pulumi:pulumi:Stack foyle-dev-dev running warning: unable to detect a global setting for GCP Project;
    gcp:organizations:Project project  
    gcp:projects:Service storage.googleapis.com  
    gcp:projects:Service secretmanager.googleapis.com  
    gcp:projects:Service artifactregistry.googleapis.com  
    gcp:projects:Service container.googleapis.com  
    gcp:projects:Service cloudbuild.googleapis.com  
    gcp:container:Cluster dev  
    pulumi:pulumi:Stack foyle-dev-dev  1 warning
Diagnostics:
  pulumi:pulumi:Stack (foyle-dev-dev):
    warning: unable to detect a global setting for GCP Project;
    Pulumi will rely on per-resource settings for this operation.
    Set the GCP Project by using:
    	`pulumi config set gcp:project <project>`

Resources:
    8 unchanged

Duration: 4s

```
