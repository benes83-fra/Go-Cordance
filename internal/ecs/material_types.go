package ecs

type MaterialType int32

const (
	MaterialBlinnPhong MaterialType = 0
	MaterialPBR        MaterialType = 1
	MaterialToon       MaterialType = 2
)
