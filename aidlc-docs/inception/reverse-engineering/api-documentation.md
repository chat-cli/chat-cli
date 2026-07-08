# API Documentation

Chat-CLI exposes no network-facing REST/HTTP API — it is a CLI. This document treats the **CLI command surface** as the "external API" (what end users/scripts invoke) and documents the **internal package APIs** used across layers.

## CLI Commands (External API)

### `chat-cli` (root, no subcommand)
- **Purpose**: Start (or resume, with `--chat-id`) an interactive streaming chat session. Equivalent to `chat-cli chat`.
- **Persistent Flags** (also inherited by all subcommands): `-r, --region` (default `us-east-1`), `-m, --model-id` (default `anthropic.claude-3-5-sonnet-20240620-v1:0`), `--custom-arn`, `--chat-id`, `--temperature` (default `1.0`), `--topP` (default `0.999`), `--max-tokens` (default `500`).
- **Behavior**: Validates the model supports text output + streaming (unless `--custom-arn` is set, which skips validation), opens the local SQLite DB, replays prior messages if `--chat-id` given, then loops reading user input and streaming Bedrock responses until `quit`/`/quit`/Ctrl-C.

### `chat-cli chat`
- **Purpose**: Alias/explicit form of the root command's default behavior.
- **Flags**: Same as root (inherited persistent flags).

### `chat-cli chat list`
- **Purpose**: Print the 10 most recent chat sessions.
- **Output**: Tab-formatted table: `Created Date | Chat ID | Title` (title = first 40 chars of the stored message).

### `chat-cli prompt "<text>"`
- **Purpose**: Send a single one-shot prompt and print the response.
- **Args**: `args[0]` = prompt text (required, `cobra.MinimumNArgs(1)`); stdin (if piped) is appended as a `<document>` block.
- **Flags**: `-m, --model-id` (default `anthropic.claude-3-5-sonnet-20240620-v1:0`), `--custom-arn`, `-i, --image` (path to PNG/JPG < 5MB), `--no-stream` (return full response instead of streaming), `--temperature`, `--topP`, `--max-tokens`.
- **Behavior**: Validates model supports text (and image input, if `--image` set; and streaming, unless `--no-stream`) via `bedrock.GetFoundationModel`, then calls `bedrockruntime.Converse` (no-stream) or `ConverseStream` (default).

### `chat-cli image "<prompt>"`
- **Purpose**: Generate an image and save it to disk.
- **Args**: `args[0]` = prompt text (required); stdin document supported like `prompt`.
- **Flags**: `-m, --model-id` (default `amazon.nova-canvas-v1:0`), `--scale` (default `10`), `--steps` (default `10`), `--seed` (default `0`), `-f, --filename` (default: `<unix-timestamp>.jpg`).
- **Behavior**: Validates model supports `IMAGE` output; builds a provider-specific request body (Stability AI or Amazon Titan schema, chosen by `model.ModelDetails.ProviderName`); calls `bedrockruntime.InvokeModel`; decodes and writes the base64 image to disk.

### `chat-cli models list`
- **Purpose**: List all foundation models available in the configured region (hardcoded to `us-east-1` inside `listModels()`, independent of the `--region` flag — see code-quality-assessment.md).
- **Output**: Tab-formatted table: `Provider | Name | Model ID`.

### `chat-cli config set <key> <value>`
- **Purpose**: Persist a default config value. Supported keys: `model-id`, `custom-arn`.

### `chat-cli config unset <key>`
- **Purpose**: Remove a persisted config value. Supported keys: `model-id`, `custom-arn`.

### `chat-cli config list`
- **Purpose**: Print all currently set supported config values (or "No configuration values set").

### `chat-cli version`
- **Purpose**: Print CLI version (currently hardcoded `v0.5.3`) and `GOOS/GOARCH`.

## Internal APIs

### `config.FileManager` (`config/config.go`)
- **Methods**:
  - `NewFileManager(appName string) (*FileManager, error)` — constructs and initializes OS-specific paths.
  - `(fm *FileManager) InitializeViper() error` — configures and loads/creates the Viper-backed YAML config.
  - `(fm *FileManager) GetDBPath() string` — full path to the SQLite file.
  - `(fm *FileManager) GetDBDriver() string` — configured DB driver (`db_driver`, default `sqlite`).
  - `(fm *FileManager) GetConfigValue(key string, flagValue, defaultValue interface{}) interface{}` — precedence resolver (flag → config file → default); supports `string`, `int32`, `float32` flag types.
