package db

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

//TodoItem schema
type TodoItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Timestamp   string `json:"timestamp"`
}

//TodoItemUpdate is the struct for updateExpression in mongodb
type TodoItemUpdate struct {
	Name        string `json:":n"`
	Description string `json:":d"`
	Timestamp   string `json:":t"`
}

//Svc is the new a dynamodb connection
type Svc struct {
	Db        *dynamodb.DynamoDB
	TableName string
}

//ConnectAndCreateTable connects to dynamodb and creates table if it does not exists yet
//noinspection ALL
func (s *Svc) ConnectAndCreateTable(envs map[string]string) {
	config := &aws.Config{
		Region:   aws.String(envs["AWS_REGION"]),
		Endpoint: aws.String(envs["DYNAMODB_ENDPOINT"]),
	}
	sess := session.Must(session.NewSession(config))
	s.Db = dynamodb.New(sess)
	s.TableName = envs["TABLE_NAME"]

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String(s.TableName),
	}
	//create new table
	if result, err := s.Db.CreateTable(input); err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(result)
	}
}

//CreateDbItem creates new item in dynamodb
func (s *Svc) CreateDbItem(item *TodoItem) error {
	info, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      info,
		TableName: aws.String(s.TableName),
	}

	_, err = s.Db.PutItem(input)
	return err
}

//GetDbItem gets item from Db by passing an "id" attribute
//noinspection ALL
func (s *Svc) GetDbItem(id string) (TodoItem, error) {
	result, err := s.Db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(s.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
	})
	item := TodoItem{}
	if err != nil {
		return item, err
	}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return item, err
	}
	return item, nil
}

//UpdateDbItem update item in Db, if he is exists
func (s *Svc) UpdateDbItem(id string, item *TodoItem) error {
	itemMap, err := dynamodbattribute.MarshalMap(TodoItemUpdate{
		item.Name,
		item.Description,
		item.Timestamp,
	})
	if err != nil {
		return err
	}
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(s.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
		ExpressionAttributeValues: itemMap,
		ExpressionAttributeNames: map[string]*string{
			"#n": aws.String("name"),
			"#d": aws.String("description"),
			"#t": aws.String("timestamp"),
		},
		UpdateExpression: aws.String("SET #n = :n, #d = :d, #t = :t"),
		ReturnValues:     aws.String("ALL_NEW"),
	}

	result, err := s.Db.UpdateItem(input)
	if err != nil {
		return err
	}
	err = dynamodbattribute.UnmarshalMap(result.Attributes, &item)
	if err != nil {
		return err
	}
	return nil
}

//DeleteDbItem deleting item from Db by passing an "id" attribute
func (s *Svc) DeleteDbItem(id string) error {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
		TableName: aws.String(s.TableName),
	}
	_, err := s.Db.DeleteItem(input)
	return err
}

//GetAllDbItems gets all existing items from dynamodb
func (s *Svc) GetAllDbItems() ([]TodoItem, error) {
	proj := expression.NamesList(
		expression.Name("id"),
		expression.Name("description"),
		expression.Name("name"),
		expression.Name("timestamp"),
	)

	expr, err := expression.NewBuilder().WithProjection(proj).Build()
	if err != nil {
		return nil, err
	}
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(s.TableName),
	}
	result, err := s.Db.Scan(params)
	if err != nil {
		return nil, err
	}
	var itemsArr []TodoItem
	for _, i := range result.Items {
		item := TodoItem{}
		err = dynamodbattribute.UnmarshalMap(i, &item)
		if err != nil {
			return nil, err
		}
		itemsArr = append(itemsArr, item)
	}
	return itemsArr, nil
}
