package bootstrapping

import (
	"bufio"
	"fmt"
	"lsf-configurator/pkg/filesystem"
	"os"
	"path"
	"regexp"
	"strings"
)

const (
	ReqFileName          = "requirements.txt"
	MainFile             = "func"
	ComponentHandlerFunc = "handler"
	Extension            = ".py"
)

type PythonBootstrapper struct {
	BaseBootstrapper
}

func (b *PythonBootstrapper) Setup() error {
	componentNames := []string{}
	filesToCopy := []string{}

	for _, component := range b.fc.Components {
		componentNames = append(componentNames, component.Name)
		filesToCopy = append(filesToCopy, component.Name+Extension)
	}

	componentFiles, err := filesystem.CopyFilesByNames(b.fc.SourcePath, b.buildDir, filesToCopy, false)
	if err != nil {
		return err
	}

	for _, file := range componentFiles {
		hasHandler, err := hasHandlerFunction(file)
		if err != nil {
			return err
		}

		if !hasHandler {
			return fmt.Errorf("no handler was found for file: %v", file)
		}
	}

	templateReqFile := path.Join(b.buildDir, ReqFileName)
	userReqFile := path.Join(b.fc.SourcePath, ReqFileName)
	err = mergeRequirements(templateReqFile, userReqFile)
	if err != nil {
		return err
	}

	return modifyMain(b.buildDir, b.fc.Id, componentNames)
}

func mergeRequirements(existingFile, newFile string) error {
	reqs, err := readRequirements(existingFile)
	if err != nil {
		return err
	}

	newReqs, err := readRequirements(newFile)
	if err != nil {
		return err
	}

	for dep := range newReqs {
		reqs[dep] = true
	}

	if filesystem.FileExists(existingFile) {
		err := os.Remove(existingFile)
		if err != nil {
			return err
		}
	}
	return writeRequirements(existingFile, reqs)
}

func readRequirements(filePath string) (map[string]bool, error) {
	deps := make(map[string]bool)

	if !filesystem.FileExists(filePath) {
		return deps, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			deps[line] = true
		}
	}
	return deps, scanner.Err()
}

func writeRequirements(filePath string, deps map[string]bool) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for dep := range deps {
		if _, err := file.WriteString(dep + "\n"); err != nil {
			return err
		}
	}

	return nil
}

func hasHandlerFunction(filePath string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	handlerRegex := regexp.MustCompile(`(?m)^def handler\(`)
	return handlerRegex.Match(content), nil
}

func modifyMain(buildDir, funcName string, componentNames []string) error {
	const functionNameStr = "FUNCTION_NAME ="
	const handlersStr = "HANDLERS = {"

	mainFilePath := path.Join(buildDir, MainFile+Extension)
	content, err := os.ReadFile(mainFilePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	modifiedContent := []string{}
	importSection := true
	handlersSection := false

	for _, line := range lines {
		if importSection && !strings.HasPrefix(line, "from") && !strings.HasPrefix(line, "import") && line != "" {
			importSection = false
			for _, component := range componentNames {
				importStatement := fmt.Sprintf("from %s import %s as %s", component, ComponentHandlerFunc, component)
				if !strings.Contains(strings.Join(modifiedContent, "\n"), importStatement) {
					modifiedContent = append(modifiedContent, importStatement)
				}
			}
		}

		if strings.Contains(line, functionNameStr) {
			modifiedContent = append(modifiedContent, fmt.Sprintf("%s\"%s\"", functionNameStr, funcName))
			continue
		}

		if strings.Contains(line, handlersStr) {
			handlersSection = true
			modifiedContent = append(modifiedContent, handlersStr)
			for _, component := range componentNames {
				modifiedContent = append(modifiedContent, fmt.Sprintf("        \"%s\" : %s,", component, component))
			}
			modifiedContent = append(modifiedContent, "    }")
			continue
		}

		if handlersSection && strings.Contains(line, "}") {
			handlersSection = false
			continue
		}
		modifiedContent = append(modifiedContent, line)
	}

	return os.WriteFile(mainFilePath, []byte(strings.Join(modifiedContent, "\n")), 0644)
}
