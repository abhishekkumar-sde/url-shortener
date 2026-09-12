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

func (r *URLRepository) Get(ctx context.Context, code string) (model.URL, error) {
	result, err := r.client.GetItem(ctx, &sdkdynamodb.GetItemInput{
		TableName: aws.String(r.table),
		Key: map[string]types.AttributeValue{
			"code": &types.AttributeValueMemberS{Value: code},
		},
	})
	if err != nil {
		return model.URL{}, fmt.Errorf("get URL: %w", err)
	}

	if len(result.Item) == 0 {
		return model.URL{}, svcerror.ErrNotFound
	}

	var u model.URL
	if err := attributevalue.UnmarshalMap(result.Item, &u); err != nil {
		return model.URL{}, fmt.Errorf("unmarshal URL: %w", err)
	}

	return u, nil
}

func (r *URLRepository) GetByLongURL(ctx context.Context, longURL string) (model.URL, error) {
	result, err := r.client.Query(ctx, &sdkdynamodb.QueryInput{
		TableName:              aws.String(r.table),
		IndexName:              aws.String("long_url-index"),
		KeyConditionExpression: aws.String("long_url = :long_url"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":long_url": &types.AttributeValueMemberS{
				Value: longURL,
			},
		},
		Limit: aws.Int32(1),
	})

	if err != nil {
		return model.URL{}, fmt.Errorf("get URL by long URL: %w", err)
	}

	if len(result.Items) == 0 {
		return model.URL{}, svcerror.ErrNotFound
	}

	var u model.URL

	if err := attributevalue.UnmarshalMap(result.Items[0], &u); err != nil {
		return model.URL{}, fmt.Errorf("unmarshal URL: %w", err)
	}

	return u, nil
}
