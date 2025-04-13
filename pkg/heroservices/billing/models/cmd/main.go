package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/heroservices/billing/models"
)

const (
	numUsers        = 1000
	numAccounts     = 3000  // Average 3 accounts per user
	numTransactions = 10000 // Reduced for quicker testing
	currencies      = "USD,EUR,GBP,JPY,CNY,AUD,CAD,CHF,HKD,SGD"
	accountTypes    = "Checking,Savings,Investment,Retirement,Business,Credit,Loan,Mortgage,Travel,Emergency"
	companies       = "Apple,Google,Microsoft,Amazon,Facebook,Tesla,Walmart,Target,Costco,HomeDepot,Lowes,BestBuy,Starbucks,McDonalds,Chipotle,Nike,Adidas,Uber,Lyft,Airbnb"
	firstNames      = "John,Jane,Michael,Emily,David,Sarah,Robert,Jennifer,William,Elizabeth,James,Linda,Richard,Barbara,Joseph,Susan,Thomas,Jessica,Charles,Mary"
	lastNames       = "Smith,Johnson,Williams,Jones,Brown,Davis,Miller,Wilson,Moore,Taylor,Anderson,Thomas,Jackson,White,Harris,Martin,Thompson,Garcia,Martinez,Robinson"
)

var (
	currencyList  []string
	accountList   []string
	companyList   []string
	firstNameList []string
	lastNameList  []string
)

func init() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Split the constant strings into slices
	currencyList = strings.Split(currencies, ",")
	accountList = strings.Split(accountTypes, ",")
	companyList = strings.Split(companies, ",")
	firstNameList = strings.Split(firstNames, ",")
	lastNameList = strings.Split(lastNames, ",")
}

// generateRandomName generates a random user name
func generateRandomName() string {
	firstName := firstNameList[rand.Intn(len(firstNameList))]
	lastName := lastNameList[rand.Intn(len(lastNameList))]
	// Add timestamp and random number to ensure uniqueness
	return fmt.Sprintf("%s_%s_%d_%d", firstName, lastName, time.Now().UnixNano()%1000000, rand.Intn(1000))
}

// generateRandomAccountName generates a random account name
func generateRandomAccountName() string {
	accountType := accountList[rand.Intn(len(accountList))]

	// 30% chance to add a company name
	if rand.Intn(100) < 30 {
		company := companyList[rand.Intn(len(companyList))]
		return fmt.Sprintf("%s %s", company, accountType)
	}

	return accountType
}

// generateRandomCurrency returns a random currency code
func generateRandomCurrency() string {
	return currencyList[rand.Intn(len(currencyList))]
}

