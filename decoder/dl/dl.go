package dl

import (
	"context"
	"fmt"

	"url-shortener/decoder/model"
	"url-shortener/decoder/svcerror"

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
