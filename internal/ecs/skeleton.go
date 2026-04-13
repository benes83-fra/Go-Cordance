package ecs

type Skeleton struct {
	Nodes []*Entity // index = glTF node index
}

func (s *Skeleton) Update(dt float32) { _ = dt }

func (s *Skeleton) EditorName() string { return "Skeleton" }

func (s *Skeleton) EditorFields() map[string]any {
	return map[string]any{
		"NodeCount": len(s.Nodes),
	}
}
