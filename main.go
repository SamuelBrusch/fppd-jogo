package main

import (
	"os"
)

func main() {
	// Inicializa a interface gráfica
	interfaceIniciar()
	defer interfaceFinalizar()

	// Configura o arquivo do mapa
	mapaFile := "mapa.txt"
	if len(os.Args) > 1 {
		mapaFile = os.Args[1]
	}

	// Inicializa o estado do jogo
	jogo := jogoNovo()
	anciao := &Anciao{}
	
	// Carrega o mapa e os inimigos
    inimigos, startChan, err := jogoCarregarMapa(mapaFile, &jogo, anciao)
    if err != nil {
        panic(err)
    }


	// Inicia as goroutines de movimento
	go anciao.mover(&jogo)
    for _, inimigo := range inimigos {
        go inimigo.mover(&jogo, startChan)
    }

	startChan = make(chan struct{})
	go anciao.verificarEncontro(&jogo, startChan)
	
	for _, inimigo := range inimigos {
		go inimigo.mover(&jogo, startChan)
	}

	// Loop principal do jogo
	for {
		// Atualiza a renderização
		interfaceDesenharJogo(&jogo, anciao, inimigos)
		
		// Processa entrada do jogador
		evento := interfaceLerEventoTeclado()
		
		// Verifica interação com inimigos (tecla 'E')
		if evento.Tipo == "interagir" {
			for _, inimigo := range inimigos {
				if distancia(inimigo.X, inimigo.Y, jogo.PosX, jogo.PosY) <= 2 {
					inimigo.ativo = false
					jogo.Mapa[inimigo.Y][inimigo.X] = Vazio
					jogo.StatusMsg = "Inimigo derrotado!"
				}
			}
			
			// Verifica se encontrou o ancião
			if jogo.Mapa[jogo.PosY][jogo.PosX].simbolo == AnciaoElem.simbolo {
				for _, inimigo := range inimigos {
					inimigo.ativo = true // Ativa todos os inimigos
				}
				jogo.StatusMsg = "Ancião: Os inimigos despertaram!"
			}
		}
		
		// Processa outras ações (movimento, saída)
		if !personagemExecutarAcao(evento, &jogo, inimigos) {
			break
		}
	}
}

// Função auxiliar para calcular distância
func distancia(x1, y1, x2, y2 int) int {
	dx := x1 - x2
	dy := y1 - y2
	return dx*dx + dy*dy // Distância quadrática (mais eficiente que calcular raiz)
}