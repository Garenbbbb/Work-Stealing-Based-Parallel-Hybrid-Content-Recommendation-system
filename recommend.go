package main

import (
	"Work-Stealing-Based-Parallel-Hybrid-Content-Recommendation-system/deque"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"sync"
)

type User struct {
	ID          string
	Preferences map[string]float64
}

// Recommendation struct represents the final recommendation for a user
type Recommendation struct {
	UserID      string
	Recommended []string
}

type Worker struct {
	ID    int
	Deque deque.Deque
}

// WorkStealingScheduler represents a work-stealing scheduler with multiple workers
type WorkStealingScheduler struct {
	Workers []Worker
}

type result struct {
	ID              string
	SimilarityScore float64
}

// Create a new worker with a unique ID
func newWorker(id int) Worker {
	return Worker{ID: id}
}

func contains(items *[]deque.Content, itemID string) bool {
	for _, item := range *items {
		if item.ID == itemID {
			return true
		}
	}
	return false
}

func CartFeatures(cart *deque.ShopingCart) map[string]float64 {
	cartFeatures := make(map[string]float64)
	featureCounts := make(map[string]int)

	for _, item := range cart.Items {
		for key, value := range item.Features {
			cartFeatures[key] += value
			featureCounts[key] += 1
		}
	}
	for key := range cartFeatures {
		cartFeatures[key] /= float64(featureCounts[key])
	}
	return cartFeatures
}

func Similarity(features1 map[string]float64, features2 map[string]float64) float64 {
	dotProduct := 0.0
	magnitude1 := 0.0
	magnitude2 := 0.0

	for key := range features1 {
		dotProduct += features1[key] * features2[key]
		magnitude1 += math.Pow(features1[key], 2)
		magnitude2 += math.Pow(features2[key], 2)
	}

	magnitude1 = math.Sqrt(magnitude1)
	magnitude2 = math.Sqrt(magnitude2)

	if magnitude1 == 0 || magnitude2 == 0 {
		return 0.0
	}

	return dotProduct / (magnitude1 * magnitude2)
}

func FindTopSimilarItems(user deque.ShopingCart, content *[]deque.Content, topN int) []result {
	var similarItems []result

	user_feature := CartFeatures(&user)
	// Iterate over items
	for _, item := range *content {
		// Check if the item is not in the user's cart
		if !contains(&user.Items, item.ID) {
			// Compute similarity
			similarity := Similarity(user_feature, item.Features)
			// Append the item with its similarity score
			similarItems = append(similarItems, result{ID: item.ID, SimilarityScore: similarity})
		}
	}
	// Sort similarItems by similarity in descending order
	sort.Slice(similarItems, func(i, j int) bool {
		return similarItems[i].SimilarityScore > similarItems[j].SimilarityScore
	})
	// Return the top N items
	return similarItems[:topN]
}

func processTask(task deque.Task, workerId int, content *[]deque.Content) string {

	topSimilarItems := FindTopSimilarItems(task.Task, content, 3)
	for _, item := range topSimilarItems {
		fmt.Println(item.ID)
	}
	// Actual task processing logic goes here
	return strconv.Itoa(workerId) + "Processing Task" + strconv.Itoa(task.ID)
}

// WorkStealingScheduler runs the work-stealing algorithm
func (ws *WorkStealingScheduler) Run(result chan string, tasks *[]deque.Task, content *[]deque.Content) {

	// Distribute tasks to workers
	for i, task := range *tasks {
		worker := &ws.Workers[i%len(ws.Workers)]
		worker.Deque.PushBack(task)
	}

	// Simulate workers processing tasks
	var wg sync.WaitGroup
	for i := range ws.Workers {
		wg.Add(1)
		go func(worker *Worker) {
			defer wg.Done()
			workerProcessTasks(worker, ws, result, content)
		}(&ws.Workers[i])
	}

	// Wait for all workers to finish
	wg.Wait()
}

// workerProcessTasks simulates a worker processing tasks
func workerProcessTasks(worker *Worker, workerPool *WorkStealingScheduler, result chan string, content *[]deque.Content) {
	for {
		task, ok := worker.Deque.PopFront()
		if !ok {
			// No tasks in own deque, try stealing from other workers
			otherWorkerIndex := rand.Intn(len(workerPool.Workers))

			otherWorker := &workerPool.Workers[otherWorkerIndex]
			if otherWorker.ID != worker.ID {
				stolenTask, ok := otherWorker.Deque.PopBack()
				if ok {
					worker.Deque.PushFront(stolenTask)
				}
			}
		} else {
			// Process the task
			result <- processTask(task, worker.ID, content)
		}
	}
}

