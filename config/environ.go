package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"

	"github.com/hashicorp/errwrap"
)

// Environ generates a folder for consumption by envdir,
// which is used by wal-e to look up secrets without
// exposing them into PostgreSQL itself.
type Environ map[string]string

// NewPatroniEnvironFromClusterSpec setups up environment variables for patroni + scripts
func NewPatroniEnvironFromClusterSpec(clusterSpec *ClusterSpecification) *Environ {
	environ := Environ{}
	environ["REPLICATION_USER"] = clusterSpec.Postgresql.Appuser.Username
	environ["PATRONI_SCOPE"] = clusterSpec.Cluster.Scope
	environ["PG_DATA_DIR"] = "/data/postgres0"

	environ["ETCD_URI"] = clusterSpec.Etcd.URI

	environ["ARCHIVE_METHOD"] = clusterSpec.Archives.Method
	if clusterSpec.UsingWale() {
		environ["AWS_ACCESS_KEY_ID"] = clusterSpec.Archives.WalE.AWSAccessKeyID
		environ["AWS_SECRET_ACCESS_KEY"] = clusterSpec.Archives.WalE.AWSSecretAccessID
		environ["WAL_S3_BUCKET"] = clusterSpec.Archives.WalE.S3Bucket
		environ["WALE_S3_PREFIX"] = clusterSpec.waleS3Prefix()
		environ["WALE_S3_ENDPOINT"] = clusterSpec.Archives.WalE.S3Endpoint
	}
	if clusterSpec.UsingRsync() {
		environ["RSYNC_URI"] = clusterSpec.Archives.Rsync.URI
	}

	return &environ
}

// AddEnv adds an addition KEY=VALUE pair
func (environ *Environ) AddEnv(envvar string) {
	if !strings.Contains(envvar, "=") {
		fmt.Fprintf(os.Stderr, "Format error for env var '%s', must be 'KEY=VALUE'", envvar)
		return
	}
	parts := strings.Split(envvar, "=")
	key := parts[0]
	value := parts[1]
	if len(key) == 0 {
		fmt.Fprintf(os.Stderr, "Missing env variable name in '%s'\n", envvar)
		return
	}
	if len(value) == 0 {
		fmt.Fprintf(os.Stderr, "Missing env variable value in '%s'\n", envvar)
		return
	}
	(*environ)[key] = value
}

// CreateEnvDirFiles creates a directory with one file per env var
func (environ *Environ) CreateEnvDirFiles(dir string) (err error) {
	err = os.RemoveAll(dir)
	if err != nil {
		return errwrap.Wrapf("Cannot delete directory: {{err}}", err)
	}

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return
	}
	for name, value := range *environ {
		data := []byte(value)
		err = ioutil.WriteFile(path.Join(dir, name), data, 0644)
		if err != nil {
			return
		}
	}
	return
}

// CreateEnvScript creates a script that exports env vars
func (environ *Environ) CreateEnvScript(filePath string, chownUser string) (err error) {
	err = os.MkdirAll(path.Dir(filePath), 0755)
	if err != nil {
		return errwrap.Wrapf("Cannot mkdir: {{err}}", err)
	}

	var f *os.File
	f, err = os.Create(filePath)
	if err != nil {
		return errwrap.Wrapf("Cannot create file: {{err}}", err)
	}

	for name, value := range *environ {
		env := fmt.Sprintf("export %s=%s\n", name, value)
		_, err = f.WriteString(env)
		if err != nil {
			return errwrap.Wrapf("Cannot create write string to file: {{err}}", err)
		}
	}
	f.Sync()

	if chownUser != "" {
		u, err := user.Lookup(chownUser)
		if err != nil {
			return errwrap.Wrapf("Cannot lookup user: {{err}}", err)
		}
		uid, err := strconv.Atoi(u.Uid)
		if err != nil {
			return errwrap.Wrapf("Cannot get user Uid: {{err}}", err)
		}
		gid, err := strconv.Atoi(u.Gid)
		if err != nil {
			return errwrap.Wrapf("Cannot get user group Gid: {{err}}", err)
		}
		err = os.Chown(filePath, uid, gid)
		if err != nil {
			return errwrap.Wrapf("Cannot chown file: {{err}}", err)
		}
	}

	return
}
