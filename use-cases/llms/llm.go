package llms

type ILLM interface {
	Generate(model, prompt string) (string, error)
}
