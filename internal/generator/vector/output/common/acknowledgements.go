package common

type Acknowledgments struct {
	ID      string
	Enabled bool
}

func NewAcknowledgments(id string, s ConfigStrategy) Acknowledgments {
	a := Acknowledgments{
		ID: id,
	}
	if s == nil {
		return a
	}
	return s.VisitAcknowledgements(a)
}

func (a Acknowledgments) Name() string {
	return "acknowledgements"
}

func (a Acknowledgments) Template() string {
	if !a.Enabled {
		return `{{define "` + a.Name() + `" -}}{{end}}`
	}
	return `{{define "` + a.Name() + `" -}}
[sinks.{{.ID}}.acknowledgements]
enabled = {{.Enabled}}
{{end}}`
}
