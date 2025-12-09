# TChat: Your Privacy-First, 100% Local AI Assistant for the Terminal

[![Go Version](https://img.shields.io/github/go-mod/go-version/vnaveen-mh/tchat)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)

**TChat is a privacy-first, fully local AI chat assistant for your terminal.** Built with **Go**, **Genkit**, and **Ollama**, it brings modern LLM and vision capabilities into a fast, keyboard-only workflow—without sending a single byte to the cloud.

![TChat Demo](images/tchat-demo.gif)

## Why TChat?

In a world full of cloud-heavy AI tools, TChat gives developers full control:

- 🏠 **100% Local and Private**

  Everything runs via Ollama on your machine. No tokens, no API keys, no telemetry.

- 🖼️ **Multimodal (Text + Vision)**

  TChat auto-detects image paths/URLs and includes media messages while making Genkit calls

- ⌨️ **Real Terminal UX**

  Includes command history, streaming tokens, rich colors, and graceful Ctrl-C cancellations.

- 🗄️ **Persistent & Searchable**

  Every session is stored in SQLite so you can resume or analyze usage later.

- ⚙️ **Developer-First Commands**

  /model, /system, /copy, /stats, /history, etc.

## Key Features

### 1. Complete Privacy

TChat runs entirely on your machine. Your conversations are yours alone. Switch between local Ollama models effortlessly.

```

> /model
> Current model: ollama/llama3.1:8b

Available models:
[1] ollama/gpt-oss:20b
[2] ollama/gemma3:4b
[3] ollama/qwen3-vl:8b
[4] ollama/llama3.1:8b
[5] ollama/qwen2.5-coder:7b

model> 5
✓ Switched to ollama/qwen2.5-coder:7b

```

### 2. Multimodal AI in Your Terminal

Switch to a model that support VL and analyze images directly from your command line. Just include the path to a local image or a URL.

```

> Analyze this architecture diagram: ~/docs/system-design.png
> 📷 Detected 1 image(s): [~/docs/system-design.png]

<system>: This diagram illustrates a classic microservices architecture...

```

### 3. A Real CLI, Not a Script

- **Command History**: Navigate your input history with arrow keys.
- **Streaming Responses**: Get real-time feedback as the AI generates text.
- **Cancellation**: Stop a generation mid-stream with `Ctrl-C`.
- **Color-Coded Output**: Enhanced readability with themed colors.

### 4. Persistent History & Stats

All conversations are stored in a local SQLite database, enabling:

- **Session Recovery**: Restart TChat without losing your conversation history.
- **Usage Insights**: Track model performance, response times, and more with `/stats`.

```

> /stats

# Conversation Statistics

Total conversations: 152
Unique models used: 3
Avg response time: 987 ms
...

```

### 5. Powerful Commands

Use `/` commands to control the assistant:

| Command    | Description                  |
| ---------- | ---------------------------- |
| `/help`    | List all commands            |
| `/model`   | List & switch Ollama models  |
| `/system`  | Set system prompt            |
| `/copy`    | Copy last AI Response        |
| `/history` | Show history details         |
| `/clear`   | Clear screen                 |
| `/reset`   | Reset conversation history   |
| `/stats`   | Show database-based insights |
| `/config`  | Print current configuration  |
| `/quit`    | Exit TChat                   |

## Getting Started

### Prerequisites

- **Go**: Version 1.24 or later.
- **Ollama**: Must be installed and running. [Install Ollama](https://ollama.com).

### Installation

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/vnaveen-mh/tchat.git
    cd tchat
    ```

2.  **Pull Ollama models:**
    Make sure you have at least one model pulled. For multimodal capabilities, use a model like `qwen3-vl` or `llava`.

    ```bash
    ollama pull llama3.1:8b         # general reasoning
    ollama pull qwen2.5-coder:7b    # coding-focused
    ollama pull qwen3-vl:8b         # multimodal (vision)
    ```

3.  **Build the binary:**

    ```bash
    go build -o tchat .
    ```

4.  **Run TChat:**
    ```bash
    ./tchat
    ```
    Or, to make it accessible from anywhere, move it to a directory in your `PATH`:
    ```bash
    sudo mv tchat /usr/local/bin/
    ```

### Running with Docker

As an alternative to building from source, you can run TChat using the official Docker image.

1.  **Pull the image from Docker Hub:**

    ```bash
    docker pull vnaveenmh/tchat:latest
    ```

2.  **Run the container:**
    The key challenge when running TChat in Docker is connecting the container to the Ollama instance running on your host machine. By default, the container's `localhost` is not the same as your computer's `localhost`.

    You must explicitly tell TChat how to reach Ollama using the `OLLAMA_HOST` environment variable.

    **For macOS or Windows (Docker Desktop):**
    This is the most straightforward setup. Use the special `host.docker.internal` DNS name.

    ```bash
    docker run -it --rm \
      -v ~/.tchat:/app/data \
      -e TCHAT_APPDIR="/app/data" \
      -e OLLAMA_HOST="http://host.docker.internal:11434" \
      vnaveenmh/tchat:latest
    ```

    **For Linux:**
    The most reliable method is to use `--network="host"`, which makes the container share your host's network stack. This allows `localhost` inside the container to correctly point to your host machine.

    ```bash
    docker run -it --rm \
      --network="host" \
      -v ~/.tchat:/app/data \
      -e TCHAT_APPDIR="/app/data" \
      vnaveenmh/tchat:latest
    ```

    _Note: If `--network="host"` is not suitable for your environment, you can find your host's IP on the docker bridge (e.g., `172.17.0.1`) and set it via `-e OLLAMA_HOST="http://<YOUR_HOST_IP>:11434"`._

    **Explanation of the flags:**

    - `-v ~/.tchat:/app/data`: Correctly maps your local `~/.tchat` directory to `/app/data` inside the container for persistent storage.
    - `-e TCHAT_APPDIR="/app/data"`: Tells TChat to use this directory for its database, logs, and history.
    - `-e OLLAMA_HOST="..."`: **(Crucial)** Points TChat to your Ollama server.
    - `--network="host"` (Linux): A simple way to resolve network communication on Linux.

## Usage

Once running, you can start a conversation immediately. Type `/help` to see the list of available commands and features.

**Quick Example**

```
tchat> /show

Current system prompt: You are a helpful Go developer
Current model: ollama/qwen2.5-coder:7b

tchat> Explain goroutines with an example
 <model response>

tchat> /cp
✓ Copied last response to clipboard (512 characters)
```

## Configuration

TChat is designed to work out-of-the-box, but you can customize its behavior through environment variables and a configuration file.

### Application Directory

By default, TChat stores all its data in the `~/.tchat` directory. This includes:

- `tchat.db`: The SQLite database for conversation history.
- `logs/`: Log files for debugging.
- `history`: The readline command history file.
- `config.json`: An optional file for custom configurations. If the file does not exist, TChat falls back to sane defaults.

To use a different application directory, set the `TCHAT_APPDIR` environment variable:

```bash
export TCHAT_APPDIR=/path/to/your/tchat_data
./tchat
```

### Environment Variables

- `OLLAMA_HOST`: Use this to specify a different Ollama server address if it's not running on the default `http://localhost:11434`.

  ```bash
  export OLLAMA_HOST=http://192.168.1.100:11434
  ./tchat
  ```

- `TCHAT_APPDIR`: As mentioned above, this overrides the default `~/.tchat` application directory.

### Configuration File

You can create an optional `config.json` file in your application directory (`~/.tchat` by default) to persist certain settings.

**Example `config.json`:**

```json
{
  "model": "ollama/codellama",
  "system_prompt": "You are an expert Python programmer.",
  "log_level": "info"
}
```

Supported fields:

- `model`: The default Ollama model to use on startup.
- `system_prompt`: A custom system prompt to use for conversations.
- `log_level`: The logging level (`debug`, `info`, `warn`, `error`).

## Technical Architecture

TChat is built with a clean and modular architecture, making it easy to maintain and extend.

- **Core Framework**: [Google Genkit for Go](https://genkit.dev/docs/get-started/?lang=go) provides a unified interface for AI models.
- **Local AI**: [Ollama](https://ollama.com) serves local models.
- **Database**: [SQLite](https://www.sqlite.org/index.html) for zero-configuration persistent storage.
- **CLI Interface**: `chzyer/readline` for a professional terminal experience.

            ┌───────────────┐
            │    Terminal   │
            └───────┬───────┘
                    │ User input
                    ▼
           ┌────────────────────┐
           │     TChat REPL     │
           └───────┬────────────┘
                    │ Routes text + images
                    ▼
       ┌────────────────────────────┐
       │     Genkit (Go SDK)        │
       └───────────┬────────────────┘
                   ▼
          ┌──────────────────┐
          │   Ollama Server  │  ← local models (text + VL)
          └──────────────────┘

          ┌──────────────────┐
          │     SQLite DB    │  ← history, stats, configs
          └──────────────────┘

The project structure separates concerns logically:

```
├── internal/
│   ├── command/      # Command implementations (/help, /model, etc.)
│   ├── config/       # Configuration management
│   ├── db/           # SQLite storage layer
│   ├── flows/        # AI generation and streaming logic
│   ├── history/      # In-memory history management
│   ├── media/        # Image processing for multimodal input
│   └── ollama/       # Ollama integration helpers
└── main.go           # Application entry point and REPL
```

## Roadmap

- **Tool Calling Support**

  Integrate Genkit’s Tool Calling to allow actions such as reading local files, running small utilities, and extending TChat with custom tools.

- **MCP (Model Context Protocol) Support**

  Add MCP server integration (via Genkit) so TChat can connect to external tools and provide richer, structured interactions.

- **Multi-Line Input Editing**

  Improve the terminal UX with proper multi-line editing for long prompts, code blocks, and structured instructions.

## Contributing

Contributions are welcome! Whether it's bug reports, feature requests, or pull requests, your input is valued. Please feel free to open an issue to discuss your ideas.

## License

This project is licensed under the **MIT License**. See the [LICENSE](LICENSE) file for details.
