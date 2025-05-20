package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Account struct {
	ID      int
	Balance int
	mu      sync.Mutex
}

type Bank struct {
	Accounts []*Account
}

func (b *Bank) TransferDeadlock(fromID, toID, amount int) {
	from := b.Accounts[fromID]
	to := b.Accounts[toID]

	from.mu.Lock()
	// Искусственная задержка для увеличения вероятности deadlock
	time.Sleep(time.Microsecond * 100)
	to.mu.Lock()

	from.Balance -= amount
	to.Balance += amount

	to.mu.Unlock()
	from.mu.Unlock()
}

func (b *Bank) TransferCorrect(fromID, toID, amount int) {
	first, second := fromID, toID
	if fromID > toID {
		first, second = second, first
	}

	b.Accounts[first].mu.Lock()
	b.Accounts[second].mu.Lock()

	b.Accounts[fromID].Balance -= amount
	b.Accounts[toID].Balance += amount

	b.Accounts[second].mu.Unlock()
	b.Accounts[first].mu.Unlock()
}

func main() {
	rand.Seed(time.Now().UnixNano())
	const numAccounts = 100
	const numTransactions = 10000

	fmt.Println("=== ДЕМОНСТРАЦИЯ DEADLOCK С 10 000 ГОРУТИН ===")

	//с deadlock
	fmt.Println("Запуск 10 000 небезопасных переводов...")
	bankDeadlock := Bank{}
	for i := 0; i < numAccounts; i++ {
		bankDeadlock.Accounts = append(bankDeadlock.Accounts, &Account{ID: i, Balance: 1000})
	}
	runTransactions(&bankDeadlock, numTransactions, true)

	//без deadlock
	fmt.Println("Запуск 10 000 безопасных переводов...")
	bankCorrect := Bank{}
	for i := 0; i < numAccounts; i++ {
		bankCorrect.Accounts = append(bankCorrect.Accounts, &Account{ID: i, Balance: 1000})
	}
	runTransactions(&bankCorrect, numTransactions, false)

	fmt.Println("Программа завершена!")
}

func runTransactions(bank *Bank, count int, useDeadlock bool) {
	var wg sync.WaitGroup
	success := make(chan bool, 1)

	// Детектор deadlock
	go func() {
		select {
		case <-time.After(5 * time.Second):
			fmt.Println("Обнаружен deadlock!")
			success <- false
		}
	}()

	// Запускаем транзакции
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			from := rand.Intn(len(bank.Accounts))
			to := rand.Intn(len(bank.Accounts))
			if from == to {
				return
			}

			if useDeadlock {
				bank.TransferDeadlock(from, to, 1)
			} else {
				bank.TransferCorrect(from, to, 1)
			}
		}()
	}

	go func() {
		wg.Wait()
		success <- true
	}()

	if <-success {
		fmt.Println("Все переводы завершены успешно!")
	}
}
