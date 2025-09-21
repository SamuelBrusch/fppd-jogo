// main.go - Loop principal do jogo
package main

import (
	"context"
	"os"
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

	// Iniciar goroutine do monstro se ele existir
	if jogo.Monstro != nil {
		go jogo.Monstro.Run(ctx, jogo.GameEvents, jogo.PlayerAlerts, jogo.PlayerState)
	}

	// Iniciar goroutines dos itens de invisibilidade
	for _, invisItem := range jogo.InvisibilityItems {
		go invisItem.Run(ctx, jogo.GameEvents, jogo.PlayerCollects)
	}

	// Desenha o estado inicial do jogo
	interfaceDesenharJogo(&jogo)

	// Loop principal de entrada
	for {
		evento := interfaceLerEventoTeclado()
		if continuar := personagemExecutarAcao(evento, &jogo); !continuar {
			cancel() // Cancelar goroutine do monstro
			break
		}

		// Processar eventos do monstro
		jogoProcessarEventos(&jogo)

		interfaceDesenharJogo(&jogo)
	}
}
