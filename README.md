# 💬 chat-cli 💬

A little terminal based program that lets you interact with LLMs available via [Amazon Bedrock](https://aws.amazon.com/bedrock).

## Prerequisites

1. You will need an [AWS account](https://aws.amazon.com)
2. You will need to enable the LLMs you wish to use in Amazon Bedrock via the [Model Access](https://us-east-1.console.aws.amazon.com/bedrock/home?region=us-east-1#/modelaccess) page in the AWS Console. The default LLMs for both Chat and Prompt commands are provided by Anthropic, so it is recommended to enable these as a starting point.
3. You will need to install the [AWS CLI](https://docs.aws.amazon.com/cli/) tool and run `aws config` from the command line to set up credentials.

## Installation

At this time you can install chat-cli via pre-packaged binaries (thanks to [GoReleaser](https://goreleaser.com/)!) for your operating system/architecture combination of choice.

### Pre-Built Binaries

1. Head to https://github.com/chat-cli/chat-cli/releases/latest to find the binary for your setup.
2. Download and unzip to find a pre-compiled binary file that should work on your system.

### Homebrew

If you have Homebrew installed on your system you can do the following two commands:

```shell
    $ brew tap chat-cli/chat-cli
    $ brew install chat-cli
```

Notes:

- You won't need Go installed on your system to use the pre-packaged binaries or Homebrew
- These are currently unsigned binary files. For most systems, this will not be an issue, but on MacOS you will need to [follow these instructions](https://support.apple.com/guide/mac-help/open-a-mac-app-from-an-unidentified-developer-mh40616/mac).

### Build from source

You will need [Go](https://go.dev) v1.22.1 installed on your system. You can type `go version` to ensure you have the correct version installed.

To build the project from source, clone this repository to your local machine and use [Make](https://www.gnu.org/software/make/manual/make.html) to build the binary.

    $ git clone git@github.com:go-micah/chat-cli.git
    $ cd chat-cli
    $ make

## Run

To run the program from within the same directory use the following command syntax. 

    $ ./bin/chat-cli <command> <args> <flags>

If you downloaded a pre-packaged binary or used Homebrew to install your path will be different. You can add your binary to your path (Homebrew does this for you) and then you can just do the following:

    $ chat-cli <command> <args> <flags>

## Help

You can get help at anytime with the `--help` flag. Typing `--help` after any command will display args and flags available to that command.

## Commands

There are currently three ways to interact with foundation models through this interface.

1. Send a single prompt to an LLM from the command line using the `prompt` command
2. Start an interactive chat with an LLM using the `chat` command
3. Generate an image with the `image` command

## Prompt

You can send a one liner prompt like this:

    $ chat-cli prompt "How are you today?"

You can also read in a file from `stdin` as part of your prompt like this:

    $ cat myfile.go | chat-cli prompt "explain this code"

    or

    $ chat-cli prompt "explain this code" < myfile.go

This will add `<document></document>` tags around your document ahead of your prompt. This syntax works especially well with [Anthropic Claude](https://www.anthropic.com/product). Other models may produce different results.

## Chat

You can start an interactive chat sessions which will remember your conversation as you chat back and forth with the LLM.

You can start an interactive chat session like this:

    $ chat-cli chat

- Type `quit` to quit the interactive chat session.

## List Models

You can get a list of all supported models in your current region like this:

    $ chat-cli models list

Please notes, this is the full list of all possible models. You will need to enable access for any models you'd like to use.

## LLMs

Currently all text based LLMs available through Amazon Bedrock are supported. The LLMs you wish to use must be enabled within Amazon Bedrock.

To switch LLMs, use the `--model-id` flag. 

You can supply the exact model id from the list above like so:

    $ chat-cli prompt "How are you today?" --model-id cohere.command-text-v14

## Streaming Response

By default, responses will stream to the command line as they are generated. This can be disabled using the `--no-stream` flag with the prompt command. Not all models offer a streaming response capability.

You can disable streaming like this:

    $ chat-cli prompt "What is event driven architecture?" --no-stream

Only streaming response capable models can be used with the `chat` command.

## Model Config

There are several flags you can use to override the default config settings. Not all config settings are used by each model.

    --max-tokens defaults to 500
    --temperature defaults to 1.0
    --topP defaults to 0.999

## Image Attachments

Some LLMs support uploading an image. Images can be either png or jpg and must be less than 5MB. To upload an image do the following:

    $ chat-cli prompt "Explain this image" --image IMG_1234.JPG

Please note this only works with supported models.

## Image

With the `image` command you can generate images with any supported Foundation Model. Simply follow the syntax below:

    $ chat-cli image "Generate an image of a cat eating cereal"

You can specify the model with the `--model-id` flag set to model's full model id or family name. You can also specify an output filename with the `--filename` flag.

