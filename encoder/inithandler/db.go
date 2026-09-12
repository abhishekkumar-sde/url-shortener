package inithandler

import (
	"context"
	"fmt"
	"time"
	"url-shortener/encoder/svcparam/envvar"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	sdkdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/spf13/viper"
)

type Client struct {
	Client *sdkdynamodb.Client
	Table  string
}

func GetDynamoDb(ctx context.Context) (*Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(viper.GetString(envvar.DynamoDBRegion)),
	)
	if err != nil {
		return nil, err
	}

	dbEndpoint := fmt.Sprintf("%s:%s", viper.GetString(envvar.DynamoDBHost), viper.GetString(envvar.DynamoDBPort))
	client := sdkdynamodb.NewFromConfig(awsCfg, func(o *sdkdynamodb.Options) {
		o.BaseEndpoint = aws.String("http://" + dbEndpoint)
	})

	return &Client{
		Client: client,
		Table:  viper.GetString(envvar.DynamoDBTable),
	}, nil
}

func (c *Client) EnsureTable(ctx context.Context) error {
	var lastErr error

	// DynamoDB Local can take a few seconds to become ready.
	// Retry the connection before giving up.
	for attempt := 1; attempt <= 10; attempt++ {
		_, err := c.Client.DescribeTable(
			ctx,
			&sdkdynamodb.DescribeTableInput{
				TableName: aws.String(c.Table),
			},
		)

		if err == nil {
			return nil
		}

		lastErr = err

		// Wait before retrying.
		if attempt < 10 {
			time.Sleep(3 * time.Second)
		}
	}

	// If DescribeTable failed, the table probably doesn't exist.
	// Create it.
	_, err := c.Client.CreateTable(
		ctx,
		&sdkdynamodb.CreateTableInput{
			TableName: aws.String(c.Table),

			AttributeDefinitions: []types.AttributeDefinition{
				{
					AttributeName: aws.String("code"),
					AttributeType: types.ScalarAttributeTypeS,
				},
			},

			KeySchema: []types.KeySchemaElement{
				{
					AttributeName: aws.String("code"),
					KeyType:       types.KeyTypeHash,
				},
			},

			BillingMode: types.BillingModePayPerRequest,
		},
	)

	if err != nil {
		return fmt.Errorf(
			"create table %q: %w; last describe error: %v",
			c.Table,
			err,
			lastErr,
		)
	}

	// Wait until DynamoDB reports that the table exists.
	waiter := sdkdynamodb.NewTableExistsWaiter(c.Client)

	return waiter.Wait(
		ctx,
		&sdkdynamodb.DescribeTableInput{
			TableName: aws.String(c.Table),
		},
		30*time.Second,
	)
}
