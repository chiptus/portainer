package kubernetes

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"

	"github.com/portainer/portainer-ee/api/internal/slices"
)

const (
	labelPortainerAppStack   = "io.portainer.kubernetes.application.stack"
	labelPortainerAppStackID = "io.portainer.kubernetes.application.stackid"
	labelPortainerAppName    = "io.portainer.kubernetes.application.name"
	labelPortainerAppOwner   = "io.portainer.kubernetes.application.owner"
	labelPortainerAppKind    = "io.portainer.kubernetes.application.kind"
)

// KubeAppLabels are labels applied to all resources deployed in a kubernetes stack
type KubeAppLabels struct {
	StackID   int
	StackName string
	Owner     string
	Kind      string
}

type KubeResource struct {
	Kind      string
	Name      string
	Namespace string
}

// convert string to valid kubernetes label by replacing invalid characters with periods and removing any periods at the beginning or end of the string
func sanitizeLabel(value string) string {
	re := regexp.MustCompile(`[^A-Za-z0-9\.\-\_]+`)
	onlyAllowedCharacterString := re.ReplaceAllString(value, ".")
	return strings.Trim(onlyAllowedCharacterString, ".-_")
}

// ToMap converts KubeAppLabels to a map[string]string
func (kal *KubeAppLabels) ToMap() map[string]string {
	return map[string]string{
		labelPortainerAppStackID: strconv.Itoa(kal.StackID),
		labelPortainerAppStack:   sanitizeLabel(kal.StackName),
		labelPortainerAppName:    sanitizeLabel(kal.StackName),
		labelPortainerAppOwner:   sanitizeLabel(kal.Owner),
		labelPortainerAppKind:    kal.Kind,
	}
}

// GetHelmAppLabels returns the labels to be applied to portainer deployed helm applications
func GetHelmAppLabels(name, owner string) map[string]string {
	return map[string]string{
		labelPortainerAppName:  name,
		labelPortainerAppOwner: sanitizeLabel(owner),
	}
}

// AddAppLabels adds required labels to "Resource"->metadata->labels.
// It'll add those labels to all Resource (nodes with a kind property excluding a list) it can find in provided yaml.
// Items in the yaml file could either be organised as a list or broken into multi documents.
func AddAppLabels(manifestYaml []byte, appLabels map[string]string) ([]byte, error) {
	if bytes.Equal(manifestYaml, []byte("")) {
		return manifestYaml, nil
	}

	postProcessYaml := func(yamlDoc interface{}) error {
		addResourceLabels(yamlDoc, appLabels)
		return nil
	}

	docs, err := ExtractDocuments(manifestYaml, postProcessYaml)
	if err != nil {
		return nil, err
	}

	return bytes.Join(docs, []byte("---\n")), nil
}

// ExtractDocuments extracts all the documents from a yaml file
// Optionally post-process each document with a function, which can modify the document in place.
// Pass in nil for postProcessYaml to skip post-processing.
func ExtractDocuments(manifestYaml []byte, postProcessYaml func(interface{}) error) ([][]byte, error) {
	docs := make([][]byte, 0)
	yamlDecoder := yaml.NewDecoder(bytes.NewReader(manifestYaml))

	for {
		m := make(map[string]interface{})
		err := yamlDecoder.Decode(&m)
		// if there are no more documents in the file
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, errors.Wrap(err, "failed to decode yaml manifest")
		}

		// if decoded document is empty
		if m == nil {
			continue
		}

		// optionally post-process yaml
		if postProcessYaml != nil {
			if err := postProcessYaml(m); err != nil {
				return nil, errors.Wrap(err, "failed to post process yaml document")
			}
		}

		var out bytes.Buffer
		yamlEncoder := yaml.NewEncoder(&out)
		yamlEncoder.SetIndent(2)
		if err := yamlEncoder.Encode(m); err != nil {
			return nil, errors.Wrap(err, "failed to marshal yaml manifest")
		}

		docs = append(docs, out.Bytes())
	}

	return docs, nil
}

// GetNamespace returns the namespace of a kubernetes resource from its metadata
// It returns an empty string if namespace is not found in the resource
func GetNamespace(manifestYaml []byte) (string, error) {
	yamlDecoder := yaml.NewDecoder(bytes.NewReader(manifestYaml))
	m := make(map[string]interface{})
	err := yamlDecoder.Decode(&m)
	if err != nil {
		return "", errors.Wrap(err, "failed to unmarshal yaml manifest when obtaining namespace")
	}

	kind := m["kind"].(string)

	if _, ok := m["metadata"]; ok {
		if kind == "Namespace" {
			if namespace, ok := m["metadata"].(map[string]interface{})["name"]; ok {
				return namespace.(string), nil
			}
		} else {
			if namespace, ok := m["metadata"].(map[string]interface{})["namespace"]; ok {
				return namespace.(string), nil
			}
		}
	}

	return "", nil
}

