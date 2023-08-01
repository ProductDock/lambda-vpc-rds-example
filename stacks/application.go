package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type ApplicationStackProperties struct {
	StackProps       awscdk.StackProps
	NetworkStackData NetworkStackExport
	StorageStackData StorageStackExport
}

func Application(scope constructs.Construct, id string, props *ApplicationStackProperties) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, &props.StackProps)
	stack.AddDependency(props.NetworkStackData.Stack, jsii.String("Need Network stack to be created"))
	stack.AddDependency(props.StorageStackData.Stack, jsii.String("Need Storage stack to be created"))

	role := createLambdaRole(stack)

	lambdaToRdsSg := createLambdaToRDSSecurityGroup(stack, props.NetworkStackData.Vpc)

	lambdaVpcEndpointSg := props.NetworkStackData.LambdaVpcEndpointSg
	secretsManagerVpcEndpointSg := props.NetworkStackData.SecretsManagerVpcEndpointSg
	allowLambdaAccessToSecretsManager(secretsManagerVpcEndpointSg, lambdaVpcEndpointSg)

	lambda := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("ping-db-lambda"), &awscdklambdagoalpha.GoFunctionProps{
		Runtime:      awslambda.Runtime_PROVIDED_AL2(),
		Architecture: awslambda.Architecture_ARM_64(),
		Entry:        jsii.String("./lambdas/pingdb"),
		Bundling: &awscdklambdagoalpha.BundlingOptions{
			GoBuildFlags: jsii.Strings(`-ldflags "-s -w"`),
		},
		Environment: &map[string]*string{
			"DB_HOST":     props.StorageStackData.TutorialDB.DbInstanceEndpointAddress(),
			"DB_PORT":     props.StorageStackData.TutorialDB.DbInstanceEndpointPort(),
			"DB_USERNAME": jsii.String("postgres"),
		},
		Role:    role,
		Timeout: awscdk.Duration_Seconds(aws.Float64(5)),
		Vpc:     props.NetworkStackData.Vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
		},
		SecurityGroups: &[]awsec2.ISecurityGroup{lambdaVpcEndpointSg, lambdaToRdsSg},
	})

	exposeLambdaEndpoint(stack, lambda)

	return stack
}

func createLambdaRole(stack awscdk.Stack) awsiam.Role {
	return awsiam.NewRole(stack, jsii.String("pingdb-lambda-role"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), &awsiam.ServicePrincipalOpts{}),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromManagedPolicyArn(stack, jsii.String("AWSLambdaBasicExecutionRole"), jsii.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole")),
			awsiam.ManagedPolicy_FromManagedPolicyArn(stack, jsii.String("AWSLambdaVPCAccessExecutionRole"), jsii.String("arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole")),
			awsiam.ManagedPolicy_FromManagedPolicyArn(stack, jsii.String("SecretsManagerReadWrite"), jsii.String("arn:aws:iam::aws:policy/SecretsManagerReadWrite")),
		},
	})
}

func allowLambdaAccessToSecretsManager(secretsManagerVpcEndpointSg awsec2.SecurityGroup, lambdaVpcEndpointSg awsec2.SecurityGroup) {

	secretsManagerVpcEndpointSg.AddIngressRule(
		lambdaVpcEndpointSg,
		awsec2.Port_Tcp(jsii.Number(443)),
		jsii.String("Allow connection from lambda security group."),
		jsii.Bool(false))

	lambdaVpcEndpointSg.AddEgressRule(
		secretsManagerVpcEndpointSg,
		awsec2.Port_Tcp(jsii.Number(443)),
		jsii.String("Allow outbound traffic to vpc endpoint."),
		jsii.Bool(false))
}
func createLambdaToRDSSecurityGroup(stack awscdk.Stack, vpc awsec2.Vpc) awsec2.SecurityGroup {
	lambdaToRdsSg := awsec2.NewSecurityGroup(stack, jsii.String("lambda-rds-sg"), &awsec2.SecurityGroupProps{
		Vpc:               vpc,
		AllowAllOutbound:  jsii.Bool(false),
		SecurityGroupName: jsii.String("lambda-rds-sg"),
	})

	lambdaToRdsSg.AddEgressRule(
		awsec2.Peer_Ipv4(jsii.String(CIDR)),
		awsec2.Port_Tcp(jsii.Number(5432)),
		jsii.String("Allow connection to the database."),
		jsii.Bool(false))
	return lambdaToRdsSg
}

func exposeLambdaEndpoint(stack awscdk.Stack, function awslambda.Function) {
	lambdaUrl := awslambda.NewFunctionUrl(stack, jsii.String("pingdb-function-url"), &awslambda.FunctionUrlProps{
		Function: function,
		AuthType: awslambda.FunctionUrlAuthType_NONE,
	})
	lambdaUrl.GrantInvokeUrl(awsiam.NewAnyPrincipal())

	awscdk.NewCfnOutput(stack, jsii.String("pingdb-function-url-output"), &awscdk.CfnOutputProps{
		ExportName:  jsii.String("pingdb-function-url"),
		Value:       lambdaUrl.Url(),
		Description: jsii.String("PingDB Function Url"),
	})
}
