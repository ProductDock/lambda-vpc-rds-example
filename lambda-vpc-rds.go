package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"lambda-vpc-rds/stacks"

	"github.com/aws/jsii-runtime-go"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	network := stacks.Network(app, "network-stack", awscdk.StackProps{
		Env: env(),
	})

	storage := stacks.Storage(app, "storage-stack", &stacks.StorageStackProperties{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
		NetworkStackData: network,
	})

	stacks.Application(app, "application-stack", &stacks.ApplicationStackProperties{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
		NetworkStackData: network,
		StorageStackData: storage,
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
