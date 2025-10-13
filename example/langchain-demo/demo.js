require('dotenv').config();

const { ChatOpenAI } = require('@langchain/openai');
const { HumanMessage, AIMessage } = require('@langchain/core/messages');

async function main() {
  console.log('🚀 Testing COO-LLM with LangChain...\n');

  // Initialize ChatOpenAI with custom base URL pointing to our COO-LLM server
  const llm = new ChatOpenAI({
    modelName: 'gemini-2.5-flash',
    openAIApiKey: 'test-12', // Valid API key for COO-LLM server
    temperature: 0.7,
    configuration: {
      baseURL: 'http://localhost:2906/v1', // Point to our COO-LLM server
    },
  });

  try {
    console.log('Sending request to COO-LLM server...');

    // Test 1: Simple message
    console.log('\nTesting simple message:');
    const response1 = await llm.invoke('Hello! Can you introduce yourself?');
    console.log('Response:', response1.content);

    // Test 2: Conversation with history
    console.log('\nTesting conversation history:');
    const messages = [
      new HumanMessage('What is the capital of France?'),
      new AIMessage('The capital of France is Paris.'),
      new HumanMessage('What about the population?')
    ];

    const response2 = await llm.invoke(messages);
    console.log('Response:', response2.content);

    // Test 3: Check usage information
    console.log('\n Checking usage information:');
    console.log('Response metadata:', response2.response_metadata);
    if (response2.usage_metadata) {
      console.log('Usage:', {
        input_tokens: response2.usage_metadata.input_tokens,
        output_tokens: response2.usage_metadata.output_tokens,
        total_tokens: response2.usage_metadata.total_tokens,
      });
    }

    console.log('\n🎉 All tests passed! COO-LLM is compatible with LangChain!');

  } catch (error) {
    console.error('❌ Error:', error.message);
    console.error('Full error:', error);

    if (error.message.includes('ECONNREFUSED')) {
      console.log('\n💡 Make sure COO-LLM server is running on http://localhost:2906');
      console.log('   Run: go run cmd/coo-llm/main.go');
    }
  }
}

// Handle unhandled promise rejections
process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled Rejection at:', promise, 'reason:', reason);
});

main();