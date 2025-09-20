// types.go - Definições de tipos para elementos especiais
package main

// Structs dos elementos especiais
type Monster struct {
	X, Y int // Posição do monster
	// Adicionar outros campos conforme necessário
}

type StarBonus struct {
	X, Y int // Posição da estrela
	// Adicionar outros campos conforme necessário
}

type Invisibility struct {
	X, Y int // Posição do item de invisibilidade
	// Adicionar outros campos conforme necessário
}

// Tipos de eventos e comunicação
type GameEvent struct {
	Type string      // Tipo do evento
	Data interface{} // Dados do evento
}

type PlayerAlert struct {
	Type string      // Tipo do alerta
	Data interface{} // Dados do alerta
}

type PlayerState struct {
	X, Y int // Posição do jogador
	// Adicionar outros campos conforme necessário
}

type PlayerCollect struct {
	X, Y int // Posição onde o jogador coletou algo
	// Adicionar outros campos conforme necessário
}
