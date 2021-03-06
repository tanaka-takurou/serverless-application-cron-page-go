package main

import (
	"os"
	"fmt"
	"log"
	"time"
	"context"
	"strconv"
	"net/http"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchevents"
	slambda "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

type APIResponse struct {
	Last     string `json:"last"`
	Message  string `json:"message"`
	Schedule string `json:"schedule"`
}

type Response events.APIGatewayProxyResponse

var cfg aws.Config
var lambdaClient *slambda.Client
var cloudwatcheventsClient *cloudwatchevents.Client

const layout  string = "2006-01-02 15:04"

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	var jsonBytes []byte
	var err error
	if &request.RequestContext != nil && &request.RequestContext.Identity != nil && len (request.RequestContext.Identity.SourceIP) > 0 {
		log.Println(request.RequestContext.Identity.SourceIP)
		d := make(map[string]string)
		json.Unmarshal([]byte(request.Body), &d)
		if v, ok := d["action"]; ok {
			switch v {
			case "describe" :
				schedule, e := describeRule(ctx)
				if e != nil {
					err = e
				} else {
					environment, e_ := getLambdaEnvironment(ctx)
					if e_ != nil {
						err = e_
					} else {
						jsonBytes, _ = json.Marshal(APIResponse{Message: "Success", Last: environment["LAST_EVENT"], Schedule: schedule})
					}
				}
			case "put" :
				if minute, ok := d["minute"]; ok {
					if hour, o2 := d["hour"]; o2 {
						if day, o3 := d["day"]; o3 {
							if month, o4 := d["month"]; o4 {
								if year, o5 := d["year"]; o5 {
									e := putRule(ctx, minute, hour, day, month, year)
									if e != nil {
										err = e
									} else {
										jsonBytes, _ = json.Marshal(APIResponse{Message: "Success", Last: "", Schedule: ""})
									}
								}
							}
						}
					}
				}
			}
		}
		if err != nil {
			log.Print(err)
			jsonBytes, _ = json.Marshal(APIResponse{Message: fmt.Sprint(err), Last: "", Schedule: ""})
			return Response{
				StatusCode: http.StatusInternalServerError,
				Body: string(jsonBytes),
			}, nil
		}
	} else {
		err := updateLambdaEnvironment(ctx)
		if err != nil {
			log.Print(err)
			jsonBytes, _ = json.Marshal(APIResponse{Message: fmt.Sprint(err), Last: "", Schedule: ""})
			return Response{
				StatusCode: http.StatusInternalServerError,
				Body: string(jsonBytes),
			}, nil
		} else {
			jsonBytes, _ = json.Marshal(APIResponse{Message: "Success", Last: "", Schedule: ""})
		}
	}
	return Response {
		StatusCode: http.StatusOK,
		Body: string(jsonBytes),
	}, nil
}

func describeRule(ctx context.Context)(string, error) {
	if cloudwatcheventsClient == nil {
		cloudwatcheventsClient = getCloudwatcheventsClient(ctx)
	}
	params := &cloudwatchevents.DescribeRuleInput{
		Name: aws.String(os.Getenv("EVENT_NAME")),
	}
	res, err := cloudwatcheventsClient.DescribeRule(ctx, params)
	if err != nil {
		log.Print(err)
		return "", err
	}
	return aws.ToString(res.ScheduleExpression), nil
}

func putRule(ctx context.Context, minute string, hour string, day string, month string, year string) error {
	var m_ int
	var h_ int
	var d_ int
	var o_ int
	var y_ int
	m_, _ = strconv.Atoi(minute)
	h_, _ = strconv.Atoi(hour)
	d_, _ = strconv.Atoi(day)
	o_, _ = strconv.Atoi(month)
	y_, _ = strconv.Atoi(year)
	if m_ < 0 {
		m_ = 0
	}
	sm := strconv.Itoa(m_)
	if h_ < 0 {
		h_ = 0
	}
	sh := strconv.Itoa(h_)
	sd := "*"
	if d_ > 0 {
		sd = strconv.Itoa(d_)
	}
	so := "*"
	if o_ > 0 {
		so = strconv.Itoa(o_)
	}
	sy := "*"
	if y_ >= 1970 {
		sy = strconv.Itoa(y_)
	}
	if cloudwatcheventsClient == nil {
		cloudwatcheventsClient = getCloudwatcheventsClient(ctx)
	}
	params := &cloudwatchevents.PutRuleInput{
		Name: aws.String(os.Getenv("EVENT_NAME")),
		ScheduleExpression: aws.String("cron(" + sm + " " + sh + " " + sd + " " + so + " ? " + sy + ")"),
	}
	res, err := cloudwatcheventsClient.PutRule(ctx, params)
	if err != nil {
		log.Print(err)
		return err
	}
	log.Printf("%+v\n", res)
	return nil
}

func getLambdaEnvironment(ctx context.Context)(map[string]string, error) {
	if lambdaClient == nil {
		lambdaClient = getLambdaClient(ctx)
	}
	res, err := lambdaClient.GetFunctionConfiguration(ctx, &slambda.GetFunctionConfigurationInput{
		FunctionName: aws.String(os.Getenv("FUNCTION_NAME")),
	})
	if err != nil {
		log.Println(err)
		return map[string]string{}, err
	}
	return res.Environment.Variables, nil
}

func updateLambdaEnvironment(ctx context.Context) error {
	t := time.Now()
	env, err := getLambdaEnvironment(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	env["LAST_EVENT"] = t.Format(layout)
	_, err = lambdaClient.UpdateFunctionConfiguration(ctx, &slambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(os.Getenv("FUNCTION_NAME")),
		Environment: &types.Environment{
			Variables: env,
		},
	})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func getCloudwatcheventsClient(ctx context.Context) *cloudwatchevents.Client {
	if cfg.Region != os.Getenv("REGION") {
		cfg = getConfig(ctx)
	}
	return cloudwatchevents.NewFromConfig(cfg)
}

func getLambdaClient(ctx context.Context) *slambda.Client {
	if cfg.Region != os.Getenv("REGION") {
		cfg = getConfig(ctx)
	}
	return slambda.NewFromConfig(cfg)
}

func getConfig(ctx context.Context) aws.Config {
	var err error
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("REGION")))
	if err != nil {
		log.Print(err)
	}
	return cfg
}

func main() {
	lambda.Start(HandleRequest)
}
