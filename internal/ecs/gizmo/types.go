package gizmo

type GizmoMode int

const (
	GizmoMove GizmoMode = iota
	GizmoRotate
	GizmoScale
	GizmoCombined
)
