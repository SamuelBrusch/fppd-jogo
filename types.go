// types.go - Definições de tipos para elementos especiais
package main

// Posição genérica (reutilizável)
type Position struct {
	X, Y int
}

// Estados do monstro
type MonsterState int

const (
	Hunting    MonsterState = iota // Perseguindo jogador
	Patrolling                     // Patrulhando área
)

// Structs dos elementos especiais
type Monster struct {
	current_position Position     // Posição atual do monster
	shift_count      int          // Contador para movimento a cada 2 turnos
	destiny_position Position     // Posição de destino (patrulha)
	last_seen        Position     // Última posição vista do jogador
	state            MonsterState // Estado atual (hunting/patrolling)
	id               string       // ID único do monster
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

// Dados específicos para movimento do monster
type MonsterMoveData struct {
	OldX, OldY int    // Posição anterior
	NewX, NewY int    // Nova posição desejada
	MonsterID  string // ID do monster
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
