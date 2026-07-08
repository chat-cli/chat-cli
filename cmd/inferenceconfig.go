/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/spf13/pflag"
)

func buildInferenceConfiguration(maxTokens int32, temperature, topP *float32) types.InferenceConfiguration {
	conf := types.InferenceConfiguration{
		MaxTokens: &maxTokens,
	}
	if temperature != nil {
		conf.Temperature = temperature
	}
	if topP != nil {
		conf.TopP = topP
	}
	return conf
}

func optionalFloat32Flag(flags *pflag.FlagSet, name string) (*float32, error) {
	if !flags.Changed(name) {
		return nil, nil
	}

	value, err := flags.GetFloat32(name)
	if err != nil {
		return nil, err
	}

	return &value, nil
}

// stripSamplingParams returns a copy of conf with temperature and topP
// removed. Some newer models (e.g. Claude Sonnet 5) reject these fields
// entirely rather than accepting them at default values.
func stripSamplingParams(conf *types.InferenceConfiguration) *types.InferenceConfiguration {
	if conf == nil {
		return nil
	}

	return &types.InferenceConfiguration{
		MaxTokens: conf.MaxTokens,
	}
}

func isDeprecatedSamplingParamsError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "deprecated") {
		return false
	}

	return strings.Contains(msg, "temperature") || strings.Contains(msg, "top_p") || strings.Contains(msg, "topp")
}

func converseWithFallbacks(ctx context.Context, svc *bedrockruntime.Client, input *bedrockruntime.ConverseInput) (*bedrockruntime.ConverseOutput, error) {
	output, err := svc.Converse(ctx, input)
	if err == nil {
		return output, nil
	}

	if hasSystemCachePoint(input.System) || (len(input.Messages) > 0 && hasContentCachePoint(input.Messages[0].Content)) {
		log.Printf("prompt caching not supported for this request, retrying without it: %v", err)
		input.System = stripSystemCachePoints(input.System)
		if len(input.Messages) > 0 {
			input.Messages[0].Content = stripContentCachePoints(input.Messages[0].Content)
		}
		output, err = svc.Converse(ctx, input)
		if err == nil {
			return output, nil
		}
	}

	if isDeprecatedSamplingParamsError(err) {
		log.Printf("sampling parameters not supported for this model, retrying without temperature/topP: %v", err)
		input.InferenceConfig = stripSamplingParams(input.InferenceConfig)
		output, err = svc.Converse(ctx, input)
	}

	return output, err
}

func converseStreamWithFallbacks(ctx context.Context, svc *bedrockruntime.Client, input *bedrockruntime.ConverseStreamInput) (*bedrockruntime.ConverseStreamOutput, error) {
	output, err := svc.ConverseStream(ctx, input)
	if err == nil {
		return output, nil
	}

	if hasSystemCachePoint(input.System) || (len(input.Messages) > 0 && hasContentCachePoint(input.Messages[0].Content)) {
		log.Printf("prompt caching not supported for this request, retrying without it: %v", err)
		input.System = stripSystemCachePoints(input.System)
		if len(input.Messages) > 0 {
			input.Messages[0].Content = stripContentCachePoints(input.Messages[0].Content)
		}
		output, err = svc.ConverseStream(ctx, input)
		if err == nil {
			return output, nil
		}
	}

	if isDeprecatedSamplingParamsError(err) {
		log.Printf("sampling parameters not supported for this model, retrying without temperature/topP: %v", err)
		input.InferenceConfig = stripSamplingParams(input.InferenceConfig)
		output, err = svc.ConverseStream(ctx, input)
	}

	return output, err
}
