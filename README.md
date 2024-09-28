# screen-complete

> Screen complete is a proof of concept universal screenshot-based text completion tool. Inspired by tools like cursor and github copilot, it allows you to fill in arbitrary selected text on your screen using a hot key.

## Quick Start (Windows/OSX)

1. Download the latest version from the [releases page](https://github.com/sshh12/screen-complete/releases)
2. Create a config file `screen_complete.yml` with `openai_api_key: ...` (in the same directory)
3. Run `screen-complete` (OSX may require enabling screen recording permissions)

The UI/UX is extremely minimal, on a page you want to fill in text:
1. Place your cursor where you want to type text (or select text you want to replace)
1. Move your mouse (without clicking) the top left corner of the applicable window
2. Hold down `Ctrl+Q`
3. Move your cursor to the bottom right corner of the applicable window
4. Release `Ctrl+Q`

## Examples

| Description | Image |
|-------------|-------|
| Writing text in a google doc | ![chrome_lOtDKw9Vsd](https://github.com/user-attachments/assets/eb7e2a84-52e5-480c-9fd8-b744c3b04f13) |
| Filling in the title of a GitHub issue | ![chrome_rLIIqjoeJE](https://github.com/user-attachments/assets/257c990f-fcfd-46cf-b0c8-d0c6a28e6137) | 
| Drafting a reddit comment | ![chrome_lnoue13hYT](https://github.com/user-attachments/assets/0f3a263c-8ab6-42da-8985-2369a2fadd5e) |


## Configuration

This tool currently supports OpenAI and Azure OpenAI. Only fields for Azure OR OpenAI are required.

### Via Environment Variables

```
AZURE_OPENAI_API_KEY=...
AZURE_OPENAI_ENDPOINT=...
AZURE_OPENAI_DEPLOYMENT=...
OPENAI_API_KEY=...
OPENAI_MODEL=... (optional)
```

### Via `screen_complete.yml`

```yaml
azure_openai_api_key: ...
azure_openai_endpoint: ...
azure_openai_deployment: ...
openai_api_key: ...
openai_model: ... (optional)
```

## Building

### Windows

1. Download the Mingw, then set system environment variables C:\mingw64\bin to the Path
2. `go build -o screen-complete.exe cmd\screen_complete\main.go`

### MacOS

1. `xcode-select --install`
2. `go build -o screen-complete cmd\screen_complete\main.go`

## Privacy

Your screen is only captured and sent to OpenAI when you release the hot key.