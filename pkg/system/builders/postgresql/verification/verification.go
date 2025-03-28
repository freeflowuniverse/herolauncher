package verification

import (
	"fmt"
	"os/exec"
)

// Verifier handles the verification of PostgreSQL installation
type Verifier struct {
	InstallPrefix string
}

// NewVerifier creates a new verifier
func NewVerifier(installPrefix string) *Verifier {
	return &Verifier{
		InstallPrefix: installPrefix,
	}
}

// VerifyPostgres verifies the PostgreSQL installation
func (v *Verifier) VerifyPostgres() (bool, error) {
	fmt.Println("Verifying PostgreSQL installation...")

	// Check for PostgreSQL binary
	postgresPath := fmt.Sprintf("%s/bin/postgres", v.InstallPrefix)
	fmt.Printf("Checking for PostgreSQL binary at %s\n", postgresPath)

	checkCmd := exec.Command("ls", "-la", postgresPath)
	output, err := checkCmd.CombinedOutput()

	if err != nil {
		fmt.Printf("❌ WARNING: PostgreSQL binary not found at expected location: %s\n", postgresPath)
		fmt.Println("This may indicate that the build process failed or installed to a different location.")

		// Search for PostgreSQL binary in other locations
		fmt.Println("Searching for PostgreSQL binary in other locations...")
		findCmd := exec.Command("find", "/", "-name", "postgres", "-type", "f")
		findOutput, _ := findCmd.CombinedOutput()
		fmt.Printf("Search results:\n%s\n", string(findOutput))

		return false, fmt.Errorf("PostgreSQL binary not found at expected location")
	}

	fmt.Printf("✅ PostgreSQL binary found at expected location:\n%s\n", string(output))
	return true, nil
}

// VerifyGoSP verifies the Go stored procedure installation
func (v *Verifier) VerifyGoSP() (bool, error) {
	fmt.Println("Verifying Go stored procedure installation...")

	// Check for Go stored procedure
	gospPath := fmt.Sprintf("%s/lib/libgosp.so", v.InstallPrefix)
	fmt.Printf("Checking for Go stored procedure at %s\n", gospPath)

	checkCmd := exec.Command("ls", "-la", gospPath)
	output, err := checkCmd.CombinedOutput()

	if err != nil {
		fmt.Printf("❌ WARNING: Go stored procedure library not found at expected location: %s\n", gospPath)

		// Search for Go stored procedure in other locations
		fmt.Println("Searching for Go stored procedure in other locations...")
		findCmd := exec.Command("find", "/", "-name", "libgosp.so", "-type", "f")
		findOutput, _ := findCmd.CombinedOutput()
		fmt.Printf("Search results:\n%s\n", string(findOutput))

		return false, fmt.Errorf("Go stored procedure library not found at expected location")
	}

	fmt.Printf("✅ Go stored procedure library found at expected location:\n%s\n", string(output))
	return true, nil
}

// Verify verifies the entire PostgreSQL installation
func (v *Verifier) Verify() (bool, error) {
	fmt.Println("=== Verifying PostgreSQL Installation ===")

	// Verify PostgreSQL
	postgresOk, postgresErr := v.VerifyPostgres()

	// Verify Go stored procedure
	gospOk, gospErr := v.VerifyGoSP()

	// Overall verification result
	success := postgresOk && gospOk

	if success {
		fmt.Println("✅ All components verified successfully!")
	} else {
		fmt.Println("⚠️ Some components could not be verified.")

		if postgresErr != nil {
			fmt.Printf("PostgreSQL verification error: %v\n", postgresErr)
		}

		if gospErr != nil {
			fmt.Printf("Go stored procedure verification error: %v\n", gospErr)
		}
	}

	return success, nil
}
