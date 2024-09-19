package main

import (
	"encoding/json"
	"log/slog"
	"math/rand"
	"os"
	"sync"
	"time"
)

type Cowboy struct {
	Name   string `json:"name"`
	Health int    `json:"health"`
	Damage int    `json:"damage"`
}

var cowboysJson = ` 
[
    {
      "name": "John",
      "health": 10,
      "damage": 1
    },
    {
      "name": "Bill",
      "health": 8,
      "damage": 2
    },
    {
      "name": "Sam",
      "health": 10,
      "damage": 1
    },
    {
      "name": "Peter",
      "health": 5,
      "damage": 3
    },
    {
      "name": "Philip",
      "health": 15,
      "damage": 1
    }
]
`

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))
var mutex sync.Mutex
var aliveCowboysCount int

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	err := writeJsonToFile("cowboys.json", cowboysJson, logger)
	if err != nil {
		logger.Error("Error to write JSON in file cowboys", err)
		return
	}

	fileCowboys, err := os.ReadFile("cowboys.json")
	if err != nil {
		logger.Error("Error to read JSON in file cowboys", err)
		return
	}

	var cowboys []Cowboy
	err = json.Unmarshal(fileCowboys, &cowboys)
	if err != nil {
		logger.Error("Error to parse JSON from file cowboys", err)
		return
	}

	mutex.Lock()
	aliveCowboysCount = len(cowboys)
	mutex.Unlock()

	stopChan := make(chan bool)
	var wg sync.WaitGroup

	logger.Info("Fight!!!")
	for i := range cowboys {
		wg.Add(1)
		go cowboyFight(&cowboys, i, stopChan, &wg, logger)
	}

	go func() {
		for {
			mutex.Lock()
			if aliveCowboysCount <= 1 {
				mutex.Unlock()
				close(stopChan)
				return
			}
			mutex.Unlock()
			time.Sleep(1 * time.Second)
		}
	}()

	wg.Wait()

	mutex.Lock()
	defer mutex.Unlock()
	if aliveCowboysCount == 1 {
		for _, cowboy := range cowboys {
			if cowboy.Health > 0 {
				logger.Info("Cowboy", cowboy.Name, "is alive")
			}
		}
	} else {
		logger.Info("Everything cowboys are dead")
	}
}

func cowboyFight(cowboys *[]Cowboy, index int, stopChan chan bool, wg *sync.WaitGroup, logger *slog.Logger) {
	defer wg.Done()
	for {
		select {
		case <-stopChan:
			return
		default:
			time.Sleep(1 * time.Second)
			targetIndex := rng.Intn(len(*cowboys))
			for targetIndex == index {
				targetIndex = rng.Intn(len(*cowboys))
			}

			attack(cowboys, index, targetIndex, logger)
		}
	}
}

func attack(cowboys *[]Cowboy, attackerIndex, targetIndex int, logger *slog.Logger) {
	attacker := &(*cowboys)[attackerIndex]
	target := &(*cowboys)[targetIndex]

	if target.Health <= 0 {
		return
	}

	target.Health -= attacker.Damage

	if target.Health <= 0 {
		mutex.Lock()
		aliveCowboysCount--
		mutex.Unlock()
		logger.Info("Cowboy", target.Name, "is dead")
	}
}

func writeJsonToFile(filename, jsonStr string, logger *slog.Logger) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			logger.Error("Error to close file cowboys", err)
		}
	}(file)

	_, err = file.WriteString(jsonStr)
	if err != nil {
		return err
	}

	return nil
}
