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

### Web search Tool

Allow models to search the web for the latest information before generating a response.

Using the [Chat Completions API](/docs/api-reference/chat), you can directly access the fine-tuned models and tool used by [Search in ChatGPT](https://openai.com/index/introducing-chatgpt-search/).

When using Chat Completions, the model always retrieves information from the web before responding to your query. To use `web_search_preview` as a tool that models like `gpt-4o` and `gpt-4o-mini` invoke only when necessary, switch to using the [Responses API](/docs/guides/tools-web-search?api-mode=responses).

Currently, you need to use one of these models to use web search in Chat Completions:

- `gpt-4o-search-preview`
- `gpt-4o-mini-search-preview`

Web search parameter example

```bash
curl -X POST "https://api.openai.com/v1/chat/completions" \
    -H "Authorization: Bearer $OPENAI_API_KEY" \
    -H "Content-type: application/json" \
    -d '{
        "model": "gpt-4o-search-preview",
        "web_search_options": {},
        "messages": [{
            "role": "user",
            "content": "What was a positive news story from today?"
        }]
    }'
```

#### Output and citations

The API response item in the `choices` array will include:

- `message.content` with the text result from the model, inclusive of any inline citations
- `annotations` with a list of cited URLs

By default, the model's response will include inline citations for URLs found in the web search results. In addition to this, the `url_citation` annotation object will contain the URL and title of the cited source, as well as the start and end index characters in the model's response where those sources were used.

When displaying web results or information contained in web results to end users, inline citations must be made clearly visible and clickable in your user interface.

```json
[
  {
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "the model response is here...",
      "refusal": null,
      "annotations": [
        {
          "type": "url_citation",
          "url_citation": {
            "end_index": 985,
            "start_index": 764,
            "title": "Page title...",
            "url": "https://..."
          }
        }
      ]
    },
    "finish_reason": "stop"
  }
]
```

#### User location

To refine search results based on geography, you can specify an approximate user location using country, city, region, and/or timezone.

- The `city` and `region` fields are free text strings, like `Minneapolis` and `Minnesota` respectively.
- The `country` field is a two-letter [ISO country code](https://en.wikipedia.org/wiki/ISO_3166-1), like `US`.
- The `timezone` field is an [IANA timezone](https://timeapi.io/documentation/iana-timezones) like `America/Chicago`.

Note that user location is not supported for deep research models using web search.

Customizing user location

```bash
curl -X POST "https://api.openai.com/v1/chat/completions" \
    -H "Authorization: Bearer $OPENAI_API_KEY" \
    -H "Content-type: application/json" \
    -d '{
        "model": "gpt-4o-search-preview",
        "web_search_options": {
            "user_location": {
                "type": "approximate",
                "approximate": {
                    "country": "GB",
                    "city": "London",
                    "region": "London"
                }
            }
        },
        "messages": [{
            "role": "user",
            "content": "What are the best restaurants around Granary Square?"
        }]
    }'
```

#### Search context size

When using this tool, the `search_context_size` parameter controls how much context is retrieved from the web to help the tool formulate a response. The tokens used by the search tool do **not** affect the context window of the main model specified in the `model` parameter in your response creation request. These tokens are also **not** carried over from one turn to another — they're simply used to formulate the tool response and then discarded.

Choosing a context size impacts:

- **Cost**: Pricing of our search tool varies based on the value of this parameter. Higher context sizes are more expensive. See tool pricing [here](/docs/pricing).
- **Quality**: Higher search context sizes generally provide richer context, resulting in more accurate, comprehensive answers.
- **Latency**: Higher context sizes require processing more tokens, which can slow down the tool's response time.

Available values:

- **`high`**: Most comprehensive context, highest cost, slower response.
- **`medium`** (default): Balanced context, cost, and latency.
- **`low`**: Least context, lowest cost, fastest response, but potentially lower answer quality.

Again, tokens used by the search tool do **not** impact main model's token usage and are not carried over from turn to turn. Check the [pricing page](/docs/pricing) for details on costs associated with each context size.

Context size configuration is not supported for o3, o3-pro, o4-mini, and deep research models.

Customizing search context size

```bash
curl -X POST "https://api.openai.com/v1/chat/completions" \
    -H "Authorization: Bearer $OPENAI_API_KEY" \
    -H "Content-type: application/json" \
    -d '{
        "model": "gpt-4o-search-preview",
        "web_search_options": {
            "search_context_size": "low"
        },
        "messages": [{
            "role": "user",
            "content": "What movie won best picture in 2025?"
        }]
    }'
```

#### Limitations

- Web search is currently not supported in the [`gpt-4.1-nano`](/docs/models/gpt-4.1-nano) model.
- The [`gpt-4o-search-preview`](/docs/models/gpt-4o-search-preview) and [`gpt-4o-mini-search-preview`](/docs/models/gpt-4o-mini-search-preview) models used in Chat Completions only support a subset of API parameters - view their model data pages for specific information on rate limits and feature support.
- When used as a tool in the [Responses API](/docs/api-reference/responses), web search has the same tiered rate limits as the models above.
- Web search is limited to a context window size of 128000 (even with [`gpt-4.1`](/docs/models/gpt-4.1) and [`gpt-4.1-mini`](/docs/models/gpt-4.1-mini) models).
- [Refer to this guide](/docs/guides/your-data) for data handling, residency, and retention information.
