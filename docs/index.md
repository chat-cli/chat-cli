# Chat-CLI

Chat-CLI is a little terminal based program that lets you interact with LLMs available via [Amazon Bedrock](https://aws.amazon.com/bedrock).

## Quick start

Using [Homebrew]() do this:

```shell
    brew tap chat-cli/chat-cli
    brew install chat-cli
```

If you have an [AWS account](#prereqs), and you have [enabled model access](#prereqs) for Anthopic Claude Sonnet 3.5, Amazon Nova Micro and Amazon Nova Canvas, you can do this:

```shell
   # set up your AWS credentials on your machine using the AWS CLI
   aws configure

   # send a prompt to Anthropic Claude Sonnet 3.5
   chat-cli prompt "What is AWS?"

   # read contents of a file to Chat-CLI via stdin
   cat your-file.go | chat-cli prompt "explain this code"

   # start an interactive chat session using Amazon Nova Micro
   chat-cli chat

   # generate an image from text using Amazon Nova Canvas
   chat-cli image "generate an image of a cat driving a car"
```

## Contents

```{toctree}
---
maxdepth: 3
---
setup
models
```