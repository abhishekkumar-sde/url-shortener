package dl

import (
	"context"
	"errors"
	"fmt"

	"url-shortener/encoder/model"
	"url-shortener/encoder/svcerror"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	sdkdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type URLRepository struct {
	client *sdkdynamodb.Client
	table  string
}

func NewURLRepository(client *sdkdynamodb.Client, table string) *URLRepository {
	return &URLRepository{client: client, table: table}
}

func (r *URLRepository) Create(ctx context.Context, u model.URL) error {
	item, err := attributevalue.MarshalMap(u)
	if err != nil {
		return fmt.Errorf("marshal URL: %w", err)
	}

	_, err = r.client.PutItem(ctx, &sdkdynamodb.PutItemInput{
		TableName:           aws.String(r.table),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(code)"),
	})
	if err != nil {
		var conditionalErr *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalErr) {
			return svcerror.ErrConflict
		}
		return fmt.Errorf("create URL: %w", err)
	}

	return nil
}
