// main.go - Loop principal do jogo
package main

import (
	"context"
	"os"
	"time"
)

func main() {
	// Inicializa a interface (termbox)
	interfaceIniciar()
	defer interfaceFinalizar()

	// Usa "mapa.txt" como arquivo padrão ou lê o primeiro argumento
	mapaFile := "mapa.txt"
	if len(os.Args) > 1 {
		mapaFile = os.Args[1]
	}

	// Inicializa o jogo
	jogo := jogoNovo()
	if err := jogoCarregarMapa(mapaFile, &jogo); err != nil {
		panic(err)
	}

	// Criar contexto para controle das goroutines
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Criar canais de comunicação
	playerStateChannel := make(chan PlayerState, 10)
	monsterEventChannel := make(chan GameEvent, 10)
	alertChannel := make(chan PlayerAlert, 10)

	// Criar e inicializar monstro
	monster := &Monster{
		current_position: Position{X: 50, Y: 18}, // Posição inicial do monstro
		state:            Patrolling,
		shift_count:      0,
	}
	monster.generateRandomDestiny()

	// Iniciar goroutine do monstro
	go monster.Run(ctx, monsterEventChannel, alertChannel, playerStateChannel)

	// Desenha o estado inicial do jogo
	interfaceDesenharJogo(&jogo)

	// Loop principal com comunicação concorrente
	for {
		select {
		case evento := <-interfaceLerEventoTecladoAsync():
			// Processar entrada do jogador
			if continuar := personagemExecutarAcaoComCanal(evento, &jogo, playerStateChannel); !continuar {
				cancel() // Cancelar todas as goroutines
				return
			}
			interfaceDesenharJogo(&jogo)

		case monsterEvent := <-monsterEventChannel:
			// Processar eventos do monstro
			processarEventoMonstro(monsterEvent, &jogo)
			interfaceDesenharJogo(&jogo)

		case <-time.After(50 * time.Millisecond):
			// Timeout para manter o jogo responsivo
			continue
		}
	}
}
