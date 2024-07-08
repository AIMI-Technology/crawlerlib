package database

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var client *dynamodb.Client
var tableName string = "ScrapedData"

type ScrapedData = map[string]types.AttributeValue

func init() {
	ctx := context.Background()

	config, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("unable to load SDK config " + err.Error())
	}

	config.Region = "us-east-1"

	client = dynamodb.NewFromConfig(config)
}

func Client() *dynamodb.Client {
	return client
}

func PutItem(ctx context.Context, item map[string]types.AttributeValue) error {
	_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &tableName,
		Item:      item,
	})

	return err
}

func ScanForEmptyPrimaryCategory(ctx context.Context) (*dynamodb.ScanOutput, error) {
	expression := "attribute_not_exists(PrimaryCategory) or PrimaryCategory = :val"
	params := map[string]types.AttributeValue{
		":val": &types.AttributeValueMemberS{Value: ""},
	}

	scanInput := &dynamodb.ScanInput{
		TableName:                 &tableName,
		FilterExpression:          &expression,
		ExpressionAttributeValues: params,
	}

	result, err := client.Scan(ctx, scanInput)
	if err != nil {
		return nil, fmt.Errorf("failed to scan for empty PrimaryCategory: %w", err)
	}

	return result, nil
}

func UpdateItem(ctx context.Context, itemId string, attributesToUpdate map[string]interface{}) error {
	// Construct the update expression parts
	var updateExpression string
	var expressionAttributeNames map[string]string = make(map[string]string)
	var expressionAttributeValues map[string]types.AttributeValue = make(map[string]types.AttributeValue)

	i := 0
	for attrName, attrValue := range attributesToUpdate {
		placeholderName := fmt.Sprintf("#N%d", i)
		placeholderValue := fmt.Sprintf(":v%d", i)
		if i == 0 {
			updateExpression = "SET " + placeholderName + " = " + placeholderValue
		} else {
			updateExpression += ", " + placeholderName + " = " + placeholderValue
		}
		expressionAttributeNames[placeholderName] = attrName
		expressionAttributeValues[placeholderValue] = &types.AttributeValueMemberS{Value: fmt.Sprint(attrValue)}
		i++
	}

	// Prepare the UpdateItem input
	input := &dynamodb.UpdateItemInput{
		TableName:                 &tableName,
		Key:                       map[string]types.AttributeValue{"Id": &types.AttributeValueMemberS{Value: itemId}},
		UpdateExpression:          &updateExpression,
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
		ReturnValues:              types.ReturnValueUpdatedNew,
	}

	// Execute the UpdateItem operation
	_, err := client.UpdateItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	return nil
}
