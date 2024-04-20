package aiproxy

import "github.com/Laisky/one-api/relay/adaptor/openai"

var ModelList = []string{""}

func init() {
	ModelList = openai.ModelList
}
