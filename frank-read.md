# About advantage
## package of http
Functional programming ideas is good idea in projects all over the world. That programmer use functional programing idea to solve problems is easy in main case. In the project, the thought of `Optional` offers a whole new approach.

However, we must to think a problem what the reason of that we use functional expression? Is it that use for use's sake? I think that the main purpose of functional programming is to make the environment independent and constant quantized.

so, about the code, I do not consider that it is a good design.

e.g:
```go
func NewClient(options ...Option) *Client {
	for _, option := range options {
		option(Client)
	}
	return Client
}
```
