package build

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
)

type dockerfilebuild struct{}

func NewDockerfile() (Interface, error) {
	return &dockerfilebuild{}, nil
}

func (dfb *dockerfilebuild) IsSupportedReference(ref string) error {
	if strings.HasPrefix(ref, "doh://") {
		return nil
	}

	return errors.New("doesn't start with doh://, d'oh." + ref)
}

func (dfb *dockerfilebuild) Build(ctx context.Context, b string) (Result, error) {
	log.Println("Building", b)

	b = strings.TrimPrefix(b, "doh://")

	tmp, err := ioutil.TempDir("", "dohbuild")
	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(tmp)

	cmd := exec.Command("docker", "build", b, "-t", "doh.local/"+b, "--iidfile", filepath.Join(tmp, "iid"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("docker build: %w", err)
	}

	iid, err := ioutil.ReadFile(filepath.Join(tmp, "iid"))
	if err != nil {
		return nil, fmt.Errorf("open iid file: %w", err)
	}

	return daemon.Image(iidref(iid))
}

// iidref implements name.Ref for a local iid from docker build.
type iidref []byte

// Context accesses the Repository context of the reference.
func (d iidref) Context() name.Repository {
	return name.Repository{}
}

// String returns a string version of the iid
func (d iidref) String() string {
	return string(d)
}

// Identifier accesses the type-specific portion of the reference.
func (d iidref) Identifier() string {
	return string(d)
}

// Name is the fully-qualified reference name.
func (d iidref) Name() string {
	return string(d)
}

// Scope is the scope needed to access this reference.
func (d iidref) Scope(string) string {
	return ""
}
