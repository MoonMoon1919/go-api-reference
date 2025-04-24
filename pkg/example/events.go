package example

type ExampleEvent string

const (
	ExampleCreatedEvent ExampleEvent = "ExampleCreated"
	ExampleUpdatedEvent ExampleEvent = "ExampleUpdated"
	ExampleDeletedEvent ExampleEvent = "ExampleDeleted"
)

type ExampleCreated struct {
	Id string
}

func (e ExampleCreated) Name() ExampleEvent {
	return ExampleCreatedEvent
}

func (e ExampleCreated) EntityId() string {
	return e.Id
}

type ExampleUpdated struct {
	Id string
}

func (e ExampleUpdated) Name() ExampleEvent {
	return ExampleUpdatedEvent
}

func (e ExampleUpdated) EntityId() string {
	return e.Id
}

type ExampleDeleted struct {
	Id string
}

func (e ExampleDeleted) Name() ExampleEvent {
	return ExampleDeletedEvent
}

func (e ExampleDeleted) EntityId() string {
	return e.Id
}
