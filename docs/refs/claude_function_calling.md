# Function Callings/Tool use with Claude

- https://docs.anthropic.com/en/docs/agents-and-tools/tool-use/overview#single-tool-example

## Examples

Here are a few code examples demonstrating various tool use patterns and techniques. For brevity’s sake, the tools are simple tools, and the tool descriptions are shorter than would be ideal to ensure best performance.

### Single tool example

```sh
curl https://api.anthropic.com/v1/messages \
     --header "x-api-key: $ANTHROPIC_API_KEY" \
     --header "anthropic-version: 2023-06-01" \
     --header "content-type: application/json" \
     --data \
'{
    "model": "claude-opus-4-20250514",
    "max_tokens": 1024,
    "tools": [{
        "name": "get_weather",
        "description": "Get the current weather in a given location",
        "input_schema": {
            "type": "object",
            "properties": {
                "location": {
                    "type": "string",
                    "description": "The city and state, e.g. San Francisco, CA"
                },
                "unit": {
                    "type": "string",
                    "enum": ["celsius", "fahrenheit"],
                    "description": "The unit of temperature, either \"celsius\" or \"fahrenheit\""
                }
            },
            "required": ["location"]
        }
    }],
    "messages": [{"role": "user", "content": "What is the weather like in San Francisco?"}]
}'
```

Claude will return a response similar to:

```json
{
  "id": "msg_01Aq9w938a90dw8q",
  "model": "claude-opus-4-20250514",
  "stop_reason": "tool_use",
  "role": "assistant",
  "content": [
    {
      "type": "text",
      "text": "<thinking>I need to call the get_weather function, and the user wants SF, which is likely San Francisco, CA.</thinking>"
    },
    {
      "type": "tool_use",
      "id": "toolu_01A09q90qw90lq917835lq9",
      "name": "get_weather",
      "input": { "location": "San Francisco, CA", "unit": "celsius" }
    }
  ]
}
```

You would then need to execute the get_weather function with the provided input, and return the result in a new user message:

```sh
curl https://api.anthropic.com/v1/messages \
     --header "x-api-key: $ANTHROPIC_API_KEY" \
     --header "anthropic-version: 2023-06-01" \
     --header "content-type: application/json" \
     --data \
'{
    "model": "claude-opus-4-20250514",
    "max_tokens": 1024,
    "tools": [
        {
            "name": "get_weather",
            "description": "Get the current weather in a given location",
            "input_schema": {
                "type": "object",
                "properties": {
                    "location": {
                        "type": "string",
                        "description": "The city and state, e.g. San Francisco, CA"
                    },
                    "unit": {
                        "type": "string",
                        "enum": ["celsius", "fahrenheit"],
                        "description": "The unit of temperature, either \"celsius\" or \"fahrenheit\""
                    }
                },
                "required": ["location"]
            }
        }
    ],
    "messages": [
        {
            "role": "user",
            "content": "What is the weather like in San Francisco?"
        },
        {
            "role": "assistant",
            "content": [
                {
                    "type": "text",
                    "text": "<thinking>I need to use get_weather, and the user wants SF, which is likely San Francisco, CA.</thinking>"
                },
                {
                    "type": "tool_use",
                    "id": "toolu_01A09q90qw90lq917835lq9",
                    "name": "get_weather",
                    "input": {
                        "location": "San Francisco, CA",
                        "unit": "celsius"
                    }
                }
            ]
        },
        {
            "role": "user",
            "content": [
                {
                    "type": "tool_result",
                    "tool_use_id": "toolu_01A09q90qw90lq917835lq9",
                    "content": "15 degrees"
                }
            ]
        }
    ]
}'
```

This will print Claude’s final response, incorporating the weather data:

```json
{
  "id": "msg_01Aq9w938a90dw8q",
  "model": "claude-opus-4-20250514",
  "stop_reason": "stop_sequence",
  "role": "assistant",
  "content": [
    {
      "type": "text",
      "text": "The current weather in San Francisco is 15 degrees Celsius (59 degrees Fahrenheit). It's a cool day in the city by the bay!"
    }
  ]
}
```

### Multiple tools example