// GetImagesFromManifest returns a list of images referenced in a kubernetes manifest
func GetImagesFromManifest(manifestYaml []byte) ([]string, error) {
	var n yaml.Node
	err := yaml.Unmarshal([]byte(manifestYaml), &n)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal yaml manifest when obtaining images")
	}

	p, err := yamlpath.NewPath("$..spec.containers[*].image")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create yamlpath when obtaining images")
	}

	q, err := p.Find(&n)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find images in yaml manifest")
	}

	images := make([]string, 0)
	for _, v := range q {
		images = append(images, v.Value)
	}

	return images, nil
}

// GetResourcesFromManifest returns a list of kubernetes resource kinds from a manifest optionally filtered by resource kind
func GetResourcesFromManifest(manifestYaml []byte, filters []string) ([]KubeResource, error) {
	kinds := make([]KubeResource, 0)

	// ensure the filters are be lower case
	for i := 0; i < len(filters); i++ {
		filters[i] = strings.ToLower(filters[i])
	}

	postProcessYaml := func(yamlDoc interface{}) error {
		kind := KubeResource{}

		if val, ok := yamlDoc.(map[string]interface{})["kind"].(string); ok {
			if len(filters) == 0 || slices.Contains(filters, strings.ToLower(val)) {
				kind.Kind = strings.ToLower(val)
			} else {
				return nil
			}
		}

		if m, ok := yamlDoc.(map[string]interface{}); ok {
			if metadata, ok := m["metadata"].(map[string]interface{}); ok {
				if name, ok := metadata["name"].(string); ok {
					kind.Name = name
				}
				if name, ok := metadata["namespace"].(string); ok {
					kind.Namespace = name
				}
			}
		}

		kinds = append(kinds, kind)
		return nil
	}

	_, err := ExtractDocuments(manifestYaml, postProcessYaml)
	if err != nil {
		return nil, err
	}

	return kinds, nil
}

func addResourceLabels(yamlDoc interface{}, appLabels map[string]string) {
	m, ok := yamlDoc.(map[string]interface{})
	if !ok {
		return
	}

	kind, ok := m["kind"]
	if ok && !strings.EqualFold(kind.(string), "list") {
		addLabels(m, appLabels)
		return
	}

	for _, v := range m {
		switch v.(type) {
		case map[string]interface{}:
			addResourceLabels(v, appLabels)
		case []interface{}:
			for _, item := range v.([]interface{}) {
				addResourceLabels(item, appLabels)
			}
		}
	}
}

func addLabels(obj map[string]interface{}, appLabels map[string]string) {
	metadata := make(map[string]interface{})
	if m, ok := obj["metadata"]; ok {
		metadata = m.(map[string]interface{})
	}

	labels := make(map[string]string)
	if l, ok := metadata["labels"]; ok {
		for k, v := range l.(map[string]interface{}) {
			labels[k] = fmt.Sprintf("%v", v)
		}
	}

	// merge app labels with existing labels
	for k, v := range appLabels {
		labels[k] = v
	}

	metadata["labels"] = labels
	obj["metadata"] = metadata
}

// UpdateContainerEnv updates variables in "Resource"->spec->containers.
// It will not add new environment variables
func UpdateContainerEnv(manifestYaml []byte, env map[string]string) ([]byte, error) {
	if bytes.Equal(manifestYaml, []byte("")) {
		return manifestYaml, nil
	}

	postProcessYaml := func(yamlDoc interface{}) error {
		updateContainerEnv(yamlDoc, env)
		return nil
	}

	docs, err := ExtractDocuments(manifestYaml, postProcessYaml)
	if err != nil {
		return nil, err
	}

	return bytes.Join(docs, []byte("---\n")), nil
}

func updateContainerEnv(yamlDoc interface{}, env map[string]string) {
	m, ok := yamlDoc.(map[string]interface{})
	if !ok {
		return
	}

	if containers, ok := m["containers"]; ok {
		c, ok := containers.([]interface{})
		if ok {
			updateEnv(c, env)
		}
		return
	}

	for _, v := range m {
		switch v.(type) {
		case map[string]interface{}:
			updateContainerEnv(v, env)
		}
	}
}

func updateEnv(containers []interface{}, env map[string]string) {
	for _, c := range containers {
		c, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		if existingEnv, ok := c["env"]; ok {
			v, ok := existingEnv.([]interface{})
			if !ok {
				continue
			}

			// What we have here is a slice of maps.  Each map contains a key value pair with the key
			// names: "name" and "value".  We need to merge this with our env which is just a
			// map of key value pairs.
			for _, v := range v {
				// Replace existing value with new value if it exists
				v, ok := v.(map[string]interface{})
				if !ok {
					continue
				}

				k, ok := v["name"].(string)
				if !ok {
					continue
				}

				if _, exists := env[k]; exists {
					v["value"] = env[k]
				}
			}

			c["env"] = v
		}
	}
}
