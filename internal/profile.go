package internal

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

type Profile struct {
	Branches []string `yaml:"branches"`
	tags     []string `yaml:"tags"`
}

func NewProfile(path string) (*Profile, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	// close file
	defer func(fp *os.File) {
		err := fp.Close()
		if err != nil {

		}
	}(fp)

	// read file
	profileBytes, err := ioutil.ReadAll(fp)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	// unmarshal
	profile := &Profile{}
	if err := yaml.Unmarshal(profileBytes, profile); err != nil {
		return nil, fmt.Errorf("cannot unmarshal file: %w", err)
	}

	return profile, nil
}
