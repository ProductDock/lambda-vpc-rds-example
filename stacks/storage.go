package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type StorageStackProperties struct {
	StackProps       awscdk.StackProps
	NetworkStackData NetworkStackExport
}

type StorageStackExport struct {
	Stack                 awscdk.Stack
	TutorialDB            awsrds.DatabaseInstance
	TutorialDBCredentials awsrds.Credentials
}

func Storage(scope constructs.Construct, id string, props *StorageStackProperties) StorageStackExport {
	stack := awscdk.NewStack(scope, &id, &props.StackProps)

	stack.AddDependency(props.NetworkStackData.Stack, jsii.String("Need Network stack to be created"))

	sg := createDBSecurityGroup(stack, props.NetworkStackData.Vpc)

	cred := awsrds.Credentials_FromGeneratedSecret(jsii.String("postgres"), &awsrds.CredentialsBaseOptions{
		SecretName:        jsii.String("postgres-secret"),
		ExcludeCharacters: jsii.String("\"@/\\:"),
	})

	engine := awsrds.DatabaseInstanceEngine_Postgres(&awsrds.PostgresInstanceEngineProps{
		Version: awsrds.PostgresEngineVersion_VER_14_6(),
	})

	db := awsrds.NewDatabaseInstance(stack, jsii.String("rds"), &awsrds.DatabaseInstanceProps{
		Port:                    jsii.Number(5432),
		Engine:                  engine,
		StorageEncrypted:        jsii.Bool(true),
		MultiAz:                 jsii.Bool(false),
		AutoMinorVersionUpgrade: jsii.Bool(false),
		AllocatedStorage:        jsii.Number(25),
		StorageType:             awsrds.StorageType_GP2,
		BackupRetention:         awscdk.Duration_Days(jsii.Number(5)),
		DeletionProtection:      jsii.Bool(false),
		DatabaseName:            jsii.String("tutorial"),
		InstanceType:            awsec2.InstanceType_Of(awsec2.InstanceClass_BURSTABLE3, awsec2.InstanceSize_SMALL),
		Credentials:             cred,
		Vpc:                     props.NetworkStackData.Vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
		},
		SecurityGroups: &[]awsec2.ISecurityGroup{sg},
		RemovalPolicy:  awscdk.RemovalPolicy_DESTROY,
	})

	return StorageStackExport{Stack: stack, TutorialDB: db, TutorialDBCredentials: cred}
}

func createDBSecurityGroup(stack awscdk.Stack, vpc awsec2.Vpc) awsec2.SecurityGroup {
	sg := awsec2.NewSecurityGroup(stack, jsii.String("rds-sg"), &awsec2.SecurityGroupProps{
		Vpc:               vpc,
		AllowAllOutbound:  jsii.Bool(false),
		SecurityGroupName: jsii.String("rds-sg"),
	})
	sg.AddIngressRule(
		awsec2.Peer_Ipv4(jsii.String(CIDR)),
		awsec2.Port_Tcp(jsii.Number(5432)),
		jsii.String("Allow connections to the database."),
		jsii.Bool(false),
	)

	return sg
}