You can provide Claude with multiple tools to choose from in a single request. Here’s an example with both a get_weather and a get_time tool, along with a user query that asks for both.

```sh
curl https://api.anthropic.com/v1/messages \
     --header "x-api-key: $ANTHROPIC_API_KEY" \
     --header "anthropic-version: 2023-06-01" \
     --header "content-type: application/json" \
     --data \
'{
    "model": "claude-opus-4-20250514",
    "max_tokens": 1024,
    "tools": [{
        "name": "get_weather",
        "description": "Get the current weather in a given location",
        "input_schema": {
            "type": "object",
            "properties": {
                "location": {
                    "type": "string",
                    "description": "The city and state, e.g. San Francisco, CA"
                },
                "unit": {
                    "type": "string",
                    "enum": ["celsius", "fahrenheit"],
                    "description": "The unit of temperature, either 'celsius' or 'fahrenheit'"
                }
            },
            "required": ["location"]
        }
    },
    {
        "name": "get_time",
        "description": "Get the current time in a given time zone",
        "input_schema": {
            "type": "object",
            "properties": {
                "timezone": {
                    "type": "string",
                    "description": "The IANA time zone name, e.g. America/Los_Angeles"
                }
            },
            "required": ["timezone"]
        }
    }],
    "messages": [{
        "role": "user",
        "content": "What is the weather like right now in New York? Also what time is it there?"
    }]
}'
```

In this case, Claude will most likely try to use two separate tools, one at a time — get_weather and then get_time — in order to fully answer the user’s question. However, it will also occasionally output two tool_use blocks at once, particularly if they are not dependent on each other. You would need to execute each tool and return their results in separate tool_result blocks within a single user message.

### JSON Mode

You can use tools to get Claude produce JSON output that follows a schema, even if you don’t have any intention of running that output through a tool or function.

When using tools in this way:

You usually want to provide a single tool
You should set tool_choice (see Forcing tool use) to instruct the model to explicitly use that tool
Remember that the model will pass the input to the tool, so the name of the tool and description should be from the model’s perspective.
The following uses a record_summary tool to describe an image following a particular format.

```sh
#!/bin/bash
IMAGE_URL="https://upload.wikimedia.org/wikipedia/commons/a/a7/Camponotus_flavomarginatus_ant.jpg"
IMAGE_MEDIA_TYPE="image/jpeg"
IMAGE_BASE64=$(curl "$IMAGE_URL" | base64)

curl https://api.anthropic.com/v1/messages \
     --header "content-type: application/json" \
     --header "x-api-key: $ANTHROPIC_API_KEY" \
     --header "anthropic-version: 2023-06-01" \
     --data \
'{
    "model": "claude-opus-4-20250514",
    "max_tokens": 1024,
    "tools": [{
        "name": "record_summary",
        "description": "Record summary of an image using well-structured JSON.",
        "input_schema": {
            "type": "object",
            "properties": {
                "key_colors": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "properties": {
                            "r": { "type": "number", "description": "red value [0.0, 1.0]" },
                            "g": { "type": "number", "description": "green value [0.0, 1.0]" },
                            "b": { "type": "number", "description": "blue value [0.0, 1.0]" },
                            "name": { "type": "string", "description": "Human-readable color name in snake_case, e.g. \"olive_green\" or \"turquoise\"" }
                        },
                        "required": [ "r", "g", "b", "name" ]
                    },
                    "description": "Key colors in the image. Limit to less than four."
                },
                "description": {
                    "type": "string",
                    "description": "Image description. One to two sentences max."
                },
                "estimated_year": {
                    "type": "integer",
                    "description": "Estimated year that the image was taken, if it is a photo. Only set this if the image appears to be non-fictional. Rough estimates are okay!"
                }
            },
            "required": [ "key_colors", "description" ]
        }
    }],
    "tool_choice": {"type": "tool", "name": "record_summary"},
    "messages": [
        {"role": "user", "content": [
            {"type": "image", "source": {
                "type": "base64",
                "media_type": "'$IMAGE_MEDIA_TYPE'",
                "data": "'$IMAGE_BASE64'"
            }},
            {"type": "text", "text": "Describe this image."}
        ]}
    ]
}'
```
