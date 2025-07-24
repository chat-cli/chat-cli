# 💬 chat-cli 💬

A little terminal based program that lets you interact with LLMs available via [Amazon Bedrock](https://aws.amazon.com/bedrock).

![Chat Chat Chat](docs/images/index-01.png)

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
    brew tap chat-cli/chat-cli
    brew install chat-cli
```

Notes:

- You won't need Go installed on your system to use the pre-packaged binaries or Homebrew
- These are currently unsigned binary files. For most systems, this will not be an issue, but on MacOS you will need to [follow these instructions](https://support.apple.com/guide/mac-help/open-a-mac-app-from-an-unidentified-developer-mh40616/mac).

### Build from source

You will need [Go](https://go.dev) v1.22.1 installed on your system. You can type `go version` to ensure you have the correct version installed.

To build the project from source, clone this repository to your local machine and use [Make](https://www.gnu.org/software/make/manual/make.html) to build the binary.

```shell
    git clone git@github.com:go-micah/chat-cli.git
    cd chat-cli
    make
```

## Run

To run the program from within the same directory use the following command syntax. 
 
```shell
    ./bin/chat-cli <command> <args> <flags>
```

If you downloaded a pre-packaged binary or used Homebrew to install your path will be different. You can add your binary to your path (Homebrew does this for you) and then you can just do the following:

```shell
    chat-cli <command> <args> <flags>
```

## Help

You can get help at anytime with the `--help` flag. Typing `--help` after any command will display args and flags available to that command.

## Commands

There are currently three ways to interact with foundation models through this interface.

1. Send a single prompt to an LLM from the command line using the `prompt` command
2. Start an interactive chat with an LLM using the `chat` command
3. Generate an image with the `image` command

## Configuration

You can manage persistent configuration settings using the `config` command. This allows you to set default values for model-id and custom-arn that will be used automatically by the chat and prompt commands.

### Setting Configuration Values

```shell
# Set a default model ID
chat-cli config set model-id "anthropic.claude-3-5-sonnet-20240620-v1:0"

# Set a custom ARN for marketplace or cross-region models
chat-cli config set custom-arn "arn:aws:bedrock:us-west-2::foundation-model/custom-model"
```

### Viewing Configuration

```shell
# List all current configuration values
chat-cli config list
```

### Removing Configuration Values

```shell
# Remove a specific configuration value
chat-cli config unset model-id
chat-cli config unset custom-arn
```

### Configuration Precedence

The configuration system follows a clear precedence order:

1. **Command line flags** (highest priority) - Values specified with `--model-id` or `--custom-arn`
2. **Configuration file** - Values set with `chat-cli config set`
3. **Built-in defaults** (lowest priority) - `anthropic.claude-3-5-sonnet-20240620-v1:0` for model-id

**Important:** When both `model-id` and `custom-arn` are set, `custom-arn` takes precedence over `model-id`. This allows you to override the default model with a custom marketplace or cross-region model.

### Supported Configuration Keys

- `model-id`: The default model identifier to use for chat and prompt commands
- `custom-arn`: A custom ARN from Bedrock marketplace or for cross-region inference

## Prompt

You can send a one liner prompt like this:

```shell
    chat-cli prompt "How are you today?"
```

You can also read in a file from `stdin` as part of your prompt like this:

```shell
    cat myfile.go | chat-cli prompt "explain this code"
```

    or

```shell
    chat-cli prompt "explain this code" < myfile.go
```

This will add `<document></document>` tags around your document ahead of your prompt. This syntax works especially well with [Anthropic Claude](https://www.anthropic.com/product). Other models may produce different results.

## Chat

You can start an interactive chat sessions which will remember your conversation as you chat back and forth with the LLM.

You can start an interactive chat session like this:

```shell
    chat-cli
```

- Type `quit` to quit the interactive chat session.
- All chat flags (model-id, custom-arn, chat-id, etc.) work directly with the root command

### Saving and Restoring Chat Sessions

Starting a chat session with the `chat-cli` command will automatically save your chats to a local sqlite database. If you would like to restore a prior chat session you can do so in the following way:

Start by using the `chat list` command to list 10 most recent chat sessions.

```shell
    chat-cli chat list
```
This will print a list that looks something like the following:

```
❯ go run main.go chat list
2024-12-17T04:29:59Z | 9be2adda-5966-45c9-8a07-f7a7d486ca36 | How do I get started with AWS?

2024-12-17T04:25:53Z | 07927821-f443-4e92-84c6-86d6fa30ebf2 | What't the best way to decide which car 
2024-12-17T04:23:57Z | 6ecdece8-9547-4b8b-9f36-2b92df2f84d6 | What is the best way to decide on which 
2024-12-16T04:29:09Z | 879c2dd7-ba3d-4f59-a576-a1ce556ceb4e | What do you know about optics?

2024-12-16T04:28:52Z | 3a51ea83-93df-4af4-a1b3-d1ce89d845d9 | What can you tell me about electronics?

2024-12-16T04:25:14Z | e16d52a8-83a9-4dc6-8e74-e41610689a9e | What is a Go package for printing markdo
2024-12-16T04:24:35Z | 7c4764e1-029d-4ebe-a7d6-43ef230e5117 | Can you help me write a poem about dogs?
2024-12-15T05:25:14Z | 5b2c9fb0-9ed4-4616-90be-b482bc640f8c | Can you summarize what you know about Gi
2024-12-15T05:24:04Z | 042ce5bc-a693-4e8b-9db6-eb4834b5dbac | What do you know about the Go programmin
2024-12-15T04:28:47Z | 56614689-356c-4d54-bb2c-10bd5af56b93 | How are you today?
```

Find the `chat-id` that corresponds to the chat session you would like to load and copy it to your clipboard. Once copied you can load that chat session like this:

```shell
    chat-cli --chat-id 9be2adda-5966-45c9-8a07-f7a7d486ca36
```

This will print out the saved chat and leave you at a prompt where you can pick up where you left off. Future chats will continue to save with the same `chat-id` as you go.

Please note: Eventually your chat session will result in a very large prompt context. Depending on the LLM you are using, you may get an error. Consider starting a new session when your chat session gets really lengthy!

## List Models

You can get a list of all supported models in your current region like this:

```shell
    chat-cli models list
```

Please notes, this is the full list of all possible models. You will need to enable access for any models you'd like to use.

## LLMs

Currently all text based LLMs available through Amazon Bedrock are supported. The LLMs you wish to use must be enabled within Amazon Bedrock.

To switch LLMs, use the `--model-id` flag. 

You can supply the exact model id from the list above like so:

```shell
    chat-cli prompt "How are you today?" --model-id cohere.command-text-v14
```

## Streaming Response

By default, responses will stream to the command line as they are generated. This can be disabled using the `--no-stream` flag with the prompt command. Not all models offer a streaming response capability.

You can disable streaming like this:

```shell
    chat-cli prompt "What is event driven architecture?" --no-stream
```

Only streaming response capable models can be used with the `chat` command.

## Model Config

There are several flags you can use to override the default config settings. Not all config settings are used by each model.

    --max-tokens defaults to 500
    --temperature defaults to 1.0
    --topP defaults to 0.999

## Image Attachments

Some LLMs support uploading an image. Images can be either png or jpg and must be less than 5MB. To upload an image do the following:

```shell
    chat-cli prompt "Explain this image" --image IMG_1234.JPG
```

Please note this only works with supported models.

## Image

With the `image` command you can generate images with any supported Foundation Model. Simply follow the syntax below:

```shell
    chat-cli image "Generate an image of a cat eating cereal"
```

You can specify the model with the `--model-id` flag set to model's full model id or family name. You can also specify an output filename with the `--filename` flag.

