package ssh

import (
	"github.com/cooper/screenmgr/device"
	"golang.org/x/crypto/ssh"
)

// convenient for methods involving device + ssh client
type sshClient struct {
	dev    *device.Device
	client *ssh.Client
}

// the list of devices with an SSH loop
var devices []*device.Device

// list of initializers for each OS family
// these are responsible for executing platform-specific commands
var initializers = make(map[string]func(sshClient) error)

func startDeviceLoop(dev *device.Device) error {
	devices = append(devices, dev)
	go deviceLoop(dev)
	return nil
}

func deviceLoop(dev *device.Device) {
	if !dev.Info.SSH.Enabled {
		return
	}

	// create ssh config
	config := &ssh.ClientConfig{
		User: dev.Info.SSH.Username,
		Auth: authMethods(dev),
	}

	// dial
	// TODO: support other ports
	client, err := ssh.Dial("tcp", dev.Info.AddrString+":22", config)
	if err != nil {
		dev.Warn("ssh dial failed: %v", err)
		return
	}

	dev.Log("SSH connection established")

	// call initializer for this OS family
	family := dev.Info.Software["OSFamily"]
	if handler, exists := initializers[family]; exists {
		dev.Debug("initializing via ssh for OS family: %s", family)
		handler(sshClient{dev, client})
	}

}

// returns preferred authentication methods
func authMethods(dev *device.Device) (methods []ssh.AuthMethod) {

	// TODO: keys
	if dev.Info.SSH.UsesKey {

	}

	// password authentication
	if pw := dev.Info.SSH.Password; pw != "" {
		methods = append(methods, ssh.Password(pw))
	}

	return methods
}

// returns combined stdout + stderr
func (s sshClient) combinedOutputBytes(command string) []byte {
	sess, err := s.client.NewSession()
	if err != nil {
		s.dev.Warn("could not create an SSH session")
		return nil
	}
	data, err := sess.CombinedOutput(command)
	if err != nil {
		s.dev.Warn("command `%s` failed: %s", command, err)
		return nil
	}
	return data
}

// returns combined stdout + stderr
func (s sshClient) combinedOutput(command string) string {
	return string(s.combinedOutputBytes(command))
}

// returns stdout
func (s sshClient) outputBytes(command string) []byte {
	sess, err := s.client.NewSession()
	if err != nil {
		s.dev.Warn("could not create an SSH session")
		return nil
	}
	data, err := sess.Output(command)
	if err != nil {
		s.dev.Warn("command `%s` failed: %s", command, err)
		return nil
	}
	return data
}

// returns stdout
func (s sshClient) output(command string) string {
	return string(s.outputBytes(command))
}

func init() {
	device.AddDeviceSetupCallback(startDeviceLoop)
}
