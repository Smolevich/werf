package image

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"

	"github.com/flant/werf/pkg/dappdeps"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/util"
)

type StageImageContainer struct {
	image                      *StageImage
	name                       string
	runCommands                []string
	serviceRunCommands         []string
	runOptions                 *StageImageContainerOptions
	commitChangeOptions        *StageImageContainerOptions
	serviceCommitChangeOptions *StageImageContainerOptions
}

func newStageImageContainer(image *StageImage) *StageImageContainer {
	c := &StageImageContainer{}
	c.image = image
	c.name = fmt.Sprintf("%s%v", StageContainerNamePrefix, util.GenerateConsistentRandomString(10))
	c.runOptions = newStageContainerOptions()
	c.commitChangeOptions = newStageContainerOptions()
	c.serviceCommitChangeOptions = newStageContainerOptions()
	return c
}

func (c *StageImageContainer) Name() string {
	return c.name
}

func (c *StageImageContainer) CommitChanges() []string {
	commitChanges, err := c.prepareCommitChanges()
	if err != nil {
		panic(err)
	}

	return commitChanges
}

func (c *StageImageContainer) AllRunCommands() []string {
	return c.prepareAllRunCommands()
}

func (c *StageImageContainer) AddRunCommands(commands ...string) {
	c.runCommands = append(c.runCommands, commands...)
}

func (c *StageImageContainer) AddServiceRunCommands(commands ...string) {
	c.serviceRunCommands = append(c.serviceRunCommands, commands...)
}

func (c *StageImageContainer) RunOptions() ContainerOptions {
	return c.runOptions
}

func (c *StageImageContainer) CommitChangeOptions() ContainerOptions {
	return c.commitChangeOptions
}

func (c *StageImageContainer) ServiceCommitChangeOptions() ContainerOptions {
	return c.serviceCommitChangeOptions
}

func (c *StageImageContainer) prepareRunArgs() ([]string, error) {
	var args []string
	args = append(args, fmt.Sprintf("--name=%s", c.name))

	runOptions, err := c.prepareRunOptions()
	if err != nil {
		return nil, err
	}

	runArgs, err := runOptions.toRunArgs()
	if err != nil {
		return nil, err
	}

	fromImageId, err := c.image.fromImage.MustGetId()
	if err != nil {
		return nil, err
	}

	args = append(args, runArgs...)
	args = append(args, fromImageId)
	args = append(args, "-ec")
	args = append(args, c.prepareRunCommand())

	return args, nil
}

func (c *StageImageContainer) prepareRunCommand() string {
	return ShelloutPack(strings.Join(c.prepareRunCommands(), " && "))
}

func (c *StageImageContainer) prepareRunCommands() []string {
	runCommands := c.prepareAllRunCommands()
	if len(runCommands) != 0 {
		return runCommands
	} else {
		return []string{dappdeps.BaseBinPath("true")}
	}
}

func (c *StageImageContainer) prepareAllRunCommands() []string {
	return append(c.runCommands, c.serviceRunCommands...)
}

func ShelloutPack(command string) string {
	return fmt.Sprintf("eval $(echo %s | %s --decode)", base64.StdEncoding.EncodeToString([]byte(command)), dappdeps.BaseBinPath("base64"))
}

func (c *StageImageContainer) prepareIntrospectBeforeArgs() ([]string, error) {
	args, err := c.prepareIntrospectArgsBase()
	if err != nil {
		return nil, err
	}

	fromImageId, err := c.image.fromImage.MustGetId()
	if err != nil {
		return nil, err
	}

	args = append(args, fromImageId)
	args = append(args, "-ec")
	args = append(args, dappdeps.BaseBinPath("bash"))

	return args, nil
}

func (c *StageImageContainer) prepareIntrospectArgs() ([]string, error) {
	args, err := c.prepareIntrospectArgsBase()
	if err != nil {
		return nil, err
	}

	imageId, err := c.image.MustGetId()
	if err != nil {
		return nil, err
	}

	args = append(args, imageId)
	args = append(args, "-ec")
	args = append(args, dappdeps.BaseBinPath("bash"))

	return args, nil
}

