package license

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// UpdateProjectLicense writes a new project license file
func UpdateProjectLicense(licenseFilepath string) error {

	pLicense, err := ioutil.ReadFile(licenseFilepath)
	if err != nil {
		return err
	}

	licenseF, err := os.Create("LICENSE")
	if err != nil {
		return err
	}
	defer licenseF.Close()
	licenseF.WriteString(string(pLicense))

	return nil
}

// UpdateSourceHeader updates the license boilerplate used for the licensing
// header in source code files
func UpdateSourceHeader(headerFilepath string) error {

	sLicense, err := ioutil.ReadFile(headerFilepath)
	if err != nil {
		return err
	}

	if _, err = os.Stat("hack"); os.IsNotExist(err) {
		err = os.Mkdir("hack", 0755)
		if err != nil {
			return err
		}
	}

	licenseB, err := os.Create("hack/boilerplate.go.txt")
	if err != nil {
		return err
	}
	defer licenseB.Close()
	licenseB.WriteString(string(sLicense) + "\n")

	return nil
}

// UpdateExistingSourceHeader rewrites the licensing header for all pre-existing
// source code files
func UpdateExistingSourceHeader(headerFilepath string) error {

	sLicense, err := ioutil.ReadFile(headerFilepath)
	if err != nil {
		return err
	}

	filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		filename := info.Name()
		if len(filename) > 3 && filename[len(filename)-3:] == ".go" {
			replaceLicenseHeader(path, sLicense)
		}
		return nil
	})

	return nil
}

func replaceLicenseHeader(filepath string, header []byte) error {

	input, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(input), "\n")

	sourceBeginningFound := false
	var output string
	for _, line := range lines {
		if !sourceBeginningFound {
			if len(line) > 7 && line[0:7] == "package" {
				sourceBeginningFound = true
				output = string(header) + "\n" + line + "\n"
			}
		} else {
			output = output + line + "\n"
		}
	}

	err = ioutil.WriteFile(filepath, []byte(output), 600)
	if err != nil {
		return err
	}

	return nil
}
