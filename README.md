# Go ReAct

This is not an officially supported Google product.

Go implementation of Reason+Act based on Google's
[blog](https://ai.googleblog.com/2022/11/react-synergizing-reasoning-and-acting.html)
post.

This library provides several components to chain together to help integrate
with Large Language Models (LLMs). It does so in a way that allows the using
code to be type safe.

## Prompters

Prompters are used to generate a prompt and LLM parameters to send to the LLM.
The normal one to use is `prompters.NewTextTemplate`. This uses the
[`text/template`](https://pkg.go.dev/text/template) package to generate the
prompt. The given `TPrompt` is passed to the template when it's executed. The
output is then retruned. The `TLLMParams` are passed through so that the LLM
can use them.

```
type Data struct {
  Product string
}
data := Data{Product: "Gophers"}
llmParams := 3

tmpl := "Come up with store names that sell {{.Product}}!"
t := prompters.NewTextTemplate[Data, int](tmpl, llmParams)

prompt, params, err := t.Hydrate(context.Background(), data)
if err != nil {
  panic(err)
}
fmt.Printf("Prompt: %s\nParams: %d\n", prompt, params)

// Output:
// Prompt: Come up with store names that sell Gophers!
// Params: 3
```

## Parsers

Parsers are used to parse the output of an LLM. The normal one to use is
`predictors.NewJSONParser`. This expects the output of the LLM to be JSON. It
uses the given type to decode into.

```
type Data struct {
  Name string `json:"name"`
  Age  int    `json:"age"`
}
p := parsers.NewJSONParser[Data]()
person, err := p.Parse(`{"name": "John", "age": 30}`)
if err != nil {
  panic(err)
}
fmt.Printf("%+v\n", person)

// Output:
// {Name:John Age:30}
```

## Predictors

Predictors are a wrapper around an LLM. It is used to predict output (`TResp`)
based on the given `TReq`. It takes an `llms.LLM`, a `prompters.Prompter` and
a `parsers.Parser`. Predictors are often chained together to add
functionality.

```
// This example demonstrates how to chain multiple predictors together.
// The fake one will always return an error indicating that the prediction
// from the LLM failed for some reason. The JSONLogger will log the request
// while the retrier will retry the request 3 times before giving up.

var predictor predictors.Predictor[int, string] = &predictorstesting.Fake[int, string]{
	Err: fmt.Errorf("%w: some-error", predictors.ErrLLM),
}
predictor = predictors.NewJSONLogger(predictor, os.Stdout)
predictor = predictors.NewRetrier(predictor)

predictor.Predict(context.Background(), 1)

// Output:
// {"request":1}
// {"request":1}
// {"request":1}
```

## Agents

Agents are a component that allow the configured LLM to decide which tools to
use to achieve a goal. An Agent has an output type that is returned as the
FinalAnswer.

### Tools

A tool is invoked when its name is returned by the LLM. It should return an
observation based on the output. If the input into a tool is invalid, then the
`tools.ErrInvalidToolInput` error should wrapped and returned.

If the tool returns an error, it is changed into an observation that is given
to the LLM with a `ERROR: ` prefix.

It is a common pattern to build a tool as an Agent. This allows a hierarchy of
Agents and allows more tools to be used with the LLM.

### app-editor example

The app-editor example demonstrates setting up a tool set and Agent. This
example allows a user to interact with the AI to build up a simple app
structure.
