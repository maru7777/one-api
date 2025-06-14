# Function calling

- <https://platform.openai.com/docs/guides/function-calling>

## ChatCompletion API

### ChatCompletion Make Call

```js
import { OpenAI } from "openai";

const openai = new OpenAI();

const tools = [
  {
    type: "function",
    function: {
      name: "search_knowledge_base",
      description:
        "Query a knowledge base to retrieve relevant info on a topic.",
      parameters: {
        type: "object",
        properties: {
          query: {
            type: "string",
            description: "The user question or search query.",
          },
          options: {
            type: "object",
            properties: {
              num_results: {
                type: "number",
                description: "Number of top results to return.",
              },
              domain_filter: {
                type: ["string", "null"],
                description:
                  "Optional domain to narrow the search (e.g. 'finance', 'medical'). Pass null if not needed.",
              },
              sort_by: {
                type: ["string", "null"],
                enum: ["relevance", "date", "popularity", "alphabetical"],
                description: "How to sort results. Pass null if not needed.",
              },
            },
            required: ["num_results", "domain_filter", "sort_by"],
            additionalProperties: false,
          },
        },
        required: ["query", "options"],
        additionalProperties: false,
      },
      strict: true,
    },
  },
];

const completion = await openai.chat.completions.create({
  model: "gpt-4.1",
  messages: [
    {
      role: "user",
      content:
        "Can you find information about ChatGPT in the AI knowledge base?",
    },
  ],
  tools,
  store: true,
});

console.log(completion.choices[0].message.tool_calls);
```

```output
[{
    "id": "call_4567xyz",
    "type": "function",
    "function": {
        "name": "search_knowledge_base",
        "arguments": "{\"query\":\"What is ChatGPT?\",\"options\":{\"num_results\":3,\"domain_filter\":null,\"sort_by\":\"relevance\"}}"
    }
}]
```

Supply model with results – so it can incorporate them into its final response.

```js
messages.push(completion.choices[0].message); // append model's function call message
messages.push({
  // append result message
  role: "tool",
  tool_call_id: toolCall.id,
  content: result.toString(),
});

const completion2 = await openai.chat.completions.create({
  model: "gpt-4.1",
  messages,
  tools,
  store: true,
});

console.log(completion2.choices[0].message.content);
```

## Response API

### Response Make Call

```js
import { OpenAI } from "openai";

const openai = new OpenAI();

const tools = [
  {
    type: "function",
    name: "get_weather",
    description: "Get current temperature for a given location.",
    parameters: {
      type: "object",
      properties: {
        location: {
          type: "string",
          description: "City and country e.g. Bogotá, Colombia",
        },
      },
      required: ["location"],
      additionalProperties: false,
    },
  },
];

const response = await openai.responses.create({
  model: "gpt-4.1",
  input: [
    { role: "user", content: "What is the weather like in Paris today?" },
  ],
  tools,
});

console.log(response.output);
```

```output
[{
    "type": "function_call",
    "id": "fc_12345xyz",
    "call_id": "call_12345xyz",
    "name": "get_weather",
    "arguments": "{\"location\":\"Paris, France\"}"
}]
```

Supply model with results – so it can incorporate them into its final response.

```js
input.push(toolCall); // append model's function call message
input.push({
  // append result message
  type: "function_call_output",
  call_id: toolCall.call_id,
  output: result.toString(),
});

const response2 = await openai.responses.create({
  model: "gpt-4.1",
  input,
  tools,
  store: true,
});

console.log(response2.output_text);
```
