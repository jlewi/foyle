package main

import (
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/container"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	project = "foyle-dev"
	// TODO(jeremy): How to share these constants with the infra project?
	org = "lewi.us"
	// orgId for lewi.us
	orgId            = "779939238227"
	billingAccountId = "011DA8-DD6D07-CA9324"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create a new GCP project named foyle-dev
		p, err := organizations.NewProject(ctx, "project", &organizations.ProjectArgs{
			Name:           pulumi.String(project),
			ProjectId:      pulumi.String(project),
			OrgId:          pulumi.String(orgId),
			BillingAccount: pulumi.String(billingAccountId),
		})
		if err != nil {
			return err
		}

		services := []string{
			"artifactregistry.googleapis.com",
			"cloudbuild.googleapis.com",
			"container.googleapis.com",
			"sheets.googleapis.com",
			"storage.googleapis.com",
			"secretmanager.googleapis.com",
		}

		svcs := make([]pulumi.Resource, 0, len(services))

		for _, s := range services {
			svc, err := projects.NewService(ctx, s, &projects.ServiceArgs{
				Project:                  p.Name,
				Service:                  pulumi.String(s),
				DisableDependentServices: pulumi.Bool(true),
			})

			if err != nil {
				return err
			}
			svcs = append(svcs, svc)
		}

		// Deploy a GKE autopilot cluster in us-west1 region
		// TODO(jeremy): We should really switch to a private cluster. That requires setting up a nat which is annoying.
		_, err = container.NewCluster(ctx, "dev", &container.ClusterArgs{
			Name:               pulumi.String("dev"),
			Location:           pulumi.String("us-west1"),
			EnableAutopilot:    pulumi.Bool(true),
			Project:            p.ProjectId,
			DeletionProtection: pulumi.Bool(false),
		}, pulumi.DependsOn(svcs))
		if err != nil {
			return err
		}
		account, err := serviceaccount.NewAccount(ctx, "developer", &serviceaccount.AccountArgs{
			AccountId:   pulumi.String("developer"),
			DisplayName: pulumi.String("Developer Service Account"),
			Project:     p.ProjectId,
		})
		if err != nil {
			return err
		}

		// Create IAM binding for the service account
		_, err = serviceaccount.NewIAMBinding(ctx, "serviceAccountIAMBinding", &serviceaccount.IAMBindingArgs{
			ServiceAccountId: account.Name,
			Role:             pulumi.String("roles/iam.serviceAccountTokenCreator"),
			Members:          pulumi.StringArray{pulumi.String("user:jeremy@lewi.us")},
		})
		if err != nil {
			return err
		}

		return nil
	})
}