func (c *StageImageContainer) prepareIntrospectArgsBase() ([]string, error) {
	var args []string

	runOptions, err := c.prepareIntrospectOptions()
	if err != nil {
		return nil, err
	}

	runArgs, err := runOptions.toRunArgs()
	if err != nil {
		return nil, err
	}

	args = append(args, []string{"-ti", "--rm"}...)
	args = append(args, runArgs...)

	return args, nil
}

func (c *StageImageContainer) prepareRunOptions() (*StageImageContainerOptions, error) {
	serviceRunOptions, err := c.prepareServiceRunOptions()
	if err != nil {
		return nil, err
	}
	return serviceRunOptions.merge(c.runOptions), nil
}

func (c *StageImageContainer) prepareServiceRunOptions() (*StageImageContainerOptions, error) {
	serviceRunOptions := newStageContainerOptions()
	serviceRunOptions.Workdir = "/"
	serviceRunOptions.Entrypoint = []string{dappdeps.BaseBinPath("bash")}
	serviceRunOptions.User = "0:0"

	baseContainerName, err := dappdeps.BaseContainer()
	if err != nil {
		return nil, err
	}

	toolchainContainerName, err := dappdeps.ToolchainContainer()
	if err != nil {
		return nil, err
	}
	serviceRunOptions.VolumesFrom = []string{baseContainerName, toolchainContainerName}

	return serviceRunOptions, nil
}

func (c *StageImageContainer) prepareIntrospectOptions() (*StageImageContainerOptions, error) {
	return c.prepareRunOptions()
}

func (c *StageImageContainer) prepareCommitChanges() ([]string, error) {
	commitOptions, err := c.prepareCommitOptions()
	if err != nil {
		return nil, err
	}

	commitChanges, err := commitOptions.toCommitChanges()
	if err != nil {
		return nil, err
	}
	return commitChanges, nil
}

func (c *StageImageContainer) prepareCommitOptions() (*StageImageContainerOptions, error) {
	inheritedCommitOptions, err := c.prepareInheritedCommitOptions()
	if err != nil {
		return nil, err
	}

	commitOptions := inheritedCommitOptions.merge(c.serviceCommitChangeOptions.merge(c.commitChangeOptions))
	return commitOptions, nil
}

func (c *StageImageContainer) prepareInheritedCommitOptions() (*StageImageContainerOptions, error) {
	inheritedOptions := newStageContainerOptions()

	if c.image.fromImage == nil {
		panic(fmt.Sprintf("runtime error: FromImage should be (%s)", c.image.name))
	}

	fromImageInspect, err := c.image.fromImage.MustGetInspect()
	if err != nil {
		return nil, err
	}

	inheritedOptions.Entrypoint = fromImageInspect.Config.Entrypoint
	inheritedOptions.Cmd = fromImageInspect.Config.Cmd
	inheritedOptions.User = fromImageInspect.Config.User
	if fromImageInspect.Config.WorkingDir != "" {
		inheritedOptions.Workdir = fromImageInspect.Config.WorkingDir
	} else {
		inheritedOptions.Workdir = "/"
	}
	return inheritedOptions, nil
}

func (c *StageImageContainer) run() error {
	runArgs, err := c.prepareRunArgs()
	if err != nil {
		return err
	}

	if err := docker.CliRun(runArgs...); err != nil {
		return fmt.Errorf("container run failed: %s", err.Error())
	}

	return nil
}

func (c *StageImageContainer) introspect() error {
	runArgs, err := c.prepareIntrospectArgs()
	if err != nil {
		return err
	}

	if err := docker.CliRun(runArgs...); err != nil {
		return err
	}

	return nil
}

func (c *StageImageContainer) introspectBefore() error {
	runArgs, err := c.prepareIntrospectBeforeArgs()
	if err != nil {
		return err
	}

	if err := docker.CliRun(runArgs...); err != nil {
		return err
	}

	return nil
}

func (c *StageImageContainer) commit() (string, error) {
	commitChanges, err := c.prepareCommitChanges()
	if err != nil {
		return "", err
	}

	commitOptions := types.ContainerCommitOptions{Changes: commitChanges}
	id, err := docker.ContainerCommit(c.name, commitOptions)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (c *StageImageContainer) rm() error {
	return docker.ContainerRemove(c.name, types.ContainerRemoveOptions{})
}
