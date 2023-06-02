package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"kusionstack.io/kpm/pkg/env"
	"kusionstack.io/kpm/pkg/utils"
)

const testDataDir = "test_data"

func getTestDir(subDir string) string {
	pwd, _ := os.Getwd()
	testDir := filepath.Join(pwd, testDataDir)
	testDir = filepath.Join(testDir, subDir)

	return testDir
}

func TestSettingInit(t *testing.T) {
	kpmHome, err := env.GetAbsPkgPath()
	assert.Equal(t, err, nil)
	settings, err := GetSettings()
	assert.Equal(t, err, nil)
	assert.Equal(t, settings.CredentialsFile, filepath.Join(kpmHome, CONFIG_JSON_PATH))
}

func TestGetFullJsonPath(t *testing.T) {
	path, err := GetFullPath("test.json")
	assert.Equal(t, err, nil)

	kpmHome, err := env.GetAbsPkgPath()
	assert.Equal(t, err, nil)
	assert.Equal(t, path, filepath.Join(kpmHome, "test.json"))
}

func TestDefaultKpmConf(t *testing.T) {
	settings := Settings{
		Conf: DefaultKpmConf(),
	}
	assert.Equal(t, settings.DefaultOciRegistry(), "ghcr.io")
	assert.Equal(t, settings.DefaultOciRepo(), "kusionstack")
}

func TestLoadOrCreateDefaultKpmJson(t *testing.T) {
	testDir := getTestDir("expected.json")
	kpmPath := filepath.Join(filepath.Join(filepath.Join(filepath.Dir(testDir), ".kpm"), "config"), "kpm.json")
	err := os.Setenv("KCL_PKG_PATH", filepath.Dir(testDir))

	assert.Equal(t, err, nil)
	assert.Equal(t, utils.DirExists(kpmPath), false)

	kpmConf, err := loadOrCreateDefaultKpmJson()
	assert.Equal(t, kpmConf.DefaultOciRegistry, "ghcr.io")
	assert.Equal(t, kpmConf.DefaultOciRepo, "kusionstack")
	assert.Equal(t, err, nil)
	assert.Equal(t, utils.DirExists(kpmPath), true)

	expectedJson, err := os.ReadFile(testDir)
	assert.Equal(t, err, nil)

	gotJson, err := os.ReadFile(kpmPath)
	assert.Equal(t, err, nil)

	var expected interface{}
	err = json.Unmarshal(expectedJson, &expected)
	assert.Equal(t, err, nil)

	var got interface{}
	err = json.Unmarshal(gotJson, &got)
	assert.Equal(t, err, nil)

	assert.Equal(t, reflect.DeepEqual(expected, got), true)

	os.RemoveAll(kpmPath)
	assert.Equal(t, utils.DirExists(kpmPath), false)
}

func TestPackageCacheLock(t *testing.T) {

	settings, err := GetSettings()
	assert.Equal(t, err, nil)

	// create the expected result of the test.
	// 10 times of "goroutine 1: %d" at first, and then 10 times of "goroutine 2: %d"

	// If goroutine 1 get the lock first, then it will append "goroutine 1: %d" to the list.
	goroutine_1_first_list := []string{}

	for i := 0; i < 10; i++ {
		goroutine_1_first_list = append(goroutine_1_first_list, fmt.Sprintf("goroutine 1: %d", i))
	}

	for i := 0; i < 10; i++ {
		goroutine_1_first_list = append(goroutine_1_first_list, fmt.Sprintf("goroutine 2: %d", i))
	}

	// If goroutine 2 get the lock first, then it will append "goroutine 2: %d" to the list.
	goroutine_2_first_list := []string{}

	for i := 0; i < 10; i++ {
		goroutine_2_first_list = append(goroutine_2_first_list, fmt.Sprintf("goroutine 2: %d", i))
	}

	for i := 0; i < 10; i++ {
		goroutine_2_first_list = append(goroutine_2_first_list, fmt.Sprintf("goroutine 1: %d", i))
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// create a list to store the result generated by 2 goroutine concurrently.
	gotlist := []string{}

	// goroutine 1: append "goroutine 1: %d" to the list
	go func() {
		defer wg.Done()
		err := settings.AcquirePackageCacheLock()
		assert.Equal(t, err, nil)
		for i := 0; i < 10; i++ {
			gotlist = append(gotlist, fmt.Sprintf("goroutine 1: %d", i))
		}
		err = settings.ReleasePackageCacheLock()
		assert.Equal(t, err, nil)
	}()

	// goroutine 2: append "goroutine 2: %d" to the list
	go func() {
		defer wg.Done()
		err := settings.AcquirePackageCacheLock()
		assert.Equal(t, err, nil)
		for i := 0; i < 10; i++ {
			gotlist = append(gotlist, fmt.Sprintf("goroutine 2: %d", i))
		}
		err = settings.ReleasePackageCacheLock()
		assert.Equal(t, err, nil)
	}()

	wg.Wait()

	// Compare the gotlist and expectedlist.
	assert.Equal(t,
		(reflect.DeepEqual(gotlist, goroutine_1_first_list) || reflect.DeepEqual(gotlist, goroutine_2_first_list)),
		true)
}