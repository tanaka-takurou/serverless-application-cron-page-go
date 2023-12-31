# serverless-application-cron kit
Simple kit for serverless application cron page using AWS Lambda.


## Dependence
- aws-lambda-go
- aws-sdk-go-v2


## Requirements
- AWS (Lambda, API Gateway)
- aws-sam-cli
- golang environment


## Usage

### Deploy
```bash
make clean build
AWS_PROFILE={profile} AWS_DEFAULT_REGION={region} make bucket={bucket} stack={stack name} deploy
```
