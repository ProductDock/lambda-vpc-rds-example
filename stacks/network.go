package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

const CIDR = "10.1.0.0/16"

type NetworkStackExport struct {
	Stack                       awscdk.Stack
	Vpc                         awsec2.Vpc
	SecretsManagerVpcEndpointSg awsec2.SecurityGroup
	LambdaVpcEndpointSg         awsec2.SecurityGroup
}

func Network(scope constructs.Construct, id string, props awscdk.StackProps) NetworkStackExport {
	stack := awscdk.NewStack(scope, &id, &props)

	vpc := awsec2.NewVpc(stack, jsii.String("vpc"), &awsec2.VpcProps{
		IpAddresses: awsec2.IpAddresses_Cidr(jsii.String(CIDR)),
		MaxAzs:      jsii.Number(2),
		NatGateways: jsii.Number(0),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:       jsii.String("private-subnet-isolated"),
				SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
				CidrMask:   jsii.Number(26),
			},
			{
				Name:       jsii.String("public-subnet"),
				SubnetType: awsec2.SubnetType_PUBLIC,
				CidrMask:   jsii.Number(26),
			},
		},
	})

	secretsManagerVpcEndpointSg := awsec2.NewSecurityGroup(stack, jsii.String("endpoint-sg"), &awsec2.SecurityGroupProps{
		Vpc:               vpc,
		AllowAllOutbound:  jsii.Bool(false),
		SecurityGroupName: jsii.String("endpoint-sg"),
	})

	lambdaVpcEndpointSg := awsec2.NewSecurityGroup(stack, jsii.String("lambda-sg"), &awsec2.SecurityGroupProps{
		Vpc:               vpc,
		AllowAllOutbound:  jsii.Bool(false),
		SecurityGroupName: jsii.String("lambda-sg"),
	})

	// https://repost.aws/knowledge-center/lambda-secret-vpc

	vpc.AddInterfaceEndpoint(jsii.String("secret-manager-endpoint"), &awsec2.InterfaceVpcEndpointOptions{
		Service:           awsec2.InterfaceVpcEndpointAwsService_SECRETS_MANAGER(),
		PrivateDnsEnabled: jsii.Bool(true),
		Open:              jsii.Bool(false),
		SecurityGroups:    &[]awsec2.ISecurityGroup{secretsManagerVpcEndpointSg},
	})

	return NetworkStackExport{
		Stack:                       stack,
		Vpc:                         vpc,
		SecretsManagerVpcEndpointSg: secretsManagerVpcEndpointSg,
		LambdaVpcEndpointSg:         lambdaVpcEndpointSg,
	}
}
