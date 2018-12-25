# TODO-API
## Made on golang
Run locally:
- Create and setup .env file like this:
  - `PORT=3000`
  - `AWS_REGION=us-east-1`
  - `DYNAMODB_ENDPOINT='http://localhost:8000'`
  - `TABLE_NAME=Tasks`
- Run dynamodb docker image: `docker run -p 8000:8000 amazon/dynamodb-local`
- Build and run: `go run main.go (--port 4001 ` [optional]`)`
