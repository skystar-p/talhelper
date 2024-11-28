package talos

import (
	"os"

	"github.com/budimanjojo/talhelper/v3/pkg/config"
	"github.com/budimanjojo/talhelper/v3/pkg/decrypt"
	"github.com/siderolabs/talos/pkg/machinery/config/types/runtime/extensions"
	"gopkg.in/yaml.v3"
)

func GenerateExtensionServicesConfigBytes(esCfgs []*config.ExtensionService) ([]byte, error) {
	exts, err := GenerateNodeExtensionServiceConfig(esCfgs)
	if err != nil {
		return nil, err
	}

	return marshalAndCombineExtensionServiceYamls(exts)
}

func GenerateNodeExtensionServiceConfigBytesFromFiles(files []string) ([]byte, error) {
	var result []*extensions.ServiceConfigV1Alpha1

	var (
		content []byte
		err     error
	)

	for _, file := range files {
		// Try to decrypt with sops first.
		content, err = decrypt.DecryptYamlWithSops(file)
		if err != nil {
			// If failed, read the file as is.
			content, err = os.ReadFile(file)
			if err != nil {
				return nil, err
			}
		}

		var config extensions.ServiceConfigV1Alpha1
		if err := yaml.Unmarshal(content, &config); err != nil {
			return nil, err
		}

		if _, err := config.Validate(nil); err != nil {
			return nil, err
		}

		result = append(result, &config)
	}

	return marshalAndCombineExtensionServiceYamls(result)
}

func GenerateNodeExtensionServiceConfig(esCfgs []*config.ExtensionService) ([]*extensions.ServiceConfigV1Alpha1, error) {
	var result []*extensions.ServiceConfigV1Alpha1

	for _, v := range esCfgs {
		esc := extensions.NewServicesConfigV1Alpha1()
		esc.ServiceName = v.Name
		esc.ServiceConfigFiles = v.ConfigFiles
		esc.ServiceEnvironment = v.Environment

		if _, err := esc.Validate(nil); err != nil {
			return nil, err
		}

		result = append(result, esc)
	}

	return result, nil
}

func marshalAndCombineExtensionServiceYamls(exts []*extensions.ServiceConfigV1Alpha1) ([]byte, error) {
	var result [][]byte

	for _, ext := range exts {
		extByte, err := marshalYaml(ext)
		if err != nil {
			return nil, err
		}

		result = append(result, extByte)
	}

	return CombineYamlBytes(result), nil
}
