package jsonnet

import (
	"encoding/json"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/brianvoe/gofakeit/v7"
	jsonnet "github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"go.k6.io/k6/js/modules"
	"go.uber.org/zap"
)

// Register the extension on module initialization
func init() {
	modules.Register("k6/x/jsonnet", new(JsonnetExtension))
}

// JsonnetExtension is the k6 extension for Jsonnet processing
type JsonnetExtension struct {
	logger *zap.SugaredLogger
}

// XJsonnet represents the JavaScript API
type XJsonnet struct {
	extension *JsonnetExtension
}

// NewModuleInstance implements the k6 modules.Module interface
func (j *JsonnetExtension) NewModuleInstance(vu modules.VU) modules.Instance {
	// Initialize logger
	rawLogger, _ := zap.NewProduction()
	logger := rawLogger.Sugar()

	return &XJsonnet{
		extension: &JsonnetExtension{
			logger: logger,
		},
	}
}

// Exports returns the exports of the JS module
func (j *XJsonnet) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{
			"processTemplate":  j.processTemplate,
			"generateTestData": j.generateTestData,
			"loadTemplate":     j.loadTemplate,
		},
	}
}

// ProcessTemplate processes a Jsonnet template and returns JSON
func (j *XJsonnet) processTemplate(templatePath string, generateTestData bool) (string, error) {
	j.extension.logger.Infof("Processing template: %s, generateTestData: %t", templatePath, generateTestData)

	jsonData := j.generateJsonFromTemplate(templatePath, generateTestData)
	return jsonData, nil
}

// GenerateTestData is a convenience method for generating test data
func (j *XJsonnet) generateTestData(templatePath string) (string, error) {
	return j.processTemplate(templatePath, true)
}

// LoadTemplate loads a template without processing (for inspection)
func (j *XJsonnet) loadTemplate(templatePath string) (string, error) {
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ProcessMultipleTemplates processes multiple templates using a pattern
func (j *XJsonnet) processMultipleTemplates(templateRoot string, pattern string, generateTestData bool) (map[string]string, error) {
	j.extension.logger.Infof("Processing multiple templates: root=%s, pattern=%s", templateRoot, pattern)

	fileSearchPath := path.Join(templateRoot, pattern)
	jsonnetFiles, globErr := doublestar.Glob(fileSearchPath)
	if globErr != nil {
		return nil, globErr
	}

	results := make(map[string]string)
	for _, jsonnetFile := range jsonnetFiles {
		jsonData := j.generateJsonFromTemplate(jsonnetFile, generateTestData)
		results[jsonnetFile] = jsonData
	}

	return results, nil
}

// createFakeFunction creates a jsonnet native function for generating fake data
func (j *XJsonnet) createFakeFunction() *jsonnet.NativeFunction {
	return &jsonnet.NativeFunction{
		Name:   "fake",
		Params: ast.Identifiers{"x"},
		Func: func(x []interface{}) (interface{}, error) {
			bytes, err := json.Marshal(x[0])
			if err != nil {
				return nil, err
			}
			return j.callGoFakeIt(strings.Trim(string(bytes), "\""))
		},
	}
}

// callGoFakeIt generates fake data using gofakeit library
func (j *XJsonnet) callGoFakeIt(pattern string) (string, error) {
	return gofakeit.Generate(pattern)
}

// generateJsonFromTemplate processes a jsonnet template file and returns the generated JSON string
func (j *XJsonnet) generateJsonFromTemplate(templateFilePath string, generateTestData bool) string {
	vm := jsonnet.MakeVM()

	// Set up the jsonnet VM
	vm.Importer(&jsonnet.FileImporter{})
	vm.ExtVar("generateTestData", strconv.FormatBool(generateTestData))

	// Only register the fake function if generateTestData is true
	if generateTestData {
		vm.NativeFunction(j.createFakeFunction())
	}

	// Evaluate the jsonnet file
	jsonStr, err := vm.EvaluateFile(templateFilePath)
	if err != nil {
		j.extension.logger.Errorf("Error evaluating template %s: %v", templateFilePath, err)
		return ""
	}
	return jsonStr
}
