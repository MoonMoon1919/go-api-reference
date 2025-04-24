package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

const EMPTY_STRING string = ""

type Configurator interface {
	Get() (string, error)
	Must() string
}

// Mark: Custom source
type CustomSource[T any] struct {
	f func() (T, error)
}

func (t CustomSource[T]) Get() (T, error) {
	return t.f()
}

func (t CustomSource[T]) Must() T {
	v, err := t.f()

	if err != nil {
		panic(err)
	}

	return v
}

// MARK: Default Val
type DefaultValueSource struct {
	value string
}

func NewDefaultValueSource(value string) DefaultValueSource {
	return DefaultValueSource{
		value: value,
	}
}

func (d DefaultValueSource) Get() (string, error) {
	return d.value, nil
}

func (d DefaultValueSource) Must() string {
	val, _ := d.Get()
	return val
}

// MARK: EnvVar
type EnvironmentSource struct {
	name string
}

func NewEnvironmentSource(name string) EnvironmentSource {
	return EnvironmentSource{
		name: name,
	}
}

func (e EnvironmentSource) Get() (string, error) {
	val, ok := os.LookupEnv(e.name)

	if !ok {
		return EMPTY_STRING, errors.New(fmt.Sprintf("cannot find environment variable %s", e.name))
	}

	return val, nil
}

func (e EnvironmentSource) Must() string {
	v, err := e.Get()

	if err != nil {
		panic(err)
	}

	return v
}

// MARK: HTTP
type httpResponse struct {
	Value string `json:"Value"`
}

type HttpSource struct {
	name    string
	baseUrl string
}

func (h HttpSource) Get() (string, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", h.baseUrl, h.name), nil)
	if err != nil {
		return EMPTY_STRING, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return EMPTY_STRING, err
	}

	var parsedResponse httpResponse
	if err = json.NewDecoder(resp.Body).Decode(&parsedResponse); err != nil {
		return EMPTY_STRING, err
	}

	return parsedResponse.Value, nil
}

func (h HttpSource) Must() string {
	v, err := h.Get()

	if err != nil {
		panic(err)
	}

	return v
}

type First struct {
	sources []Configurator
}

func NewFirst(configurators ...Configurator) *First {
	return &First{
		sources: configurators,
	}
}

func (f First) Get() (string, error) {
	errs := make([]error, 0)

	for _, source := range f.sources {
		v, err := source.Get()
		if err == nil {
			return v, nil
		}

		errs = append(errs, err)
	}

	return EMPTY_STRING, fmt.Errorf("No values found, errors: %w", errors.Join(errs...))
}

func (f First) Must() string {
	v, err := f.Get()

	if err != nil {
		panic(err)
	}

	return v
}
