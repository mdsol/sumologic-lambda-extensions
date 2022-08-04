package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/SumoLogic/sumologic-lambda-extensions/lambda-extensions/utils"

	"github.com/sirupsen/logrus"
)

// LambdaExtensionConfig config for storing all configurable parameters
type LambdaExtensionConfig struct {
	SumoHTTPEndpoint       string
	EnableFailover         bool
	S3BucketName           string
	S3BucketRegion         string
	NumRetry               int
	AWSLambdaRuntimeAPI    string
	LogTypes               []string
	FunctionName           string
	FunctionVersion        string
	LogLevel               logrus.Level
	MaxDataQueueLength     int
	MaxConcurrentRequests  int
	MaxRetryAttempts       int
	RetrySleepTime         time.Duration
	ConnectionTimeoutValue time.Duration
	MaxDataPayloadSize     int
	LambdaRegion           string
	SourceCategoryOverride string
}

var defaultLogTypes = []string{"platform", "function"}
var validLogTypes = []string{"platform", "function", "extension"}

// GetConfig to get config instance
func GetConfig() (*LambdaExtensionConfig, error) {
	sumoHttpEndpoint, ret := os.LookupEnv("SUMO_HTTP_ENDPOINT")
	if ret == false {
		sumoHttpEndpoint = "<REPLACE ME>"
	}

	config := &LambdaExtensionConfig{
		SumoHTTPEndpoint:       sumoHttpEndpoint,
		S3BucketName:           os.Getenv("SUMO_S3_BUCKET_NAME"),
		S3BucketRegion:         os.Getenv("SUMO_S3_BUCKET_REGION"),
		AWSLambdaRuntimeAPI:    os.Getenv("AWS_LAMBDA_RUNTIME_API"),
		FunctionName:           os.Getenv("AWS_LAMBDA_FUNCTION_NAME"),
		FunctionVersion:        os.Getenv("AWS_LAMBDA_FUNCTION_VERSION"),
		LambdaRegion:           os.Getenv("AWS_REGION"),
		SourceCategoryOverride: os.Getenv("SOURCE_CATEGORY_OVERRIDE"),
		MaxRetryAttempts:       5,
		ConnectionTimeoutValue: 10000 * time.Millisecond,
		MaxDataPayloadSize:     1024 * 1024, // 1 MB
	}

	(*config).setDefaults()

	err := (*config).validateConfig()

	if err != nil {
		return config, err
	}
	return config, nil
}
func (cfg *LambdaExtensionConfig) setDefaults() {
	numRetry := os.Getenv("SUMO_NUM_RETRIES")
	retrySleepTime := os.Getenv("SUMO_RETRY_SLEEP_TIME_MS")
	logLevel := os.Getenv("SUMO_LOG_LEVEL")
	maxDataQueueLength := os.Getenv("SUMO_MAX_DATAQUEUE_LENGTH")
	maxConcurrentRequests := os.Getenv("SUMO_MAX_CONCURRENT_REQUESTS")
	enableFailover := os.Getenv("SUMO_ENABLE_FAILOVER")
	logTypes := os.Getenv("SUMO_LOG_TYPES")

	if numRetry == "" {
		cfg.NumRetry = 3
	}
	if logLevel == "" {
		cfg.LogLevel = logrus.InfoLevel
	}
	if maxDataQueueLength == "" {
		cfg.MaxDataQueueLength = 20
	}
	if maxConcurrentRequests == "" {
		cfg.MaxConcurrentRequests = 3
	}

	if enableFailover == "" {
		cfg.EnableFailover = false
	}
	if cfg.AWSLambdaRuntimeAPI == "" {
		cfg.AWSLambdaRuntimeAPI = "127.0.0.1:9001"
	}
	if logTypes == "" {
		cfg.LogTypes = defaultLogTypes
	} else {
		cfg.LogTypes = strings.Split(logTypes, ",")
	}
	if retrySleepTime == "" {
		cfg.RetrySleepTime =  300 * time.Millisecond
	}

}

func (cfg *LambdaExtensionConfig) validateConfig() error {
	numRetry := os.Getenv("SUMO_NUM_RETRIES")
	logLevel := os.Getenv("SUMO_LOG_LEVEL")
	maxDataQueueLength := os.Getenv("SUMO_MAX_DATAQUEUE_LENGTH")
	maxConcurrentRequests := os.Getenv("SUMO_MAX_CONCURRENT_REQUESTS")
	enableFailover := os.Getenv("SUMO_ENABLE_FAILOVER")
	retrySleepTime := os.Getenv("SUMO_RETRY_SLEEP_TIME_MS")

	var allErrors []string
	var err error

	if cfg.SumoHTTPEndpoint == "" {
		allErrors = append(allErrors, "SUMO_HTTP_ENDPOINT not set in environment variable")
	}

	// Todo test url valid
	if cfg.SumoHTTPEndpoint != "" {
		_, err = url.ParseRequestURI(cfg.SumoHTTPEndpoint)
		if err != nil {
			allErrors = append(allErrors, "SUMO_HTTP_ENDPOINT is not Valid")
		}
	}

	if enableFailover != "" {
		cfg.EnableFailover, err = strconv.ParseBool(enableFailover)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("Unable to parse SUMO_ENABLE_FAILOVER: %v", err))
		}
	}

	if cfg.EnableFailover == true {
		if cfg.S3BucketName == "" {
			allErrors = append(allErrors, "SUMO_S3_BUCKET_NAME not set in environment variable")
		}
		if cfg.S3BucketRegion == "" {
			allErrors = append(allErrors, "SUMO_S3_BUCKET_REGION not set in environment variable")
		}
	}

	if numRetry != "" {
		customNumRetry, err := strconv.ParseInt(numRetry, 10, 32)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("Unable to parse SUMO_NUM_RETRIES: %v", err))
		} else {
			cfg.NumRetry = int(customNumRetry)
		}
	}

	if retrySleepTime != "" {
		customRetrySleepTime, err := strconv.ParseInt(retrySleepTime, 10, 32)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("Unable to parse SUMO_RETRY_SLEEP_TIME_MS: %v", err))
		} else {
			cfg.RetrySleepTime = time.Duration(customRetrySleepTime) * time.Millisecond
		}
	}

	if maxDataQueueLength != "" {
		customMaxDataQueueLength, err := strconv.ParseInt(maxDataQueueLength, 10, 32)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("Unable to parse SUMO_MAX_DATAQUEUE_LENGTH: %v", err))
		} else {
			cfg.MaxDataQueueLength = int(customMaxDataQueueLength)
		}

	}
	if maxConcurrentRequests != "" {
		customMaxConcurrentRequests, err := strconv.ParseInt(maxConcurrentRequests, 10, 32)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("Unable to parse SUMO_MAX_CONCURRENT_REQUESTS: %v", err))
		} else {
			cfg.MaxConcurrentRequests = int(customMaxConcurrentRequests)
		}

	}
	if logLevel != "" {
		customloglevel, err := logrus.ParseLevel(logLevel)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("Unable to parse SUMO_LOG_LEVEL: %v", err))
		} else {
			cfg.LogLevel = customloglevel
		}

	}

	// test valid log format type
	for _, logType := range cfg.LogTypes {
		if !utils.StringInSlice(strings.TrimSpace(logType), validLogTypes) {
			allErrors = append(allErrors, fmt.Sprintf("logType %s is unsupported", logType))
		}
	}

	if len(allErrors) > 0 {
		err = errors.New(strings.Join(allErrors, ", "))
	}

	return err
}