- **Parameters/Return Types**: See above; `GetConfigValue` returns `interface{}` and callers type-assert (e.g. `.(string)`).

### `db.Database` (`db/db.go`)
- **Methods**: `GetDB() *sql.DB`, `Connect() error`, `Close() error`, `Migrate() error`.
- **Implementations**: `sqlite.SQLiteDB`.

### `db.Migration` (`db/migrations.go`)
- **Methods**: `MigrateUp() error`, `MigrateDown() error`.
- **Implementations**: `sqlite.SQLiteMigration`.

### `factory.CreateDatabase` (`factory/database.go`)
- **Signature**: `CreateDatabase(config *db.Config) (db.Database, error)`.
- **Behavior**: Dispatches on `config.Driver`; only `"sqlite"` implemented today; unknown drivers return an error.

### `repository.Repository[T]` (`repository/base.go`)
- **Methods** (generic interface, not fully implemented by `ChatRepository`): `Create(entity *T) error`, `GetByID(id int) (*T, error)`, `Update(entity *T) error`, `Delete(id int) error`, `List() ([]T, error)`.

### `repository.ChatRepository` (`repository/chat.go`)
- **Methods**:
  - `NewChatRepository(db db.Database) *ChatRepository`
  - `(r *ChatRepository) Create(chat *Chat) error` — inserts a row; note `RETURNING id` in a raw SQL string with modernc.org/sqlite (see code-quality-assessment.md for a caveat).
  - `(r *ChatRepository) List() ([]Chat, error)` — 10 most recent chats, grouped by `chat_id`, newest first.
  - `(r *ChatRepository) GetMessages(chatId string) ([]Chat, error)` — full ordered transcript for one chat.
- **Parameters**: `Chat{ID int, ChatId string, Persona string, Message string, Created string}`.

### `utils` package (`utils/utils.go`, `utils/bubbleinput.go`)
- **Methods**:
  - `ProcessStreamingOutput(output *bedrockruntime.ConverseStreamOutput, handler StreamingOutputHandler) (types.Message, error)` — drains the Bedrock stream, invoking `handler(ctx, textDelta)` per chunk, returns the assembled `types.Message`.
  - `ReadImage(filename string) (data []byte, imageType string, err error)` — path-traversal-safe local image read; supports jpg/jpeg/png/gif/webp.
  - `DecodeImage(base64Image string) ([]byte, error)` — base64 decode.
  - `StringPrompt(label string) string` — TTY-aware input: uses `BubbleInput()` interactively, falls back to buffered stdin read otherwise.
  - `LoadDocument() (string, error)` — reads piped stdin (if not a TTY) and wraps it in `<document>...</document>`.
  - `BubbleInput() (string, bool)` — runs the BubbleTea `InputField` program and returns the submitted line (`/quit` normalized to `quit\n`).

## Data Models

### `repository.Chat` (`repository/chat.go`)
- **Fields**: `ID int`, `ChatId string`, `Persona string` ("User" or "Assistant"), `Message string`, `Created string` (populated from `created_at` on read).
- **Relationships**: Many `Chat` rows share one `ChatId` (a conversation); no foreign keys — SQLite `chats` table is flat.
- **Validation**: None at the Go struct level; all fields are `NOT NULL` at the SQL schema level (`db/sqlite/migrations.go`).

### `db.Config` (`db/db.go`)
- **Fields**: `Port int`, `Driver string`, `Host string`, `Name string`, `Username string`, `Password string`.
- **Relationships**: Passed to `factory.CreateDatabase`; only `Driver` and `Name` are actually used by the SQLite implementation today — `Port`/`Host`/`Username`/`Password` are unused placeholders for a future networked backend (e.g. Postgres).
- **Validation**: None.

### SQL Schema — `chats` table (`db/sqlite/migrations.go`)
- **Fields**: `id INTEGER PRIMARY KEY AUTOINCREMENT`, `chat_id TEXT NOT NULL`, `persona TEXT NOT NULL`, `message TEXT NOT NULL`, `created_at DATETIME DEFAULT CURRENT_TIMESTAMP`, `updated_at DATETIME DEFAULT CURRENT_TIMESTAMP`.
- **Relationships**: `chat_id` groups rows into conversations (no FK constraint).
- **Validation**: `NOT NULL` on `chat_id`/`persona`/`message`; trigger `chats_updated_at` keeps `updated_at` current on row update.
