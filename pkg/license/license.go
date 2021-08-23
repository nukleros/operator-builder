package license

import (
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// UpdateProjectLicense writes a new project license file using a local file or
// HTTP URL as the source for license content.
func UpdateProjectLicense(source string) error {
	sourceLicense, err := getSourceLicense(source)
	if err != nil {
		return err
	}

	licenseF, err := os.Create("LICENSE")
	if err != nil {
		return err
	}
	defer licenseF.Close()

	if _, err := licenseF.WriteString(string(sourceLicense)); err != nil {
		return err
	}

	return nil
}

// UpdateSourceHeader writes the license boilerplate used by Kubebuilder for the
// licensing header in source code files.  It uses a local file or HTTP URL as
// the source for the header content.
func UpdateSourceHeader(source string) error {
	const directoryPermissions = 0755

	sourceLicense, err := getSourceLicense(source)
	if err != nil {
		return err
	}

	if _, err = os.Stat("hack"); os.IsNotExist(err) {
		err = os.Mkdir("hack", directoryPermissions)
		if err != nil {
			return err
		}
	}

	licenseB, err := os.Create("hack/boilerplate.go.txt")
	if err != nil {
		return err
	}
	defer licenseB.Close()

	if _, err := licenseB.WriteString(string(sourceLicense) + "\n"); err != nil {
		return err
	}

	return nil
}

// UpdateExistingSourceHeader rewrites the licensing header for all pre-existing
// source code files.  It uses a local file or HTTP URL as the source for the
// header content.
func UpdateExistingSourceHeader(source string) error {
	sourceLicense, err := getSourceLicense(source)
	if err != nil {
		return err
	}

	if err := filepath.WalkDir("./",
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			filename := d.Name()
			if len(filename) > 3 && filename[len(filename)-3:] == ".go" {
				if err := replaceLicenseHeader(path, sourceLicense); err != nil {
					return err
				}
			}

			return nil
		},
	); err != nil {
		return err
	}

	return nil
}

func getSourceLicense(source string) ([]byte, error) {
	var sourceLicense []byte

	if source[0:4] == "http" {
		// source is HTTP URL
		resp, err := http.Get(source)
		if err != nil {
			return []byte{}, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return []byte{}, err
		}

		sourceLicense = body
	} else {
		// source is local file
		fileContent, err := ioutil.ReadFile(source)
		if err != nil {
			return []byte{}, err
		}
		sourceLicense = fileContent
	}

	return sourceLicense, nil
}

func replaceLicenseHeader(file string, header []byte) error {
	const filePermissions = 0600

	input, err := ioutil.ReadFile(file)
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

	err = ioutil.WriteFile(file, []byte(output), filePermissions)
	if err != nil {
		return err
	}

	return nil
}