func main() {

	contents := []deque.Content{
		{ID: "Item1", Features: map[string]float64{"Feature1": 0.2, "Feature2": 0.8}},
		{ID: "Item2", Features: map[string]float64{"Feature1": 0.1, "Feature2": 0.3}},
		{ID: "Item3", Features: map[string]float64{"Feature1": 0.2, "Feature2": 0.1}},
		{ID: "Item4", Features: map[string]float64{"Feature1": 0.3, "Feature2": 0.2}},
		{ID: "Item5", Features: map[string]float64{"Feature1": 0.6, "Feature2": 0.2}},
		{ID: "Item6", Features: map[string]float64{"Feature1": 0.1, "Feature2": 0.5}},
		{ID: "Item7", Features: map[string]float64{"Feature1": 0.2, "Feature2": 0.0}},
		{ID: "Item8", Features: map[string]float64{"Feature1": 0.5, "Feature2": 0.5}},
		{ID: "Item9", Features: map[string]float64{"Feature1": 0.5, "Feature2": 0.5}},
		{ID: "Item10", Features: map[string]float64{"Feature1": 0, "Feature2": 0.4}},
		// Add more content items
	}

	taskPool := []deque.ShopingCart{
		{ID: "User1", Items: []deque.Content{{ID: "Item1", Features: map[string]float64{"Feature1": 0.2, "Feature2": 0.8}}}},
		{ID: "User2", Items: []deque.Content{{ID: "Item2", Features: map[string]float64{"Feature1": 0.5, "Feature2": 0.1}}}},
		{ID: "User3", Items: []deque.Content{{ID: "Item3", Features: map[string]float64{"Feature1": 0.6, "Feature2": 0.3}}}},
		{ID: "User4", Items: []deque.Content{{ID: "Item4", Features: map[string]float64{"Feature1": 0.1, "Feature2": 0.5}}}},
		{ID: "User5", Items: []deque.Content{{ID: "Item5", Features: map[string]float64{"Feature1": 0.4, "Feature2": 0.3}}}},
		{ID: "User6", Items: []deque.Content{{ID: "Item6", Features: map[string]float64{"Feature1": 0.1, "Feature2": 0.9}}}},
		{ID: "User7", Items: []deque.Content{{ID: "Item7", Features: map[string]float64{"Feature1": 0.5, "Feature2": 0.3}}}},
		{ID: "User8", Items: []deque.Content{{ID: "Item8", Features: map[string]float64{"Feature1": 0.1, "Feature2": 0.0}}}},
		{ID: "User9", Items: []deque.Content{{ID: "Item2", Features: map[string]float64{"Feature1": 0.6, "Feature2": 0.3}}}},
		{ID: "User10", Items: []deque.Content{{ID: "Item3", Features: map[string]float64{"Feature1": 0.2, "Feature2": 0.7}}}},
		{ID: "User11", Items: []deque.Content{{ID: "Item5", Features: map[string]float64{"Feature1": 0.1, "Feature2": 0.3}}}},
		{ID: "User12", Items: []deque.Content{{ID: "Item1", Features: map[string]float64{"Feature1": 0.1, "Feature2": 0.1}}}},
		{ID: "User13", Items: []deque.Content{{ID: "Item5", Features: map[string]float64{"Feature1": 0.5, "Feature2": 0.3}}}},
		{ID: "User14", Items: []deque.Content{{ID: "Item7", Features: map[string]float64{"Feature1": 0.1, "Feature2": 0.8}}}},
		{ID: "User15", Items: []deque.Content{{ID: "Item9", Features: map[string]float64{"Feature1": 0.1, "Feature2": 0.4}}}},
	}

	result := make(chan string)
	workers := 4
	workerPool := WorkStealingScheduler{}
	taskCount := len(taskPool)
	tasks := make([]deque.Task, taskCount)
	for i := 0; i < taskCount; i++ {
		tasks[i] = deque.Task{ID: i + 1, Task: taskPool[i]}
	}
	for i := 0; i < workers; i++ {
		workerPool.Workers = append(workerPool.Workers, newWorker(i))
	}
	go workerPool.Run(result, &tasks, &contents)

	for i := 0; i < taskCount; i++ {
		fmt.Println(<-result)
	}

	close(result)

	// users := []User{
	// 	{ID: "User1", Preferences: map[string]float64{"Item1": 5, "Item2": 3, "Item3": 2, "Item4": 5, "Item5": 5, "Item6": 5, "Item7": 2, "Item8": 5}},
	// 	{ID: "User2", Preferences: map[string]float64{"Item1": 3, "Item2": 2, "Item3": 5, "Item4": 1, "Item5": 5, "Item6": 3, "Item7": 1, "Item8": 2}},
	// 	{ID: "User3", Preferences: map[string]float64{"Item1": 2, "Item2": 4, "Item3": 3, "Item4": 4, "Item5": 2, "Item6": 5, "Item7": 6, "Item8": 1}},
	// 	{ID: "User4", Preferences: map[string]float64{"Item1": 6, "Item2": 2, "Item3": 6, "Item4": 5, "Item5": 2, "Item6": 2, "Item7": 1, "Item8": 5}},
	// 	{ID: "User5", Preferences: map[string]float64{"Item1": 1, "Item2": 5, "Item3": 1, "Item4": 2, "Item5": 5, "Item6": 5, "Item7": 2, "Item8": 4}},
	// 	{ID: "User6", Preferences: map[string]float64{"Item1": 2, "Item2": 6, "Item3": 2, "Item4": 5, "Item5": 4, "Item6": 6, "Item7": 2, "Item8": 5}},
	// 	{ID: "User7", Preferences: map[string]float64{"Item1": 0, "Item2": 8, "Item3": 5, "Item4": 0, "Item5": 5, "Item6": 0, "Item7": 5, "Item8": 0}},
	// 	{ID: "User8", Preferences: map[string]float64{"Item1": 3, "Item2": 6, "Item3": 1, "Item4": 5, "Item5": 6, "Item6": 0, "Item7": 2, "Item8": 5}},
	// 	{ID: "User9", Preferences: map[string]float64{"Item1": 5, "Item2": 3, "Item3": 2, "Item5": 1, "Item6": 5, "Item7": 4, "Item8": 5}},
	// 	// Add more users
	// }

	// shopping_card := []ShopingCart{
	// 	{ID: "User1", Items: []string{"Item1", "Item2"}},
	// 	{ID: "User2", Items: []string{"Item3", "Item4"}},
	// }

	// contentAnalysisCh := make(chan map[string]map[string]float64)
	// userAnalysisCh := make(chan map[string]map[string]float64)

}

func euclideanDistance(features1, features2 map[string]float64) float64 {
	var sum float64
	for key := range features1 {
		sum += math.Pow(features1[key]-features2[key], 2)
	}
	return math.Sqrt(sum)
}
