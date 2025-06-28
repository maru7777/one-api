# Structured output 📑

- <https://cloud.google.com/vertex-ai/generative-ai/docs/multimodal/control-generated-output>
- <https://ai.google.dev/gemini-api/docs/structured-output>

**Release Notes**

To see an example of structured output, run the "Intro to structured output" Jupyter notebook in one of the following environments:

[Open in Colab](https://www.google.com/search?q=https://colab.research.google.com/github/google-cloud-vertex-ai/generative-ai/blob/main/gemini/function-calling/intro_structured_output.ipynb) | [Open in Colab Enterprise](https://www.google.com/search?q=https://console.cloud.google.com/colab/welcome) | [Open in Vertex AI Workbench](https://www.google.com/search?q=https://console.cloud.google.com/vertex-ai/workbench/deploy-managed-notebook%3Fq%3Ddownload_url%253Dhttps%25253A%25252F%25252Fraw.githubusercontent.com%25252FGoogleCloudPlatform%25252Fvertex-ai-samples%25252Fmain%25252Fnotebooks%25252Fofficial%25252Fgenerative-ai%25252Fintro_structured_output.ipynb) | [View on GitHub](https://www.google.com/search?q=https://github.com/GoogleCloudPlatform/vertex-ai-samples/blob/main/notebooks/official/generative-ai/intro_structured_output.ipynb)

---

You can guarantee that a model's generated output always adheres to a specific schema so that you receive consistently formatted responses. For example, you might have an established data schema that you use for other tasks. If you have the model follow the same schema, you can directly extract data from the model's output without any post-processing.

To specify the structure of a model's output, define a **response schema**, which works like a blueprint for model responses. When you submit a prompt and include the **response schema**, the model's response always follows your defined schema.

You can control generated output when using the following models:

- Vertex AI Model Optimizer 🔬
- Gemini 1.5 Pro (preview)
- Gemini 1.5 Flash (preview)
- Gemini 1.0 Pro
- Gemini 1.0 Pro Vision

> **Note:** Using structured output on tuned Gemini models can result in decreased model quality.

---

## Example use cases

One use case for applying a response schema is to ensure that a model's response produces **valid JSON** and conforms to your schema. Generative model outputs can have some degree of variability, so including a response schema ensures that you always receive valid JSON. Consequently, your downstream tasks can reliably expect valid JSON input from generated responses.

Another example is to constrain how a model can respond. For example, you can have a model annotate text with user-defined labels, not with labels that the model produces. This constraint is useful when you expect a specific set of labels such as `positive` or `negative` and don't want to receive a mixture of other labels that the model might generate like `good`, `positive`, `negative`, or `bad`.

---

## Considerations

The following considerations discuss potential limitations if you plan on using a response schema:

- You must use the API to define and use a response schema. There's no console support.
- The size of your response schema counts towards the input token limit.
- Only certain output formats are supported, such as `application/json` or `text/x.enum`. For more information, see the `responseMimeType` parameter in the [Gemini API reference](https://cloud.google.com/vertex-ai/docs/generative-ai/model-reference/gemini).
- Structured output supports a subset of the [Vertex AI schema reference](https://www.google.com/search?q=https://cloud.google.com/vertex-ai/docs/reference/rest/v1/Tool%23schema). For more information, see [Supported schema fields](https://www.google.com/search?q=%23supported-schema-fields).
- A complex schema can result in an `InvalidArgument: 400` error. Complexity might come from long property names, long array length limits, enums with many values, objects with lots of optional properties, or a combination of these factors.
  - If you get this error with a valid schema, make one or more of the following changes to resolve the error:
    - Shorten property names or enum names.
    - Flatten nested arrays.
    - Reduce the number of properties with constraints, such as numbers with minimum and maximum limits.
    - Reduce the number of properties with complex constraints, such as properties with complex formats like `date-time`.
    - Reduce the number of optional properties.
    - Reduce the number of valid values for enums.

---

## Supported schema fields

Structured output supports the following fields from the Vertex AI schema. If you use an unsupported field, Vertex AI can still handle your request but ignores the field.

- `anyOf`
- `enum`
- `format`
- `items`
- `maximum`
- `maxItems`
- `minimum`
- `minItems`
- `nullable`
- `properties`
- `propertyOrdering`\*
- `required`

\* **\*propertyOrdering** is specifically for structured output and not part of the Vertex AI schema. This field defines the order in which properties are generated. The listed properties must be unique and must be valid keys in the `properties` dictionary.\*

For the `format` field, Vertex AI supports the following values: `date`, `date-time`, `duration`, and `time`. The description and format of each value is described in the [OpenAPI Initiative Registry](https://www.google.com/search?q=https://spec.openapis.org/oas/v3.0.3%23data-types).

---

## Before you begin

- Define a response schema to specify the structure of a model's output, the field names, and the expected data type for each field. Use only the supported fields as listed in the [Considerations](https://www.google.com/search?q=%23considerations) section. All other fields are ignored.
- Include your response schema as part of the **responseSchema** field only. Don't duplicate the schema in your input prompt. If you do, the generated output might be lower in quality.

For sample schemas, see the [Example schemas and model responses](https://www.google.com/search?q=%23example-schemas-for-json-output) section.

---

## Model behavior and response schema

When a model generates a response, it uses the field name and context from your prompt. As such, we recommend that you use a clear structure and unambiguous field names so that your intent is clear.

By default, fields are optional, meaning the model can populate the fields or skip them. You can set fields as required to force the model to provide a value. If there's insufficient context in the associated input prompt, the model generates responses mainly based on the data it was trained on.

If you aren't seeing the results you expect, add more context to your input prompts or revise your response schema. For example, review the model's response without structured output to see how the model responds. You can then update your response schema that better fits the model's output.

---

## Send a prompt with a response schema

To see an example of a response schema and structured output, run the "Introduction to structured output" Jupyter notebook in one of the following environments:

[Open in Colab](https://www.google.com/search?q=https://colab.research.google.com/github/google-cloud-vertex-ai/generative-ai/blob/main/gemini/function-calling/intro_structured_output.ipynb) | [Open in Colab Enterprise](https://www.google.com/search?q=https://console.cloud.google.com/colab/welcome) | [Open in Vertex AI Workbench](https://www.google.com/search?q=https://console.cloud.google.com/vertex-ai/workbench/deploy-managed-notebook%3Fq%3Ddownload_url%253Dhttps%25253A%25252F%25252Fraw.githubusercontent.com%25252FGoogleCloudPlatform%25252Fvertex-ai-samples%25252Fmain%25252Fnotebooks%25252Fofficial%25252Fgenerative-ai%25252Fintro_structured_output.ipynb) | [View on GitHub](https://www.google.com/search?q=https://github.com/GoogleCloudPlatform/vertex-ai-samples/blob/main/notebooks/official/generative-ai/intro_structured_output.ipynb)

By default, all fields are optional, meaning a model might generate a response to a field. To force the model to always generate a response to a field, set the field as required.

### Gen AI SDK for Python

(See examples below)

### Gen AI SDK for Go

(Examples available in product documentation)

### REST

Before using any of the request data, make the following replacements:

- `GENERATE_RESPONSE_METHOD`: The type of response that you want the model to generate. Choose `streamGenerateContent` for a streamed response or `generateContent` for a full response.
- `LOCATION`: The region to process the request.
- `PROJECT_ID`: Your project ID.
- `MODEL_ID`: The model ID of the multimodal model that you want to use.
- `ROLE`: The role in a conversation associated with the content. Required. Use `USER` for your content.
- `TEXT`: The text instructions to include in the prompt.
- `RESPONSE_MIME_TYPE`: The format type of the generated candidate text (e.g., `application/json`).
- `RESPONSE_SCHEMA`: Schema for the model to follow when generating responses.

#### HTTP method and URL:

```
POST https://LOCATION-aiplatform.googleapis.com/v1/projects/PROJECT_ID/locations/LOCATION/publishers/google/models/MODEL_ID:GENERATE_RESPONSE_METHOD
```

#### Request JSON body:

```json
{
  "contents": {
    "role": "ROLE",
    "parts": {
      "text": "TEXT"
    }
  },
  "generation_config": {
    "responseMimeType": "RESPONSE_MIME_TYPE",
    "responseSchema": RESPONSE_SCHEMA,
  }
}
```

To send your request, choose one of these options:

#### `curl`

> **Note:** The following command assumes that you have logged in to the `gcloud` CLI with your user account by running `gcloud init` or `gcloud auth login`, or by using Cloud Shell.

Save the request body in a file named `request.json`, and execute the following command:

```bash
curl -X POST \
     -H "Authorization: Bearer $(gcloud auth print-access-token)" \
     -H "Content-Type: application/json; charset=utf-8" \
     -d @request.json \
     "https://LOCATION-aiplatform.googleapis.com/v1/projects/PROJECT_ID/locations/LOCATION/publishers/google/models/MODEL_ID:GENERATE_RESPONSE_METHOD"
```

#### Example `curl` command

```bash
LOCATION="us-central1"
MODEL_ID="gemini-1.0-pro"
PROJECT_ID="test-project"
GENERATE_RESPONSE_METHOD="generateContent"

cat << EOF > request.json
{
  "contents": {
    "role": "user",
    "parts": {
      "text": "List a few popular cookie recipes."
    }
  },
  "generation_config": {
    "maxOutputTokens": 2048,
    "responseMimeType": "application/json",
    "responseSchema": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "recipe_name": {
            "type": "string"
          }
        },
        "required": ["recipe_name"]
      }
    }
  }
}
EOF

curl \
-X POST \
-H "Authorization: Bearer $(gcloud auth print-access-token)" \
-H "Content-Type: application/json" \
https://${LOCATION}-aiplatform.googleapis.com/v1/projects/${PROJECT_ID}/locations/${LOCATION}/publishers/google/models/${MODEL_ID}:${GENERATE_RESPONSE_METHOD} \
-d '@request.json'
```

---

## Example schemas for JSON output

The following sections demonstrate a variety of sample prompts and response schemas.

- [Forecast the weather for each day of the week in an array](https://www.google.com/search?q=%23forecast-the-weather-for-each-day-of-the-week)
- [Classify a product with a well-defined enum](https://www.google.com/search?q=%23classify-a-product)

### Forecast the weather for each day of the week

The following example outputs a `forecast` object for each day of the week that includes an array of properties such as the expected temperature and humidity level. Some properties are set to `nullable` so the model can return a null value when it doesn't have enough context, which helps reduce hallucinations.

#### Gen AI SDK for Python

**Install**

```bash
pip install --upgrade google-cloud-aiplatform
```

**Set environment variables**

```bash
# Replace with your project and location
export GOOGLE_CLOUD_PROJECT=your-gcp-project
export GOOGLE_CLOUD_LOCATION=us-central1
```

**Code**

```python
import vertexai
from vertexai.generative_models import GenerativeModel, Part

# Initialize Vertex AI
vertexai.init(project=GOOGLE_CLOUD_PROJECT, location=GOOGLE_CLOUD_LOCATION)

response_schema = {
    "type": "object",
    "properties": {
        "forecast": {
            "type": "array",
            "items": {
                "type": "object",
                "properties": {
                    "Day": {"type": "string"},
                    "Forecast": {"type": "string"},
                    "Temperature": {"type": "integer"},
                    "Humidity": {"type": "string", "nullable": True},
                    "Wind Speed": {"type": "integer", "nullable": True},
                },
                "required": ["Day", "Temperature", "Forecast"],
            },
        }
    },
}

prompt = """
The week ahead brings a mix of weather conditions.
Sunday is expected to be sunny with a temperature of 77°F and a humidity level of 50%. Winds will be light at around 10 km/h.
Monday will see partly cloudy skies with a slightly cooler temperature of 72°F and the winds will pick up slightly to around 15 km/h.
Tuesday brings rain showers, with temperatures dropping to 64°F and humidity rising to 70%.
Wednesday may see thunderstorms, with a temperature of 68°F.
Thursday will be cloudy with a temperature of 66°F and moderate humidity at 60%.
Friday returns to partly cloudy conditions, with a temperature of 73°F and the Winds will be light at 12 km/h.
Finally, Saturday rounds off the week with sunny skies, a temperature of 80°F, and a humidity level of 40%. Winds will be gentle at 8 km/h.
"""

# Load the model
model = GenerativeModel("gemini-1.0-pro")

# Generate content
response = model.generate_content(
    [prompt],
    generation_config={
        "response_mime_type": "application/json",
        "response_schema": response_schema,
    },
)

print(response.text)

# Example output:
# {"forecast": [{"Day": "Sunday", "Forecast": "sunny", "Temperature": 77, "Wind Speed": 10, "Humidity": "50%"},
#   {"Day": "Monday", "Forecast": "partly cloudy", "Temperature": 72, "Wind Speed": 15},
#   {"Day": "Tuesday", "Forecast": "rain showers", "Temperature": 64, "Wind Speed": null, "Humidity": "70%"},
#   {"Day": "Wednesday", "Forecast": "thunderstorms", "Temperature": 68, "Wind Speed": null},
#   {"Day": "Thursday", "Forecast": "cloudy", "Temperature": 66, "Wind Speed": null, "Humidity": "60%"},
#   {"Day": "Friday", "Forecast": "partly cloudy", "Temperature": 73, "Wind Speed": 12},
#   {"Day": "Saturday", "Forecast": "sunny", "Temperature": 80, "Wind Speed": 8, "Humidity": "40%"}]}
```

### Classify a product

The following example includes `enums` where the model must classify an object's type from a list of given values.

#### Gen AI SDK for Python

**Install**

```bash
pip install --upgrade google-cloud-aiplatform
```

**Set environment variables**

```bash
# Replace with your project and location
export GOOGLE_CLOUD_PROJECT=your-gcp-project
export GOOGLE_CLOUD_LOCATION=us-central1
```

**Code**

```python
import vertexai
from vertexai.generative_models import GenerativeModel, Part

# Initialize Vertex AI
vertexai.init(project=GOOGLE_CLOUD_PROJECT, location=GOOGLE_CLOUD_LOCATION)

# Load the model
model = GenerativeModel("gemini-1.0-pro")

# Generate content
response = model.generate_content(
    "What type of instrument is an oboe?",
    generation_config={
        "response_mime_type": "application/json",
        "response_schema": {
            "type": "object",
            "properties": {
              "instrument_type": {
                "type": "string",
                "enum": ["Percussion", "String", "Woodwind", "Brass", "Keyboard"],
              }
            }
        },
    },
)

print(response.text)

# Example output:
# {"instrument_type": "Woodwind"}
```
