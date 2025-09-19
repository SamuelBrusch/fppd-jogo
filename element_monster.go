package main

func (m *Monster) Run(ctx context.Context, out chan<- GameEvent, alerts <-chan PlayerAlert, pstate <-chan PlayerState)
