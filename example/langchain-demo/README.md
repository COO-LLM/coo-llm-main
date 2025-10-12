# COO-LLM LangChain Demo

Demo testing COO-LLM compatibility with LangChain.js

## Prerequisites

- Node.js >= 18.0.0
- COO-LLM server running on `http://localhost:8080`

## Setup

1. Install dependencies:
```bash
npm install
```

2. Copy environment file:
```bash
cp .env.example .env
```

3. Edit `.env` if needed (default should work with local COO-LLM server)

## Running the Demo

1. Start COO-LLM server in another terminal:
```bash
cd ..
go run cmd/coo-llm/main.go
```

2. Run the demo:
```bash
npm start
```

## What it tests

1. **Simple message**: Basic chat completion with string input
2. **Conversation history**: Multi-turn conversation using LangChain message objects
3. **Usage information**: Token usage and response metadata

## Expected Output

```
🚀 Testing COO-LLM with LangChain...

📤 Sending request to COO-LLM server...

1️⃣ Testing simple message:
✅ Response: [AI response here]

2️⃣ Testing conversation history:
✅ Response: [AI response here]

3️⃣ Checking usage information:
Response metadata: {...}
Usage: { input_tokens: X, output_tokens: Y, total_tokens: Z }

🎉 All tests passed! COO-LLM is compatible with LangChain!
```

## Troubleshooting

- **ECONNREFUSED**: Make sure COO-LLM server is running on port 8080
- **API Key errors**: COO-LLM ignores the API key, but LangChain requires one
- **Model not found**: Check that the model is configured in COO-LLM config