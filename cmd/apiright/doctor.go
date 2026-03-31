package apiright

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/bata94/apiright/pkg/config"
	"github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/database"
	"github.com/spf13/cobra"
)

func NewDoctorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "doctor",
		Short:   "Check system requirements",
		Long:    `Verifies that all required tools and dependencies are properly installed.`,
		Example: `  apiright doctor`,
		RunE:    runDoctor,
	}

	return cmd
}

func runDoctor(cmd *cobra.Command, args []string) error {
	fmt.Println(core.Boldf("APIRight Doctor - System Check"))
	fmt.Println(strings.Repeat("=", 40))

	allPassed := true

	fmt.Print("\n" + core.Cyanf("Checking Go installation..."))
	if runtime.Version() != "" {
		fmt.Printf(" \r%s Go %s\n", core.Greenf("✓"), runtime.Version())
	} else {
		fmt.Printf(" \r%s Go not found\n", core.Redf("✗"))
		allPassed = false
	}

	fmt.Print(core.Cyanf("Checking sqlc..."))
	if err := checkSqlc(); err != nil {
		fmt.Printf(" \r%s sqlc: %s\n", core.Redf("✗"), err)
		allPassed = false
	} else {
		fmt.Printf(" \r%s sqlc installed\n", core.Greenf("✓"))
	}

	fmt.Print(core.Cyanf("Checking project..."))
	projectDir, _ := GetProjectDir(cmd)
	if err := checkProject(projectDir); err != nil {
		fmt.Printf(" \r%s %s\n", core.Yellowf("!"), err)
	} else {
		fmt.Printf(" \r%s Project looks good\n", core.Greenf("✓"))
	}

	fmt.Print(core.Cyanf("Checking database..."))
	if err := checkDatabase(projectDir); err != nil {
		fmt.Printf(" \r%s %s\n", core.Yellowf("!"), err)
	} else {
		fmt.Printf(" \r%s Database connected\n", core.Greenf("✓"))
	}

	fmt.Println("\n" + strings.Repeat("=", 40))
	if allPassed {
		fmt.Println(core.Greenf("All checks passed!"))
	} else {
		fmt.Println(core.Yellowf("Some checks failed. See above for details."))
	}

	return nil
}

func checkSqlc() error {
	cmd := exec.Command("sqlc", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not installed or not in PATH")
	}
	return nil
}

func checkProject(projectDir string) error {
	required := []string{"migrations", "sqlc.yaml", "apiright.yaml"}
	missing := []string{}

	for _, dir := range required {
		path := projectDir + "/" + dir
		if _, err := os.Stat(path); os.IsNotExist(err) {
			missing = append(missing, dir)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing: %s", strings.Join(missing, ", "))
	}
	return nil
}

func checkDatabase(projectDir string) error {
	cfg, err := config.LoadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	logger, _ := core.NewLogger(false)
	db, err := database.NewDatabase(&cfg.Database, logger)
	if err != nil {
		return fmt.Errorf("database init error: %w", err)
	}

	if err := db.Connect(); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer core.Close("database", db, logger)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}
