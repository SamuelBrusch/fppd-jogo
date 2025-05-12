package main

import (
    "time"
)

type Anciao struct {
    X, Y      int  
    direcaoX  int  
}

func (a *Anciao) mover(jogo *Jogo) {
    a.direcaoX = 1  
    a.Y = 26        

    for {
        time.Sleep(300 * time.Millisecond)
        jogo.Mutex.Lock()

        if a.Y >= 0 && a.Y < len(jogo.Mapa) && a.X >= 0 && a.X < len(jogo.Mapa[a.Y]) {
            jogo.Mapa[a.Y][a.X] = Vazio
        }

        novaPosX := a.X + a.direcaoX

        if novaPosX >= 0 && novaPosX < len(jogo.Mapa[0]) {
            if !jogo.Mapa[a.Y][novaPosX].tangivel {
                a.X = novaPosX 
            } else {
                a.direcaoX *= -1 
            }
        } else {
            a.direcaoX *= -1 
        }

        if a.Y >= 0 && a.Y < len(jogo.Mapa) && a.X >= 0 && a.X < len(jogo.Mapa[a.Y]) {
            jogo.Mapa[a.Y][a.X] = AnciaoElem
        }

        if a.X == jogo.PosX && a.Y == jogo.PosY {
            jogo.StatusMsg = "Ancião: Você recebeu uma espada!"
        }
        jogo.Mutex.Unlock()
    }
}

func (a *Anciao) verificarEncontro(jogo *Jogo, startChan chan<- struct{}) {
    for {
        time.Sleep(100 * time.Millisecond)
        if jogo.PosX == a.X && jogo.PosY == a.Y {
            close(startChan) // Libera os inimigos
            return
        }
    }
}