// main.go - Loop principal do jogo
package main

import "os"

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
	anciao := &Anciao{}
	if err := jogoCarregarMapa(mapaFile, &jogo, anciao); err != nil {
		panic(err)
	}

	go anciao.mover(&jogo)

	// Loop principal de entrada
	for {
		interfaceDesenharJogo(&jogo, anciao)
		evento := interfaceLerEventoTeclado()
		if continuar := personagemExecutarAcao(evento, &jogo); !continuar {
			break
		}
	}
}