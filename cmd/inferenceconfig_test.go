/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/spf13/pflag"
)

func TestBuildInferenceConfiguration(t *testing.T) {
	t.Run("omits sampling params when unset", func(t *testing.T) {
		conf := buildInferenceConfiguration(500, nil, nil)

		if conf.MaxTokens == nil || *conf.MaxTokens != 500 {
			t.Fatalf("expected maxTokens 500, got %v", conf.MaxTokens)
		}
		if conf.Temperature != nil {
			t.Fatalf("expected temperature to be omitted, got %v", conf.Temperature)
		}
		if conf.TopP != nil {
			t.Fatalf("expected topP to be omitted, got %v", conf.TopP)
		}
	})

	t.Run("includes only explicitly provided sampling params", func(t *testing.T) {
		temperature := float32(0.7)
		topP := float32(0.9)
		conf := buildInferenceConfiguration(500, &temperature, &topP)

		if conf.Temperature == nil || *conf.Temperature != 0.7 {
			t.Fatalf("expected temperature 0.7, got %v", conf.Temperature)
		}
		if conf.TopP == nil || *conf.TopP != 0.9 {
			t.Fatalf("expected topP 0.9, got %v", conf.TopP)
		}
	})
}

func TestOptionalFloat32Flag(t *testing.T) {
	t.Run("returns nil when flag was not set", func(t *testing.T) {
		flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
		flags.Float32("topP", 0.999, "top-P")

		value, err := optionalFloat32Flag(flags, "topP")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if value != nil {
			t.Fatalf("expected nil when flag unset, got %v", *value)
		}
	})

	t.Run("returns value when flag was set", func(t *testing.T) {
		flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
		flags.Float32("topP", 0.999, "top-P")
		if err := flags.Set("topP", "0.5"); err != nil {
			t.Fatalf("Set error: %v", err)
		}

		value, err := optionalFloat32Flag(flags, "topP")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if value == nil || *value != 0.5 {
			t.Fatalf("expected 0.5, got %v", value)
		}
	})
}

func TestStripSamplingParams(t *testing.T) {
	t.Run("nil input yields nil", func(t *testing.T) {
		if got := stripSamplingParams(nil); got != nil {
			t.Fatalf("expected nil, got %#v", got)
		}
	})

	t.Run("removes temperature and topP but keeps maxTokens", func(t *testing.T) {
		in := &types.InferenceConfiguration{
			MaxTokens:   aws.Int32(500),
			Temperature: aws.Float32(1.0),
			TopP:        aws.Float32(0.999),
		}

		out := stripSamplingParams(in)
		if out.MaxTokens == nil || *out.MaxTokens != 500 {
			t.Fatalf("expected maxTokens 500, got %v", out.MaxTokens)
		}
		if out.Temperature != nil {
			t.Fatalf("expected temperature to be stripped, got %v", out.Temperature)
		}
		if out.TopP != nil {
			t.Fatalf("expected topP to be stripped, got %v", out.TopP)
		}
	})
}

func TestIsDeprecatedSamplingParamsError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "unrelated error",
			err:  errors.New("model not found"),
			want: false,
		},
		{
			name: "temperature deprecated",
			err:  errors.New("The model returned the following errors: `temperature` is deprecated for this model."),
			want: true,
		},
		{
			name: "top_p deprecated",
			err:  errors.New("The model returned the following errors: `top_p` is deprecated for this model."),
			want: true,
		},
		{
			name: "deprecated but unrelated field",
			err:  errors.New("The model returned the following errors: `foo` is deprecated for this model."),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDeprecatedSamplingParamsError(tt.err); got != tt.want {
				t.Fatalf("isDeprecatedSamplingParamsError() = %v, want %v", got, tt.want)
			}
		})
	}
}
