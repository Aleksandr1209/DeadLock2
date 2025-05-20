package main

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestTransferDeadlock(t *testing.T) {
	bank := Bank{
		Accounts: []*Account{
			{ID: 1, Balance: 1000},
			{ID: 2, Balance: 1000},
		},
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		bank.TransferDeadlock(0, 1, 100) // 1 → 2
	}()

	go func() {
		defer wg.Done()
		bank.TransferDeadlock(1, 0, 50) // 2 → 1
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		t.Error("Ожидался deadlock, но все операции завершились")
	case <-time.After(2 * time.Second):
	}
}

func TestTransferCorrect(t *testing.T) {
	bank := Bank{
		Accounts: []*Account{
			{ID: 1, Balance: 1000},
			{ID: 2, Balance: 1000},
		},
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Запускаем два перевода с правильной блокировкой
	go func() {
		defer wg.Done()
		bank.TransferCorrect(0, 1, 100) // 1 → 2
	}()

	go func() {
		defer wg.Done()
		bank.TransferCorrect(1, 0, 50) // 2 → 1
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Проверяем балансы
		if bank.Accounts[0].Balance != 950 {
			t.Errorf("Неправильный баланс счета 1: got %d, want 950", bank.Accounts[0].Balance)
		}
		if bank.Accounts[1].Balance != 1050 {
			t.Errorf("Неправильный баланс счета 2: got %d, want 1050", bank.Accounts[1].Balance)
		}
	case <-time.After(2 * time.Second):
		t.Error("Обнаружен deadlock в методе, который должен работать корректно")
	}
}

func TestConcurrentTransfers(t *testing.T) {
	bank := Bank{}
	for i := 0; i < 100; i++ {
		bank.Accounts = append(bank.Accounts, &Account{ID: i, Balance: 1000})
	}

	var wg sync.WaitGroup
	totalTransactions := 10000

	// Запускаем множество параллельных переводов
	for i := 0; i < totalTransactions; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			from := rand.Intn(len(bank.Accounts))
			to := rand.Intn(len(bank.Accounts))
			if from != to {
				bank.TransferCorrect(from, to, 1)
			}
		}()
	}

	wg.Wait()

	// Проверяем сохранение общего баланса
	total := 0
	for _, acc := range bank.Accounts {
		total += acc.Balance
	}
	expectedTotal := 100 * 1000 // 100 счетов по 1000
	if total != expectedTotal {
		t.Errorf("Общий баланс изменился: got %d, want %d", total, expectedTotal)
	}
}
