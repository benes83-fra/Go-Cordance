package ecs

type Billboard struct{}

func NewBillboard() *Billboard { return &Billboard{} }

func (b *Billboard) Update(dt float32) { _ = dt }
