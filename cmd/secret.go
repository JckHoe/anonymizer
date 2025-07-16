package main

type Secrets map[string][]string

func (ad Secrets) Merge(other Secrets) {
	for key, values := range other {
		ad[key] = append(ad[key], values...)
	}
}

type SecretSelector interface {
	Select(input string) ([]string, error)
}