// generateRandomAmount generates a random amount between min and max
func generateRandomAmount(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// generateRandomComment generates a realistic transaction comment
func generateRandomComment(fromAccount, toAccount *models.Account, amount float64) string {
	commentTypes := []string{
		"Payment",
		"Transfer",
		"Deposit",
		"Withdrawal",
		"Refund",
		"Purchase",
		"Subscription",
		"Salary",
		"Dividend",
		"Interest",
	}

	commentType := commentTypes[rand.Intn(len(commentTypes))]

	switch commentType {
	case "Payment":
		return fmt.Sprintf("Payment to %s", toAccount.Name)
	case "Transfer":
		return fmt.Sprintf("Transfer to %s", toAccount.Name)
	case "Deposit":
		return fmt.Sprintf("Deposit to %s", toAccount.Name)
	case "Withdrawal":
		return fmt.Sprintf("Withdrawal from %s", fromAccount.Name)
	case "Refund":
		return fmt.Sprintf("Refund from %s", fromAccount.Name)
	case "Purchase":
		company := companyList[rand.Intn(len(companyList))]
		return fmt.Sprintf("Purchase at %s", company)
	case "Subscription":
		company := companyList[rand.Intn(len(companyList))]
		return fmt.Sprintf("%s subscription", company)
	case "Salary":
		company := companyList[rand.Intn(len(companyList))]
		return fmt.Sprintf("Salary from %s", company)
	case "Dividend":
		company := companyList[rand.Intn(len(companyList))]
		return fmt.Sprintf("Dividend from %s", company)
	case "Interest":
		return fmt.Sprintf("Interest payment %.2f%%", rand.Float64()*5)
	default:
		return fmt.Sprintf("Transfer of %.2f %s", amount, fromAccount.Currency)
	}
}

func main() {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "billing-perf-test-*")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create subdirectories
	dbPath := filepath.Join(tempDir, "db")
	metaPath := filepath.Join(tempDir, "meta")

	if err := os.MkdirAll(dbPath, 0755); err != nil {
		log.Fatalf("Failed to create db directory: %v", err)
	}

	if err := os.MkdirAll(metaPath, 0755); err != nil {
		log.Fatalf("Failed to create meta directory: %v", err)
	}

	// Set environment variables to use the temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Start CPU profiling
	cpuProfile, err := os.Create("cpu_profile.prof")
	if err != nil {
		log.Fatalf("Failed to create CPU profile: %v", err)
	}
	pprof.StartCPUProfile(cpuProfile)
	defer pprof.StopCPUProfile()

	// Create a new database store
	fmt.Println("Creating database store...")
	startTime := time.Now()
	dbStore, err := models.NewDBStore()
	if err != nil {
		log.Fatalf("Failed to create database store: %v", err)
	}
	defer dbStore.Close()
	fmt.Printf("Database store created in %v\n", time.Since(startTime))

	// Create user, account, and transaction stores
	userStore := &models.UserStore{Store: dbStore}
	accountStore := models.NewAccountStore(dbStore)
	transactionStore := models.NewTransactionStore(dbStore)

	// Create users
	fmt.Printf("Creating %d users...\n", numUsers)
	startTime = time.Now()
	users := make([]*models.User, numUsers)
	for i := 0; i < numUsers; i++ {
		user := &models.User{
			Key:      generateRandomName(),
			Accounts: []uint32{},
		}

		if err := userStore.Save(user); err != nil {
			log.Fatalf("Failed to save user %d: %v", i, err)
		}

		users[i] = user

		if (i+1)%100 == 0 {
			fmt.Printf("Created %d users...\n", i+1)
		}
	}
	fmt.Printf("Created %d users in %v\n", numUsers, time.Since(startTime))

	// Create accounts
	fmt.Printf("Creating %d accounts...\n", numAccounts)
	startTime = time.Now()
	accounts := make([]*models.Account, numAccounts)
	for i := 0; i < numAccounts; i++ {
		// Assign to a random user
		userIndex := rand.Intn(numUsers)
		user := users[userIndex]

		account := &models.Account{
			UserID:   user.ID,
			Name:     generateRandomAccountName(),
			Currency: generateRandomCurrency(),
			Amount:   generateRandomAmount(1000, 10000),
		}

		if err := accountStore.Save(account); err != nil {
			log.Fatalf("Failed to save account %d: %v", i, err)
		}

		accounts[i] = account

		if (i+1)%300 == 0 {
			fmt.Printf("Created %d accounts...\n", i+1)
		}
	}
	fmt.Printf("Created %d accounts in %v\n", numAccounts, time.Since(startTime))

	// Create transactions
	fmt.Printf("Creating %d transactions...\n", numTransactions)
	startTime = time.Now()
	var memStats runtime.MemStats

	for i := 0; i < numTransactions; i++ {
		// Select random source and destination accounts
		fromIndex := rand.Intn(numAccounts)
		var toIndex int
		for {
			toIndex = rand.Intn(numAccounts)
			if toIndex != fromIndex {
				break
			}
		}

		fromAccount := accounts[fromIndex]
		toAccount := accounts[toIndex]

		// Generate a realistic amount (usually smaller than the available balance)
		maxAmount := fromAccount.Amount * 0.2
		if maxAmount < 10 {
			maxAmount = 10
		}
		amount := generateRandomAmount(1, maxAmount)

		transaction := &models.Transaction{
			From:    fromAccount.ID,
			To:      toAccount.ID,
			Comment: generateRandomComment(fromAccount, toAccount, amount),
			Amount:  amount,
		}

		if err := transactionStore.Save(transaction); err != nil {
			log.Fatalf("Failed to save transaction %d: %v", i, err)
		}

		// Update our local copy of the accounts
		fromAccount.Amount -= amount
		toAccount.Amount += amount

		if (i+1)%1000 == 0 {
			elapsed := time.Since(startTime)
			transPerSec := float64(i+1) / elapsed.Seconds()

			// Collect memory stats
			runtime.ReadMemStats(&memStats)

			fmt.Printf("Created %d transactions (%.2f/sec), Memory: %.2f MB\n",
				i+1,
				transPerSec,
				float64(memStats.Alloc)/1024/1024)
		}
	}

	totalTime := time.Since(startTime)
	transPerSec := float64(numTransactions) / totalTime.Seconds()
	fmt.Printf("Created %d transactions in %v (%.2f transactions/sec)\n",
		numTransactions,
		totalTime,
		transPerSec)

	// Memory profile
	memProfile, err := os.Create("mem_profile.prof")
	if err != nil {
		log.Fatalf("Failed to create memory profile: %v", err)
	}
	runtime.GC() // Run garbage collection before taking memory profile
	if err := pprof.WriteHeapProfile(memProfile); err != nil {
		log.Fatalf("Failed to write memory profile: %v", err)
	}
	memProfile.Close()

	// Final memory stats
	runtime.ReadMemStats(&memStats)
	fmt.Printf("\nFinal Memory Stats:\n")
	fmt.Printf("Alloc: %.2f MB\n", float64(memStats.Alloc)/1024/1024)
	fmt.Printf("TotalAlloc: %.2f MB\n", float64(memStats.TotalAlloc)/1024/1024)
	fmt.Printf("Sys: %.2f MB\n", float64(memStats.Sys)/1024/1024)
	fmt.Printf("NumGC: %d\n", memStats.NumGC)

	// Test retrieving random transactions
	fmt.Println("\nTesting random transaction retrieval...")
	startTime = time.Now()
	for i := 0; i < 1000; i++ {
		txID := rand.Uint32()%uint32(numTransactions) + 1
		_, err := transactionStore.GetByID(txID)
		if err != nil {
			// Skip errors for non-existent transactions
			continue
		}
	}
	fmt.Printf("Retrieved 1000 random transactions in %v\n", time.Since(startTime))

	// Test retrieving random accounts
	fmt.Println("\nTesting random account retrieval...")
	startTime = time.Now()
	for i := 0; i < 1000; i++ {
		accountID := rand.Uint32()%uint32(numAccounts) + 1
		_, err := accountStore.GetByID(accountID)
		if err != nil {
			// Skip errors for non-existent accounts
			continue
		}
	}
	fmt.Printf("Retrieved 1000 random accounts in %v\n", time.Since(startTime))

	// Test retrieving random users
	fmt.Println("\nTesting random user retrieval...")
	startTime = time.Now()
	for i := 0; i < 1000; i++ {
		userID := rand.Uint32()%uint32(numUsers) + 1
		_, err := userStore.GetByID(userID)
		if err != nil {
			// Skip errors for non-existent users
			continue
		}
	}
	fmt.Printf("Retrieved 1000 random users in %v\n", time.Since(startTime))

	fmt.Println("\nPerformance test completed successfully!")
	fmt.Printf("CPU profile saved to cpu_profile.prof\n")
	fmt.Printf("Memory profile saved to mem_profile.prof\n")
	fmt.Printf("To analyze profiles, run: go tool pprof cpu_profile.prof\n")
}
