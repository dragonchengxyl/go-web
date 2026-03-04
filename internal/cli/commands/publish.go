package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/studio/platform/internal/cli/config"
	"github.com/studio/platform/internal/cli/uploader"
)

// PublishCmd represents the publish command
var PublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a new game release",
	Long:  `Upload and publish a new game release to the Studio platform.`,
	RunE:  runPublish,
}

var (
	gameSlug     string
	branchName   string
	version      string
	releaseTitle string
	changelog    string
	packagePath  string
	platform     string
	autoPublish  bool
)

func init() {
	PublishCmd.Flags().StringVar(&gameSlug, "game", "", "Game slug identifier")
	PublishCmd.Flags().StringVar(&branchName, "branch", "main", "Branch name (main, beta, experimental)")
	PublishCmd.Flags().StringVar(&version, "version", "", "Version number (e.g., v1.2.3)")
	PublishCmd.Flags().StringVar(&releaseTitle, "title", "", "Release title")
	PublishCmd.Flags().StringVar(&changelog, "changelog", "", "Path to changelog file")
	PublishCmd.Flags().StringVar(&packagePath, "package", "", "Path to game package (.zip)")
	PublishCmd.Flags().StringVar(&platform, "platform", "windows", "Target platform (windows, macos, linux)")
	PublishCmd.Flags().BoolVar(&autoPublish, "auto-publish", false, "Automatically publish after upload")

	PublishCmd.MarkFlagRequired("game")
	PublishCmd.MarkFlagRequired("version")
	PublishCmd.MarkFlagRequired("package")
}

func runPublish(cmd *cobra.Command, args []string) error {
	// Load credentials
	store, err := config.NewCredentialsStore()
	if err != nil {
		return fmt.Errorf("failed to create credentials store: %w", err)
	}

	creds, err := store.Load()
	if err != nil {
		return err
	}

	fmt.Println("🚀 Starting game release publication...")
	fmt.Printf("🎮 Game: %s\n", gameSlug)
	fmt.Printf("🌿 Branch: %s\n", branchName)
	fmt.Printf("📦 Version: %s\n", version)
	fmt.Printf("💻 Platform: %s\n", platform)

	// Validate package file
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		return fmt.Errorf("package file not found: %s", packagePath)
	}

	// Calculate checksum
	fmt.Println("\n🔍 Calculating package checksum...")
	checksum, err := uploader.CalculateChecksum(packagePath)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}
	fmt.Printf("✅ SHA256: %s\n", checksum)

	// Read changelog if provided
	var changelogContent string
	if changelog != "" {
		data, err := os.ReadFile(changelog)
		if err != nil {
			return fmt.Errorf("failed to read changelog: %w", err)
		}
		changelogContent = string(data)
	}

	// Upload package
	fmt.Println("\n📤 Uploading package...")
	uploaderInstance, err := uploader.NewChunkedUploader(creds.APIUrl, creds.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to create uploader: %w", err)
	}
	downloadURL, err := uploaderInstance.Upload(packagePath, gameSlug, version)
	if err != nil {
		return fmt.Errorf("failed to upload package: %w", err)
	}
	fmt.Printf("✅ Upload complete: %s\n", downloadURL)

	// Get branch ID
	branchID, err := getBranchID(creds, gameSlug, branchName)
	if err != nil {
		return fmt.Errorf("failed to get branch ID: %w", err)
	}

	// Create release
	fmt.Println("\n📝 Creating release record...")
	releaseID, err := createRelease(creds, branchID, version, releaseTitle, changelogContent, downloadURL, checksum, platform)
	if err != nil {
		return fmt.Errorf("failed to create release: %w", err)
	}
	fmt.Printf("✅ Release created (ID: %s)\n", releaseID)

	// Auto-publish if requested
	if autoPublish {
		fmt.Println("\n🌐 Publishing release...")
		if err := publishRelease(creds, releaseID); err != nil {
			return fmt.Errorf("failed to publish release: %w", err)
		}
		fmt.Println("✅ Release published successfully!")
	} else {
		fmt.Println("\n⏸️  Release created but not published. Use --auto-publish to publish automatically.")
	}

	fmt.Printf("\n🎉 Done! Release URL: %s/games/%s/releases/%s\n", creds.APIUrl, gameSlug, releaseID)

	return nil
}

func getBranchID(creds *config.Credentials, gameSlug, branchName string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/games/%s/branches?name=%s", creds.APIUrl, gameSlug, branchName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+creds.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get branch with status %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Data) == 0 {
		return "", fmt.Errorf("branch not found: %s", branchName)
	}

	return result.Data[0].ID, nil
}

func createRelease(creds *config.Credentials, branchID, version, title, changelog, downloadURL, checksum, platform string) (string, error) {
	if title == "" {
		title = fmt.Sprintf("Release %s", version)
	}

	reqBody := map[string]interface{}{
		"version":      version,
		"title":        title,
		"changelog":    changelog,
		"download_url": downloadURL,
		"checksum":     checksum,
		"platform":     platform,
		"file_size":    getFileSize(downloadURL),
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/api/v1/branches/%s/releases", creds.APIUrl, branchID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+creds.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create release with status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Data.ID, nil
}

func publishRelease(creds *config.Credentials, releaseID string) error {
	url := fmt.Sprintf("%s/api/v1/releases/%s/publish", creds.APIUrl, releaseID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+creds.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to publish release with status %d", resp.StatusCode)
	}

	return nil
}

func getFileSize(packagePath string) int64 {
	info, err := os.Stat(packagePath)
	if err != nil {
		return 0
	}
	return info.Size()
}
