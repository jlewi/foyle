package main

import (
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/artifactregistry"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	project = "foyle-public"
	org     = "lewi.us"
	// orgId for lewi.us
	orgId            = "779939238227"
	billingAccountId = "011DA8-DD6D07-CA9324"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
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
			"storage.googleapis.com",
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

		if err != nil {
			return err
		}
		// Create an artifact registry to store Docker images
		_, err = artifactregistry.NewRepository(ctx, "images", &artifactregistry.RepositoryArgs{
			Description:  pulumi.String("Artifact Registry for images related to foyle and hydros"),
			Location:     pulumi.String("us-west1"),
			Format:       pulumi.String("DOCKER"),
			Project:      pulumi.String(project),
			RepositoryId: pulumi.String("images"),
			// TODO(jeremy): Could we make the depends on for the services more scoped so we only depend on the
			// artifactregistry service.
		}, pulumi.DependsOn(svcs))

		if err != nil {
			return err
		}

		// Create the bucket for the google cloud build

		bucket, err := storage.NewBucket(ctx, "builds-foyle-public", &storage.BucketArgs{
			Name:     pulumi.String("builds-foyle-public"),
			Location: pulumi.String("US"),
			Project:  pulumi.String(project),
			// Artifacts should be GC'd after 14 days
			LifecycleRules: storage.BucketLifecycleRuleArray{
				&storage.BucketLifecycleRuleArgs{
					Action: storage.BucketLifecycleRuleActionArgs{
						Type: pulumi.String("Delete"),
					},
					Condition: storage.BucketLifecycleRuleConditionArgs{
						Age:       pulumi.Int(14),
						WithState: pulumi.String("ANY"),
					},
				},
			},
			UniformBucketLevelAccess: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		// Export the DNS name of the bucket
		// N.B this is mostly here as an example of using ctx.Export. I think it can be used to pass along
		// resource values using the context
		ctx.Export("buildsBucketName", bucket.Url)
		return nil

	})
}
