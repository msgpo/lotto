package mothership

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/mnordsletten/lotto/environment"
	"github.com/sirupsen/logrus"
)

// pushNacl appends lotto-test- to naclfilename and pushes it to mothership
func (m *Mothership) pushNacl(naclFileName string) (string, error) {
	logrus.Infof("Pushing NaCl: %s", naclFileName)
	fileName := path.Base(naclFileName)
	targetName := fmt.Sprintf("lotto-test-%s", strings.TrimSuffix(fileName, filepath.Ext(fileName)))
	request := fmt.Sprintf("push-nacl %s --name %s -o id", naclFileName, targetName)
	response, err := m.bin(request)
	if err != nil {
		return "", fmt.Errorf("could not push nacl %s: %v", naclFileName, err)
	}
	return response, nil
}

// pushUplink deletes any existing uplink with same name and pushes a new one
func (m *Mothership) pushUplink(file, name string) error {
	// Delete any existing uplink if it exists
	request := fmt.Sprintf("inspect-uplink %s", name)
	if _, err := m.bin(request); err == nil {
		request = fmt.Sprintf("delete-uplink %s", name)
		if _, err = m.bin(request); err != nil {
			return fmt.Errorf("couldn't remove uplink %s: %v", name, err)
		}
	}
	// Push the uplink
	request = fmt.Sprintf("push-uplink %s", file)
	if _, err := m.bin(request); err != nil {
		return fmt.Errorf("couldn't push uplink %s: %v", file, err)
	}
	return nil
}

// build will build with the specified naclID that needs to exist on mothership
func (m *Mothership) build(naclID string) (string, error) {
	logrus.Infof("Building image with nacl: %s", naclID)
	m.lastBuildTag = fmt.Sprintf("lotto-%s", time.Now().Format("20060102150405"))
	request := fmt.Sprintf("build --waitAndPrint -n %s -u %s --tag %s Starbase", naclID, m.uplinkname, m.lastBuildTag)
	checksum, err := m.bin(request)
	if err != nil {
		return "", fmt.Errorf("error building: %v", err)
	}

	return checksum, nil
}

// pullImage saves an image with targetName
func (m *Mothership) pullImage(checksum, targetName string) error {
	logrus.Debugf("Pulling down image: %s", checksum)
	_, err := m.bin(fmt.Sprintf("pull-image %s %s", checksum, targetName))
	if err != nil {
		return fmt.Errorf("error pulling image: %v", err)
	}
	return nil
}

// deploy takes an image checksum and deploys to starbase
func (m *Mothership) deploy(checksum string) error {
	logrus.Infof("Deploying %s to %s", checksum, m.alias)
	request := fmt.Sprintf("deploy %s %s", m.alias, checksum)
	if _, err := m.bin(request); err != nil {
		return fmt.Errorf("error deploying: %v", err)
	}
	return nil
}

// setAlias takes an ID and gives it the supplied alias
func (m *Mothership) setAlias(alias, ID string) error {
	request := fmt.Sprintf("instance-alias %s %s", ID, alias)
	if _, err := m.bin(request); err != nil {
		return err
	}
	return nil
}

// deleteInstanceByAlias deletes the given instance from mothership
func (m *Mothership) deleteInstanceByAlias(alias string) error {
	request := fmt.Sprintf("delete-instance %s", alias)
	if _, err := m.bin(request); err != nil {
		return err
	}
	return nil
}

// Launch uses requires the environment to specify the launch options
// to launch an instance where it is required.
func (m *Mothership) Launch(imageName string, env environment.Environment) error {
	options := env.LaunchCmdOptions(imageName)
	cmd := exec.Command(m.Binary, options...)
	logrus.Debugf("Launch command: %v", cmd.Args)
	cmd.Env = append(os.Environ())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error running launch cmd: %s, %v", string(output), err)
	}
	return nil
}
