# Usage

(config)=
## Config

Chat-CLI provides a configuration system that allows you to set persistent default values for commonly used settings. This eliminates the need to specify the same flags repeatedly when using the `chat` and `prompt` commands.

### Managing Configuration

#### Setting Values

Use the `config set` command to store default values:

```shell
# Set a default model ID
chat-cli config set model-id "anthropic.claude-3-5-sonnet-20240620-v1:0"

# Set a custom ARN for marketplace or cross-region models  
chat-cli config set custom-arn "arn:aws:bedrock:us-west-2::foundation-model/custom-model"
```

#### Viewing Configuration

List all current configuration values:

```shell
chat-cli config list
```

Example output:
```
Current configuration:
  model-id = anthropic.claude-3-5-sonnet-20240620-v1:0
  custom-arn = arn:aws:bedrock:us-west-2::foundation-model/custom-model
```

#### Removing Values

Remove specific configuration values when no longer needed:

```shell
chat-cli config unset model-id
chat-cli config unset custom-arn
```

### Configuration Precedence

The configuration system uses a clear precedence hierarchy to determine which values to use:

1. **Command line flags** (highest priority)
   - Values specified with `--model-id` or `--custom-arn` flags
   - Always override configuration file and defaults

2. **Configuration file** (medium priority)
   - Values set using `chat-cli config set`
   - Used when no command line flag is provided

3. **Built-in defaults** (lowest priority)
   - Default model: `anthropic.claude-3-5-sonnet-20240620-v1:0`
   - Used when no configuration or flags are set

### Custom ARN Priority

When both `model-id` and `custom-arn` are configured, `custom-arn` takes precedence. This design allows you to:

- Set a default `model-id` for regular use
- Override with `custom-arn` for marketplace or cross-region models
- Use command line flags to override either setting temporarily

### Supported Settings

| Setting | Description | Example |
|---------|-------------|---------|
| `model-id` | Default model identifier for Bedrock foundation models | `anthropic.claude-3-5-sonnet-20240620-v1:0` |
| `custom-arn` | Custom ARN for marketplace or cross-region inference | `arn:aws:bedrock:us-west-2::foundation-model/custom-model` |
| `system-prompt` | Default system prompt used by `chat` and `prompt` | `You are a terse, no-nonsense assistant.` |

### Configuration Storage

Configuration values are stored in a YAML file in your system's standard configuration directory:

- **macOS**: `~/Library/Application Support/chat-cli/config.yaml`
- **Linux**: `~/.config/chat-cli/config.yaml` 
- **Windows**: `%APPDATA%\chat-cli\config.yaml`

(prompt)=
## Prompt

Use `--system` to give the model a system prompt for a single one-shot request:

```shell
chat-cli prompt "How are you today?" --system "You are a terse, no-nonsense assistant."
```

If no `--system` flag is given, the persisted `system-prompt` config value (if any, see [Config](#config)) is used instead. If neither is set, no system prompt is sent — behavior is unchanged from before this feature existed.

(chat)=
## Chat

Use `--system` to set a system prompt for the whole interactive session (it applies at session start and doesn't change mid-conversation):

```shell
chat-cli --system "You are a terse, no-nonsense assistant."
```

Like `prompt`, this falls back to the persisted `system-prompt` config value, then to no system prompt at all.

### Tool Use

Pass `--tools` to let the model call tools mid-conversation:

```shell
chat-cli --tools
```

This is off by default — Bedrock doesn't expose whether a given model supports tool use, so `chat` behaves exactly as before unless you opt in. With `--tools` set, one built-in tool is available: `read_file`, which lets the model read a file in your current working directory (it can't read anything outside that directory). If the model asks for a tool that doesn't exist, or a tool call fails, you'll see the conversation continue normally — chat-cli reports the failure back to the model rather than crashing.

(image)=
## Image

