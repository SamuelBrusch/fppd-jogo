package main

import "context" 

func (s *StarBonus) Run(ctx context.Context, out chan<- GameEvent, collected <-chan PlayerCollect){
	
	//sasdasdasdasdas
	
}

type StarBonus struct {
    Value int
}

type GameEvent struct {
    Message string
}

type PlayerCollect struct {
    PlayerID int
}
